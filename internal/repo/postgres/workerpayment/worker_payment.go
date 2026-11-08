package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"stitchforce-server/internal/entity"
	"stitchforce-server/internal/usecase/input"
	"stitchforce-server/internal/usecase/output"

	appUtils "stitchforce-server/pkg/utils"
)

type WorkerPaymentRepository struct {
	db *sql.DB
}

func New(db *sql.DB) *WorkerPaymentRepository {
	return &WorkerPaymentRepository{
		db: db,
	}
}

// CREATE WITH ALLOCATIONS

func (r *WorkerPaymentRepository) CreateWithAllocations(
	ctx context.Context,
	organizationID string,
	payment *entity.WorkerPayment,
	filter input.PayableEntriesFilter,
) error {
	if payment == nil {
		return errors.New("worker payment is required")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin payment transaction: %w", err)
	}
	defer tx.Rollback()

	toDateExclusive := filter.ToDate.AddDate(0, 0, 1)

	// 1. Lock eligible production entries.
	// This prevents concurrent payments from allocating the same
	// outstanding production entries.
	rows, err := tx.QueryContext(
		ctx,
		`
		SELECT id
		FROM production_entries
		WHERE organization_id = $1
			AND worker_id = $2
			AND deleted_at IS NULL
			AND production_date >= $3
			AND production_date < $4
		ORDER BY id
		FOR UPDATE
		`,
		organizationID,
		payment.WorkerID,
		filter.FromDate,
		toDateExclusive,
	)
	if err != nil {
		return fmt.Errorf("lock production entries: %w", err)
	}

	for rows.Next() {
		var id string

		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return fmt.Errorf("scan locked production entry: %w", err)
		}
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return fmt.Errorf("iterate locked production entries: %w", err)
	}

	if err := rows.Close(); err != nil {
		return fmt.Errorf("close production entry rows: %w", err)
	}

	// 2. Calculate pending amount for each production entry.
	// total_amount is a generated column in your schema.
	// Only completed payments count as paid.
	rows, err = tx.QueryContext(
		ctx,
		`
		SELECT
			pe.id,
			pe.total_amount - COALESCE(
				(
					SELECT SUM(wpa.amount)
					FROM worker_payment_allocations wpa
					INNER JOIN worker_payments wp
						ON wp.id = wpa.worker_payment_id
						AND wp.organization_id = wpa.organization_id
					WHERE wpa.organization_id = pe.organization_id
						AND wpa.production_entry_id = pe.id
						AND wp.status = $5
				),
				0
			) AS pending_amount
		FROM production_entries pe
		WHERE pe.organization_id = $1
			AND pe.worker_id = $2
			AND pe.deleted_at IS NULL
			AND pe.production_date >= $3
			AND pe.production_date < $4
		ORDER BY pe.id
		`,
		organizationID,
		payment.WorkerID,
		filter.FromDate,
		toDateExclusive,
		entity.PaymentStatusCompleted,
	)
	if err != nil {
		return fmt.Errorf("calculate pending production amounts: %w", err)
	}

	type allocation struct {
		productionEntryID string
		amount            int64
	}

	allocations := make([]allocation, 0)
	var totalAmount int64

	for rows.Next() {
		var item allocation

		if err := rows.Scan(
			&item.productionEntryID,
			&item.amount,
		); err != nil {
			rows.Close()
			return fmt.Errorf("scan pending production entry: %w", err)
		}

		if item.amount < 0 {
			rows.Close()
			return fmt.Errorf(
				"production entry %s has an invalid negative pending amount",
				item.productionEntryID,
			)
		}

		if item.amount == 0 {
			continue
		}

		allocations = append(allocations, item)
		totalAmount += item.amount
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return fmt.Errorf("iterate pending production entries: %w", err)
	}

	if err := rows.Close(); err != nil {
		return fmt.Errorf("close pending production entry rows: %w", err)
	}

	if totalAmount <= 0 {
		return errors.New(
			"worker has no pending amount for the selected date range",
		)
	}

	// 3. Set payment amount from the calculated allocations.
	payment.Amount = totalAmount

	// 4. Insert payment.
	_, err = tx.ExecContext(
		ctx,
		`
		INSERT INTO worker_payments (
			id,
			organization_id,
			worker_id,
			amount,
			payment_date,
			payment_method,
			reference,
			remarks,
			status,
			created_at,
			updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)
		`,
		payment.ID,
		organizationID,
		payment.WorkerID,
		payment.Amount,
		payment.PaymentDate,
		payment.PaymentMethod,
		payment.Reference,
		payment.Remarks,
		payment.Status,
		payment.CreatedAt,
		payment.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert worker payment: %w", err)
	}

	// 5. Insert allocations.
	for _, item := range allocations {
		allocationID, err := appUtils.NewID()
		if err != nil {
			return fmt.Errorf("generate allocation ID: %w", err)
		}

		_, err = tx.ExecContext(
			ctx,
			`
			INSERT INTO worker_payment_allocations (
				id,
				organization_id,
				worker_payment_id,
				production_entry_id,
				amount,
				created_at
			)
			VALUES ($1, $2, $3, $4, $5, $6)
			`,
			allocationID,
			organizationID,
			payment.ID,
			item.productionEntryID,
			item.amount,
			payment.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("insert payment allocation: %w", err)
		}
	}

	// 6. Commit payment and allocations together.
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit worker payment: %w", err)
	}

	return nil
}

