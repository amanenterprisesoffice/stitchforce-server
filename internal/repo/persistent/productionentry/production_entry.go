package productionentry

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"stitchforce-server/internal/entity"
	"stitchforce-server/internal/repo"
	"stitchforce-server/internal/usecase/input"
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) repo.ProductionEntryRepository {
	return &Repository{db: db}
}

// =========================================
// CREATE WITH BUNDLE QUANTITY LIMIT
// =========================================

func (r *Repository) CreateWithBundleLimit(
    ctx context.Context,
    organizationID string,
    entry *entity.ProductionEntry,
) error {

    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // 1. LOCK AND FETCH BUNDLE.

    var (
        bundleQuantity int
        bundleStyleID  string
        bundleStatus   string
    )

    err = tx.QueryRowContext(
        ctx,
        `
        SELECT
            quantity,
            style_id,
            status
        FROM bundles
        WHERE id = $1
          AND organization_id = $2
          AND deleted_at IS NULL
        FOR UPDATE
        `,
        entry.BundleID,
        organizationID,
    ).Scan(
        &bundleQuantity,
        &bundleStyleID,
        &bundleStatus,
    )

    if errors.Is(err, sql.ErrNoRows) {
        return repo.ErrNotFound
    }

    if err != nil {
        return err
    }

    if bundleStatus != "ready" && bundleStatus != "in_progress" {
        return errors.New("bundle is not available for production")
    }

    if bundleQuantity <= 0 {
        return errors.New("bundle quantity must be greater than zero")
    }

    // Always enforce the actual bundle quantity.
    entry.Quantity = bundleQuantity

    // 2. VERIFY WORKER.

    var workerID string

    err = tx.QueryRowContext(
        ctx,
        `
        SELECT id
        FROM garment_workers
        WHERE id = $1
          AND organization_id = $2
          AND status = 'active'
          AND deleted_at IS NULL
        FOR SHARE
        `,
        entry.WorkerID,
        organizationID,
    ).Scan(&workerID)

    if errors.Is(err, sql.ErrNoRows) {
        return errors.New("garment worker not found or inactive")
    }

    if err != nil {
        return err
    }

    // 3. VERIFY STYLE OPERATION.

    var operationRate int64

    err = tx.QueryRowContext(
        ctx,
        `
        SELECT rate_per_piece
        FROM style_operations
        WHERE id = $1
          AND style_id = $2
          AND organization_id = $3
        FOR SHARE
        `,
        entry.StyleOperationID,
        bundleStyleID,
        organizationID,
    ).Scan(&operationRate)

    if errors.Is(err, sql.ErrNoRows) {
        return errors.New("style operation does not exist for this bundle's style")
    }

    if err != nil {
        return err
    }

    // Ensure rate has not changed since the usecase fetched it.
    if operationRate != entry.RatePerPiece {
        return errors.New("style operation rate changed; please retry")
    }

    // 4. CHECK DUPLICATE OPERATION ENTRY.

    var exists bool

    err = tx.QueryRowContext(
        ctx,
        `
        SELECT EXISTS (
            SELECT 1
            FROM production_entries
            WHERE organization_id = $1
              AND bundle_id = $2
              AND style_operation_id = $3
              AND deleted_at IS NULL
        )
        `,
        organizationID,
        entry.BundleID,
        entry.StyleOperationID,
    ).Scan(&exists)

    if err != nil {
        return err
    }

    if exists {
        return errors.New(
            "production entry already exists for this bundle operation",
        )
    }

    // 5. INSERT PRODUCTION ENTRY.

    query := `
        INSERT INTO production_entries (
            id,
            organization_id,
            bundle_id,
            worker_id,
            style_operation_id,
            quantity,
            rate_per_piece,
            production_date,
            remarks
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        RETURNING
            created_at,
            updated_at,
            deleted_at
    `

    err = tx.QueryRowContext(
        ctx,
        query,
        entry.ID,
        organizationID,
        entry.BundleID,
        entry.WorkerID,
        entry.StyleOperationID,
        entry.Quantity,
        entry.RatePerPiece,
        entry.ProductionDate,
        entry.Remarks,
    ).Scan(
        &entry.CreatedAt,
        &entry.UpdatedAt,
        &entry.DeletedAt,
    )

    if err != nil {
        return err
    }

    // 6. COMMIT.

    if err := tx.Commit(); err != nil {
        return err
    }

    return nil
}

