package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"stitchforce-server/internal/entity"
	"stitchforce-server/internal/repo"
	"stitchforce-server/internal/usecase/output"
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) repo.ArchiveRepository {
	return &Repository{db: db}
}

// LIST ARCHIVE-ELIGIBLE ORGANIZATIONS

// ListEligibleOrganizations returns only organizations whose licenses are
// currently active and have not expired. Filtering is done by PostgreSQL so
// the report use case never loads ineligible organizations.
func (r *Repository) ListEligibleOrganizations(
	ctx context.Context,
) ([]*output.ArchiveOrganization, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			id,
			created_at
		FROM organizations
		WHERE license_status = 'active'
		  AND license_expires_at IS NOT NULL
		  AND license_expires_at > CURRENT_TIMESTAMP
		ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("query archive-eligible organizations: %w", err)
	}
	defer rows.Close()

	organizations := make([]*output.ArchiveOrganization, 0)

	for rows.Next() {
		organization := new(output.ArchiveOrganization)
		if err := rows.Scan(
			&organization.ID,
			&organization.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan archive-eligible organization: %w", err)
		}

		organizations = append(organizations, organization)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate archive-eligible organizations: %w", err)
	}

	return organizations, nil
}

// HAS ARCHIVED DATA

// LIST ARCHIVED PERIODS

func (r *Repository) ListArchivedPeriods(
    ctx context.Context,
    organizationID string,
    from, to time.Time,
) ([]output.ArchivePeriod, error) {
    rows, err := r.db.QueryContext(ctx, `
        WITH archive_periods AS (
            SELECT date_trunc(
                'month',
                production_date AT TIME ZONE 'Asia/Kolkata'
            ) AS period_start
            FROM production_entries
            WHERE organization_id = $1
              AND production_date >= $2
              AND production_date < $3
            GROUP BY 1

            UNION

            SELECT date_trunc(
                'month',
                payment_date AT TIME ZONE 'Asia/Kolkata'
            ) AS period_start
            FROM worker_payments
            WHERE organization_id = $1
              AND payment_date >= $2
              AND payment_date < $3
            GROUP BY 1

            UNION

            SELECT date_trunc(
                'month',
                updated_at AT TIME ZONE 'Asia/Kolkata'
            ) AS period_start
            FROM bundles
            WHERE organization_id = $1
              AND status = 'completed'
              AND updated_at >= $2
              AND updated_at < $3
            GROUP BY 1
        )
        SELECT
            EXTRACT(YEAR FROM period_start)::INT AS year,
            EXTRACT(MONTH FROM period_start)::INT AS month
        FROM archive_periods
        ORDER BY period_start
    `, organizationID, from, to)
    if err != nil {
        return nil, fmt.Errorf("query archived periods: %w", err)
    }
    defer rows.Close()

    periods := make([]output.ArchivePeriod, 0)

    for rows.Next() {
        var period output.ArchivePeriod

        if err := rows.Scan(
            &period.Year,
            &period.Month,
        ); err != nil {
            return nil, fmt.Errorf("scan archived period: %w", err)
        }

        periods = append(periods, period)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("iterate archived periods: %w", err)
    }

    return periods, nil
}

// PRODUCTION ENTRIES

func (r *Repository) StreamProductionEntries(
	ctx context.Context,
	organizationID string,
	from, to time.Time,
	write func(*entity.ProductionEntry) error,
) (int, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			id,
			organization_id,
			bundle_id,
			worker_id,
			style_operation_id,
			quantity,
			rate_per_piece,
			production_date,
			remarks,
			created_at,
			updated_at,
			deleted_at
		FROM production_entries
		WHERE organization_id = $1
		  AND production_date >= $2
		  AND production_date < $3
	`, organizationID, from, to)
	if err != nil {
		return 0, fmt.Errorf("query production entries: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		entry := new(entity.ProductionEntry)
		if err := rows.Scan(
			&entry.ID,
			&entry.OrganizationID,
			&entry.BundleID,
			&entry.WorkerID,
			&entry.StyleOperationID,
			&entry.Quantity,
			&entry.RatePerPiece,
			&entry.ProductionDate,
			&entry.Remarks,
			&entry.CreatedAt,
			&entry.UpdatedAt,
			&entry.DeletedAt,
		); err != nil {
			return count, fmt.Errorf("scan production entry: %w", err)
		}

		if err := write(entry); err != nil {
			return count, fmt.Errorf("write production entry: %w", err)
		}
		count++
	}

	if err := rows.Err(); err != nil {
		return count, fmt.Errorf("iterate production entries: %w", err)
	}

	return count, nil
}

// WORKER PAYMENTS

func (r *Repository) StreamWorkerPayments(
	ctx context.Context,
	organizationID string,
	from, to time.Time,
	write func(*entity.WorkerPayment) error,
) (int, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
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
		FROM worker_payments
		WHERE organization_id = $1
		  AND payment_date >= $2
		  AND payment_date < $3
	`, organizationID, from, to)
	if err != nil {
		return 0, fmt.Errorf("query worker payments: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		payment := new(entity.WorkerPayment)
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
		); err != nil {
			return count, fmt.Errorf("scan worker payment: %w", err)
		}

		if err := write(payment); err != nil {
			return count, fmt.Errorf("write worker payment: %w", err)
		}
		count++
	}

	if err := rows.Err(); err != nil {
		return count, fmt.Errorf("iterate worker payments: %w", err)
	}

	return count, nil
}

// PAYMENT ALLOCATIONS