// GET BY ID

func (r *WorkerPaymentRepository) GetByID(
	ctx context.Context,
	organizationID string,
	paymentID string,
) (*output.WorkerPayment, error) {
	var payment output.WorkerPayment

	// 1. Fetch payment and worker.
	paymentQuery := `
		SELECT
			wp.id,
			wp.organization_id,
			wp.worker_id,
			wp.amount,
			wp.payment_date,
			wp.payment_method,
			wp.reference,
			wp.remarks,
			wp.status,
			wp.created_at,
			wp.updated_at,

			gw.id,
			gw.name,
			COALESCE(gw.phone, ''),
			COALESCE(gw.photo_key, '')

		FROM worker_payments wp

		INNER JOIN garment_workers gw
			ON gw.id = wp.worker_id
			AND gw.organization_id = wp.organization_id

		WHERE wp.organization_id = $1
			AND wp.id = $2
	`

	var workerPhotoKey string

	err := r.db.QueryRowContext(
		ctx,
		paymentQuery,
		organizationID,
		paymentID,
	).Scan(
		&payment.ID,
		&payment.OrganizationID,
		&payment.WorkerID,
		&payment.Amount,
		&payment.PaymentDate,
		&payment.PaymentMethod,
		&payment.Reference,
		&payment.Remarks,
		&payment.Status,
		&payment.CreatedAt,
		&payment.UpdatedAt,

		&payment.Worker.ID,
		&payment.Worker.Name,
		&payment.Worker.Phone,
		&workerPhotoKey,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("get worker payment: %w", err)
	}

	// Resolve workerPhotoKey to a URL here if your repository
	// has access to your object-storage URL resolver.

	// 2. Fetch all production entries allocated to this payment.
	entriesQuery := `
		SELECT
			pe.id,
			pe.organization_id,
			pe.bundle_id,
			pe.worker_id,
			pe.style_operation_id,
			pe.quantity,
			pe.rate_per_piece,
			pe.total_amount,
			pe.production_date,
			pe.remarks,
			pe.created_at,
			pe.updated_at,

			b.id,
			b.bundle_number,
			b.quantity,
			b.status,

			s.id,
			s.name,
			s.style_number,

			so.id,
			so.name,

			pe.rate_per_piece,

			wpa.amount AS allocated_amount

		FROM worker_payment_allocations wpa

		INNER JOIN production_entries pe
			ON pe.id = wpa.production_entry_id
			AND pe.organization_id = wpa.organization_id

		INNER JOIN bundles b
			ON b.id = pe.bundle_id
			AND b.organization_id = pe.organization_id

		INNER JOIN styles s
			ON s.id = b.style_id
			AND s.organization_id = b.organization_id

		INNER JOIN style_operations so
			ON so.id = pe.style_operation_id
			AND so.organization_id = pe.organization_id

		WHERE wpa.organization_id = $1
			AND wpa.worker_payment_id = $2

		ORDER BY pe.production_date, pe.id
	`

	rows, err := r.db.QueryContext(
		ctx,
		entriesQuery,
		organizationID,
		paymentID,
	)
	if err != nil {
		return nil, fmt.Errorf("get worker payment production entries: %w", err)
	}
	defer rows.Close()

	payment.ProductionEntry = make([]output.PaidProductionEntry, 0)

	for rows.Next() {
		var entry output.PaidProductionEntry

		if err := rows.Scan(
			&entry.ID,
			&entry.OrganizationID,
			&entry.BundleID,
			&entry.WorkerID,
			&entry.StyleOperationID,
			&entry.Quantity,
			&entry.RatePerPiece,
			&entry.TotalAmount,
			&entry.ProductionDate,
			&entry.Remarks,
			&entry.CreatedAt,
			&entry.UpdatedAt,

			&entry.Bundle.ID,
			&entry.Bundle.BundleNumber,
			&entry.Bundle.Quantity,
			&entry.Bundle.Status,

			&entry.Style.ID,
			&entry.Style.Name,
			&entry.Style.StyleNumber,

			&entry.Operation.ID,
			&entry.Operation.Name,
			&entry.Operation.RatePerPiece,

			&entry.AllocatedAmount,
		); err != nil {
			return nil, fmt.Errorf("scan worker payment production entry: %w", err)
		}

		payment.ProductionEntry = append(
			payment.ProductionEntry,
			entry,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate worker payment production entries: %w", err)
	}

	return &payment, nil
}

// LIST
func (r *WorkerPaymentRepository) List(
	ctx context.Context,
	organizationID string,
	filter input.WorkerPaymentFilter,
) ([]*output.WorkerPayment, error) {

	query := `
		SELECT
			wp.id,
			wp.organization_id,
			wp.worker_id,
			wp.amount,
			wp.payment_date,
			wp.payment_method,
			wp.reference,
			wp.remarks,
			wp.status,
			wp.created_at,
			wp.updated_at,

			gw.id,
			gw.name,
			COALESCE(gw.phone, ''),
			COALESCE(gw.photo_key, '')

		FROM worker_payments wp

		INNER JOIN garment_workers gw
			ON gw.id = wp.worker_id
			AND gw.organization_id = wp.organization_id

		WHERE wp.organization_id = $1
	`

	args := []any{organizationID}
	argPosition := 2

	if filter.WorkerID != "" {
		query += fmt.Sprintf(
			" AND wp.worker_id = $%d",
			argPosition,
		)
		args = append(args, filter.WorkerID)
		argPosition++
	}

	if filter.Status != "" {
		query += fmt.Sprintf(
			" AND wp.status = $%d",
			argPosition,
		)
		args = append(args, filter.Status)
		argPosition++
	}

	if filter.FromDate != nil {
		query += fmt.Sprintf(
			" AND wp.payment_date >= $%d",
			argPosition,
		)
		args = append(args, *filter.FromDate)
		argPosition++
	}

	if filter.ToDate != nil {
		query += fmt.Sprintf(
			" AND wp.payment_date < $%d",
			argPosition,
		)
		args = append(args, filter.ToDate.AddDate(0, 0, 1))
		argPosition++
	}

	limit := filter.Limit
	offset := filter.Offset

	query += fmt.Sprintf(`
		ORDER BY wp.payment_date DESC, wp.created_at DESC
		LIMIT $%d OFFSET $%d
	`, argPosition, argPosition+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list worker payments: %w", err)
	}
	defer rows.Close()

	payments := make([]*output.WorkerPayment, 0)

	for rows.Next() {
		payment := &output.WorkerPayment{}

		if err := rows.Scan(
			&payment.ID,
			&payment.OrganizationID,
			&payment.WorkerID,
			&payment.Amount,
			&payment.PaymentDate,
			&payment.PaymentMethod,
			&payment.Reference,
			&payment.Remarks,
			&payment.Status,
			&payment.CreatedAt,
			&payment.UpdatedAt,

			&payment.Worker.ID,
			&payment.Worker.Name,
			&payment.Worker.Phone,
			&payment.Worker.PhotoKey,
		); err != nil {
			return nil, fmt.Errorf("scan worker payment: %w", err)
		}

		payments = append(payments, payment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate worker payments: %w", err)
	}

	return payments, nil
}

// GET WORKER PAYMENT SUMMARY
func (r *WorkerPaymentRepository) GetSummary(
	ctx context.Context,
	organizationID string,
	filter input.WorkerPaymentSummaryFilter,
) (*output.WorkerPaymentSummary, error) {

	args := []interface{}{
		organizationID,
		entity.PaymentStatusCompleted,
	}

	earnedQuery := `
        SELECT COALESCE(SUM(pe.total_amount), 0)
        FROM production_entries pe
        WHERE pe.organization_id = $1
            AND pe.deleted_at IS NULL
    `

	paidQuery := `
        SELECT COALESCE(SUM(wpa.amount), 0)
        FROM worker_payment_allocations wpa
        INNER JOIN worker_payments wp
            ON wp.id = wpa.worker_payment_id
            AND wp.organization_id = wpa.organization_id
        INNER JOIN production_entries pe
            ON pe.id = wpa.production_entry_id
            AND pe.organization_id = wpa.organization_id
        WHERE wpa.organization_id = $1
            AND wp.status = $2
            AND pe.deleted_at IS NULL
    `

	if filter.WorkerID != "" {
		args = append(args, filter.WorkerID)

		earnedQuery += fmt.Sprintf(
			" AND pe.worker_id = $%d",
			len(args),
		)

		paidQuery += fmt.Sprintf(
			" AND pe.worker_id = $%d",
			len(args),
		)
	}

	if filter.FromDate != nil {
		args = append(args, *filter.FromDate)

		earnedQuery += fmt.Sprintf(
			" AND pe.production_date >= $%d",
			len(args),
		)

		paidQuery += fmt.Sprintf(
			" AND pe.production_date >= $%d",
			len(args),
		)
	}

	if filter.ToDate != nil {
		args = append(args, filter.ToDate.AddDate(0, 0, 1))

		earnedQuery += fmt.Sprintf(
			" AND pe.production_date < $%d",
			len(args),
		)

		paidQuery += fmt.Sprintf(
			" AND pe.production_date < $%d",
			len(args),
		)
	}

	query := fmt.Sprintf(`
        SELECT
            (%s) AS total_earned,
            (%s) AS total_paid
    `, earnedQuery, paidQuery)

	var summary output.WorkerPaymentSummary

	err := r.db.QueryRowContext(
		ctx,
		query,
		args...,
	).Scan(
		&summary.TotalEarned,
		&summary.TotalPaid,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"get worker payment summary: %w",
			err,
		)
	}

	summary.TotalPending =
		summary.TotalEarned - summary.TotalPaid

	return &summary, nil
}

// LIST PAYABLE PRODUCTION ENTRIES

func (r *WorkerPaymentRepository) ListPayableEntries(
	ctx context.Context,
	organizationID string,
	workerID string,
	filter input.PayableEntriesFilter,
) ([]*output.PayableProductionEntry, error) {
	toDateExclusive := filter.ToDate.AddDate(0, 0, 1)

	query := `
		WITH entry_balances AS (
			SELECT
				pe.id AS production_entry_id,
				pe.worker_id,
				pe.production_date,
				pe.quantity,
				pe.rate_per_piece,
				pe.total_amount,

				b.id AS bundle_id,
				b.bundle_number,
				b.quantity AS bundle_quantity,
				b.status AS bundle_status,

				s.id AS style_id,
				s.name AS style_name,
				s.style_number,

				so.id AS operation_id,
				so.name AS operation_name,

				COALESCE(
					(
						SELECT SUM(wpa.amount)
						FROM worker_payment_allocations wpa
						INNER JOIN worker_payments wp
							ON wp.id = wpa.worker_payment_id
							AND wp.organization_id = wpa.organization_id
						WHERE wpa.organization_id = pe.organization_id
							AND wpa.production_entry_id = pe.id
							AND wp.status = $5
					),
					0
				) AS paid_amount

			FROM production_entries pe

			INNER JOIN bundles b
				ON b.id = pe.bundle_id
				AND b.organization_id = pe.organization_id

			INNER JOIN styles s
				ON s.id = b.style_id
				AND s.organization_id = b.organization_id

			INNER JOIN style_operations so
				ON so.id = pe.style_operation_id
				AND so.organization_id = pe.organization_id

			WHERE pe.organization_id = $1
				AND pe.worker_id = $2
				AND pe.deleted_at IS NULL
				AND b.deleted_at IS NULL
				AND pe.production_date >= $3
				AND pe.production_date < $4
		)

		SELECT
			production_entry_id,
			worker_id,

			bundle_id,
			bundle_number,
			bundle_quantity,
			bundle_status,

			style_id,
			style_name,
			style_number,

			operation_id,
			operation_name,

			production_date,
			quantity,
			rate_per_piece,
			total_amount,
			paid_amount,
			total_amount - paid_amount AS pending_amount

		FROM entry_balances
		WHERE total_amount - paid_amount > 0
		ORDER BY production_date, production_entry_id
	`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		organizationID,
		workerID,
		filter.FromDate,
		toDateExclusive,
		entity.PaymentStatusCompleted,
	)
	if err != nil {
		return nil, fmt.Errorf("list payable production entries: %w", err)
	}
	defer rows.Close()

	entries := make([]*output.PayableProductionEntry, 0)

	for rows.Next() {
		var entry output.PayableProductionEntry

		if err := rows.Scan(
			&entry.ProductionEntryID,
			&entry.WorkerId,

			&entry.Bundle.ID,
			&entry.Bundle.BundleNumber,
			&entry.Bundle.Quantity,
			&entry.Bundle.Status,

			&entry.Style.ID,
			&entry.Style.Name,
			&entry.Style.StyleNumber,

			&entry.Operation.ID,
			&entry.Operation.Name,

			&entry.ProductionDate,
			&entry.Quantity,
			&entry.RatePerPiece,
			&entry.TotalAmount,
			&entry.PaidAmount,
			&entry.PendingAmount,
		); err != nil {
			return nil, fmt.Errorf("scan payable production entry: %w", err)
		}

		// The operation rate is the snapshot stored on the production entry.
		entry.Operation.RatePerPiece = entry.RatePerPiece

		entries = append(entries, &entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate payable production entries: %w", err)
	}

	return entries, nil
}

// VOID PAYMENT

func (r *WorkerPaymentRepository) Void(
	ctx context.Context,
	organizationID string,
	paymentID string,
) error {
	result, err := r.db.ExecContext(
		ctx,
		`
		UPDATE worker_payments
		SET
			status = $1,
			updated_at = $2
		WHERE organization_id = $3
			AND id = $4
			AND status = $5
		`,
		entity.PaymentStatusVoided,
		time.Now().UTC(),
		organizationID,
		paymentID,
		entity.PaymentStatusCompleted,
	)
	if err != nil {
		return fmt.Errorf("void worker payment: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected payment rows: %w", err)
	}

	if affected == 0 {
		return errors.New("worker payment not found or already voided")
	}

	return nil
}