// =========================================
// GET BY ID
// =========================================

func (r *Repository) GetByID(
	ctx context.Context,
	organizationID string,
	id string,
) (*entity.ProductionEntry, error) {

	query := `
		SELECT
			id,
			organization_id,
			bundle_id,
			worker_id,
			quantity,
			production_date,
			rate_per_piece,
			style_operation_id,
			remarks,
			created_at,
			updated_at,
			deleted_at
		FROM production_entries
		WHERE id = $1
		  AND organization_id = $2
		  AND deleted_at IS NULL
	`

	entry := &entity.ProductionEntry{}

	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(
		ctx,
		query,
		id,
		organizationID,
	).Scan(
		&entry.ID,
		&entry.OrganizationID,
		&entry.BundleID,
		&entry.WorkerID,
		&entry.Quantity,
		&entry.ProductionDate,
		&entry.RatePerPiece,
		&entry.StyleOperationID,
		&entry.Remarks,
		&entry.CreatedAt,
		&entry.UpdatedAt,
		&deletedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, repo.ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	if deletedAt.Valid {
		entry.DeletedAt = &deletedAt.Time
	}

	return entry, nil
}

// =========================================
// LIST
// =========================================

func (r *Repository) List(
	ctx context.Context,
	organizationID string,
	filter input.ProductionEntryFilter,
) ([]*entity.ProductionEntry, error) {

	query := `
		SELECT
			pe.id,
			pe.organization_id,
			pe.bundle_id,
			pe.worker_id,
			pe.quantity,
			pe.production_date,
			pe.rate_per_piece,
			pe.style_operation_id,
			pe.remarks,
			pe.created_at,
			pe.updated_at,
			pe.deleted_at
		FROM production_entries pe
		INNER JOIN bundles b
			ON b.id = pe.bundle_id
		   AND b.organization_id = pe.organization_id
		   AND b.deleted_at IS NULL
		INNER JOIN garment_workers gw
			ON gw.id = pe.worker_id
		   AND gw.organization_id = pe.organization_id
		INNER JOIN style_operations so
			ON so.id = pe.style_operation_id
		   AND so.organization_id = pe.organization_id
		WHERE pe.organization_id = $1
		  AND pe.deleted_at IS NULL
	`

	args := []any{organizationID}
	argPosition := 2

	if filter.Search != "" {
		query += fmt.Sprintf(`
			AND (
				b.bundle_number ILIKE $%d
				OR gw.name ILIKE $%d
				OR gw.phone ILIKE $%d
				OR so.name ILIKE $%d
			)
		`, argPosition, argPosition, argPosition, argPosition)

		args = append(args, "%"+filter.Search+"%")
		argPosition++
	}

	if filter.BundleID != "" {
		query += fmt.Sprintf(" AND pe.bundle_id = $%d", argPosition)
		args = append(args, filter.BundleID)
		argPosition++
	}

	if filter.WorkerID != "" {
		query += fmt.Sprintf(" AND pe.worker_id = $%d", argPosition)
		args = append(args, filter.WorkerID)
		argPosition++
	}

	if filter.StyleID != "" {
		query += fmt.Sprintf(" AND b.style_id = $%d", argPosition)
		args = append(args, filter.StyleID)
		argPosition++
	}

	if filter.StyleOperationID != "" {
		query += fmt.Sprintf(" AND pe.style_operation_id = $%d", argPosition)
		args = append(args, filter.StyleOperationID)
		argPosition++
	}

	if filter.FromDate != nil {
		query += fmt.Sprintf(" AND pe.production_date >= $%d", argPosition)
		args = append(args, *filter.FromDate)
		argPosition++
	}

	if filter.ToDate != nil {
		query += fmt.Sprintf(" AND pe.production_date < $%d", argPosition)
		args = append(args, *filter.ToDate)
		argPosition++
	}

	query += " ORDER BY pe.production_date DESC, pe.created_at DESC"

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query += fmt.Sprintf(
		" LIMIT $%d OFFSET $%d",
		argPosition,
		argPosition+1,
	)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]*entity.ProductionEntry, 0)

	for rows.Next() {
		entry := &entity.ProductionEntry{}
		var deletedAt sql.NullTime

		if err := rows.Scan(
			&entry.ID,
			&entry.OrganizationID,
			&entry.BundleID,
			&entry.WorkerID,
			&entry.Quantity,
			&entry.ProductionDate,
			&entry.RatePerPiece,
			&entry.StyleOperationID,
			&entry.Remarks,
			&entry.CreatedAt,
			&entry.UpdatedAt,
			&deletedAt,
		); err != nil {
			return nil, err
		}

		if deletedAt.Valid {
			entry.DeletedAt = &deletedAt.Time
		}

		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

// =========================================
// UPDATE WITH BUNDLE QUANTITY LIMIT
// =========================================

func (r *Repository) UpdateWithBundleLimit(
	ctx context.Context,
	organizationID string,
	entry *entity.ProductionEntry,
) error {

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Verify entry exists and lock it.
	var existingID string

	err = tx.QueryRowContext(
		ctx,
		`
		SELECT id
		FROM production_entries
		WHERE id = $1
		  AND organization_id = $2
		  AND deleted_at IS NULL
		FOR UPDATE
		`,
		entry.ID,
		organizationID,
	).Scan(&existingID)

	if errors.Is(err, sql.ErrNoRows) {
		return repo.ErrNotFound
	}

	if err != nil {
		return err
	}

	// Lock target bundle.
	var bundleQuantity int

	err = tx.QueryRowContext(
		ctx,
		`
		SELECT quantity
		FROM bundles
		WHERE id = $1
		  AND organization_id = $2
		  AND deleted_at IS NULL
		FOR UPDATE
		`,
		entry.BundleID,
		organizationID,
	).Scan(&bundleQuantity)

	if errors.Is(err, sql.ErrNoRows) {
		return repo.ErrNotFound
	}

	if err != nil {
		return err
	}

	// Verify worker belongs to organization and is active.
	var workerID string

	err = tx.QueryRowContext(
		ctx,
		`
		SELECT id
		FROM garment_workers
		WHERE id = $1
		  AND organization_id = $2
		  AND status = 'active'
		  AND deleted_at IS NULL
		FOR SHARE
		`,
		entry.WorkerID,
		organizationID,
	).Scan(&workerID)

	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("garment worker not found or inactive")
	}

	if err != nil {
		return err
	}

	// Calculate production quantity excluding current entry.
	var producedQuantity int

	err = tx.QueryRowContext(
		ctx,
		`
		SELECT COALESCE(SUM(quantity), 0)
		FROM production_entries
		WHERE organization_id = $1
		  AND bundle_id = $2
		  AND id != $3
		  AND deleted_at IS NULL
		`,
		organizationID,
		entry.BundleID,
		entry.ID,
	).Scan(&producedQuantity)

	if err != nil {
		return err
	}

	// Enforce bundle quantity.
	if producedQuantity+entry.Quantity > bundleQuantity {
		return fmt.Errorf(
			"production quantity exceeds bundle limit: remaining quantity is %d",
			bundleQuantity-producedQuantity,
		)
	}

	// Update entry.
	query := `
		UPDATE production_entries
		SET
			bundle_id = $3,
			worker_id = $4,
			quantity = $5,
			production_date = $6,
			remarks = $7,
			updated_at = NOW()
		WHERE id = $1
		  AND organization_id = $2
		  AND deleted_at IS NULL
		RETURNING
			created_at,
			updated_at,
			deleted_at
	`

	var deletedAt sql.NullTime

	err = tx.QueryRowContext(
		ctx,
		query,
		entry.ID,
		organizationID,
		entry.BundleID,
		entry.WorkerID,
		entry.Quantity,
		entry.ProductionDate,
		entry.Remarks,
	).Scan(
		&entry.CreatedAt,
		&entry.UpdatedAt,
		&deletedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return repo.ErrNotFound
	}

	if err != nil {
		return err
	}

	if deletedAt.Valid {
		entry.DeletedAt = &deletedAt.Time
	}

	return tx.Commit()
}

// =========================================
// SOFT DELETE
// =========================================

func (r *Repository) Delete(
	ctx context.Context,
	organizationID string,
	id string,
) error {

	query := `
		UPDATE production_entries
		SET
			deleted_at = NOW(),
			updated_at = NOW()
		WHERE id = $1
		  AND organization_id = $2
		  AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		id,
		organizationID,
	)

	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return repo.ErrNotFound
	}

	return nil
}