func (r *Repository) StreamPaymentAllocations(
	ctx context.Context,
	organizationID string,
	from, to time.Time,
	write func(*output.PaymentAllocation) error,
) (int, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			wpa.id,
			wpa.organization_id,
			wpa.worker_payment_id,
			wpa.production_entry_id,
			wpa.amount,
			wpa.created_at
		FROM worker_payment_allocations wpa
		WHERE wpa.organization_id = $1
		  AND (
			EXISTS (
				SELECT 1
				FROM worker_payments wp
				WHERE wp.id = wpa.worker_payment_id
				  AND wp.organization_id = wpa.organization_id
				  AND wp.payment_date >= $2
				  AND wp.payment_date < $3
			)
			OR EXISTS (
				SELECT 1
				FROM production_entries pe
				WHERE pe.id = wpa.production_entry_id
				  AND pe.organization_id = wpa.organization_id
				  AND pe.production_date >= $2
				  AND pe.production_date < $3
			)
		  )
		ORDER BY wpa.id
	`, organizationID, from, to)
	if err != nil {
		return 0, fmt.Errorf("query payment allocations: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		allocation := new(output.PaymentAllocation)
		if err := rows.Scan(
			&allocation.ID,
			&allocation.OrganizationID,
			&allocation.WorkerPaymentID,
			&allocation.ProductionEntryID,
			&allocation.Amount,
			&allocation.CreatedAt,
		); err != nil {
			return count, fmt.Errorf("scan payment allocation: %w", err)
		}

		if err := write(allocation); err != nil {
			return count, fmt.Errorf("write payment allocation: %w", err)
		}
		count++
	}

	if err := rows.Err(); err != nil {
		return count, fmt.Errorf("iterate payment allocations: %w", err)
	}

	return count, nil
}

// COMPLETED BUNDLES

func (r *Repository) StreamCompletedBundles(
	ctx context.Context,
	organizationID string,
	from, to time.Time,
	write func(*entity.Bundle) error,
) (int, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			id,
			organization_id,
			bundle_number,
			style_id,
			size_id,
			quantity,
			attributes,
			status,
			remarks,
			created_at,
			updated_at,
			deleted_at
		FROM bundles
		WHERE organization_id = $1
		  AND status = 'completed'
		  AND updated_at >= $2
		  AND updated_at < $3
	`, organizationID, from, to)
	if err != nil {
		return 0, fmt.Errorf("query completed bundles: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		bundle := new(entity.Bundle)
		if err := rows.Scan(
			&bundle.ID,
			&bundle.OrganizationID,
			&bundle.BundleNumber,
			&bundle.StyleID,
			&bundle.SizeID,
			&bundle.Quantity,
			&bundle.Attributes,
			&bundle.Status,
			&bundle.Remarks,
			&bundle.CreatedAt,
			&bundle.UpdatedAt,
			&bundle.DeletedAt,
		); err != nil {
			return count, fmt.Errorf("scan completed bundle: %w", err)
		}

		if err := write(bundle); err != nil {
			return count, fmt.Errorf("write completed bundle: %w", err)
		}
		count++
	}

	if err := rows.Err(); err != nil {
		return count, fmt.Errorf("iterate completed bundles: %w", err)
	}

	return count, nil
}

// DELETE ARCHIVED DATA

func (r *Repository) DeleteArchivedData(
	ctx context.Context,
	organizationID string,
	from, to time.Time,
) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return fmt.Errorf("begin archive deletion transaction: %w", err)
	}
	defer tx.Rollback()

	// Serialize archive cleanup for this organization.
	var lockedID string
	if err := tx.QueryRowContext(ctx, `
		SELECT id
		FROM organizations
		WHERE id = $1
		FOR UPDATE
	`, organizationID).Scan(&lockedID); err != nil {
		return fmt.Errorf("lock organization for archive deletion: %w", err)
	}

	// Delete allocations before their referenced payments and entries.
	_, err = tx.ExecContext(ctx, `
		DELETE FROM worker_payment_allocations wpa
		WHERE wpa.organization_id = $1
		  AND (
			EXISTS (
				SELECT 1
				FROM worker_payments wp
				WHERE wp.id = wpa.worker_payment_id
				  AND wp.organization_id = wpa.organization_id
				  AND wp.payment_date >= $2
				  AND wp.payment_date < $3
			)
			OR EXISTS (
				SELECT 1
				FROM production_entries pe
				WHERE pe.id = wpa.production_entry_id
				  AND pe.organization_id = wpa.organization_id
				  AND pe.production_date >= $2
				  AND pe.production_date < $3
			)
		  )
	`, organizationID, from, to)
	if err != nil {
		return fmt.Errorf("delete archived payment allocations: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		DELETE FROM worker_payments
		WHERE organization_id = $1
		  AND payment_date >= $2
		  AND payment_date < $3
	`, organizationID, from, to)
	if err != nil {
		return fmt.Errorf("delete archived worker payments: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		DELETE FROM production_entries
		WHERE organization_id = $1
		  AND production_date >= $2
		  AND production_date < $3
	`, organizationID, from, to)
	if err != nil {
		return fmt.Errorf("delete archived production entries: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		DELETE FROM bundles b
		WHERE b.organization_id = $1
		  AND b.status = 'completed'
		  AND b.updated_at >= $2
		  AND b.updated_at < $3
		  AND NOT EXISTS (
			SELECT 1
			FROM production_entries pe
			WHERE pe.organization_id = b.organization_id
			  AND pe.bundle_id = b.id
		  )
	`, organizationID, from, to)
	if err != nil {
		return fmt.Errorf("delete archived completed bundles: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit archive deletion: %w", err)
	}

	return nil
}

var _ repo.ArchiveRepository = (*Repository)(nil)
