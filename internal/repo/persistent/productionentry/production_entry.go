package productionentry

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"stitchforce-server/internal/entity"
	"stitchforce-server/internal/repo"
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

	// 1. LOCK AND FETCH BUNDLE

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

	// 2. VERIFY WORKER

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

	// 3. VERIFY STYLE OPERATION

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

	// Ensure the rate has not changed since the usecase fetched it.
	if operationRate != entry.RatePerPiece {
		return errors.New("style operation rate changed; please retry")
	}

	// 4. CALCULATE PRODUCED QUANTITY

	var producedQuantity int

	err = tx.QueryRowContext(
		ctx,
		`
		SELECT COALESCE(SUM(quantity), 0)
		FROM production_entries
		WHERE organization_id = $1
		  AND bundle_id = $2
		  AND deleted_at IS NULL
		`,
		organizationID,
		entry.BundleID,
	).Scan(&producedQuantity)

	if err != nil {
		return err
	}

	// 5. ENFORCE BUNDLE QUANTITY

	if producedQuantity+entry.Quantity > bundleQuantity {
		remaining := bundleQuantity - producedQuantity

		return fmt.Errorf(
			"production quantity exceeds bundle limit: remaining quantity is %d",
			remaining,
		)
	}

	// 6. INSERT PRODUCTION ENTRY

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

	// 7. COMMIT

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
) ([]*entity.ProductionEntry, error) {

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
		WHERE organization_id = $1
		  AND deleted_at IS NULL
		ORDER BY production_date DESC, created_at DESC
	`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		organizationID,
	)

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
