
package garmentworker

import (
	"context"
	"database/sql"
	"errors"

	"stitchforce-server/internal/entity"
	"stitchforce-server/internal/repo"
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) repo.GarmentWorkerRepository {
	return &Repository{db: db}
}

// =========================================
// CREATE
// =========================================

func (r *Repository) Create(
	ctx context.Context,
	organizationID string,
	worker *entity.GarmentWorker,
) error {

	query := `
		INSERT INTO garment_workers (
			id,
			organization_id,
			name,
			phone,
			gender,
			photo_key,
			status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING
			id,
			status,
			created_at,
			updated_at,
			deleted_at
	`

	return r.db.QueryRowContext(
		ctx,
		query,
		worker.ID,
		organizationID,
		worker.Name,
		worker.Phone,
		worker.Gender,
		worker.PhotoKey,
		worker.Status,
	).Scan(
		&worker.ID,
		&worker.Status,
		&worker.CreatedAt,
		&worker.UpdatedAt,
		&worker.DeletedAt,
	)
}

// =========================================
// GET BY ID
// =========================================

func (r *Repository) GetByID(
	ctx context.Context,
	organizationID string,
	id string,
) (*entity.GarmentWorker, error) {

	query := `
		SELECT
			id,
			organization_id,
			name,
			COALESCE(phone, ''),
			COALESCE(gender, ''),
			COALESCE(photo_key, ''),
			status,
			created_at,
			updated_at,
			deleted_at
		FROM garment_workers
		WHERE id = $1
		  AND organization_id = $2
		  AND deleted_at IS NULL
	`

	worker := &entity.GarmentWorker{}

	err := r.db.QueryRowContext(
		ctx,
		query,
		id,
		organizationID,
	).Scan(
		&worker.ID,
		&worker.OrganizationID,
		&worker.Name,
		&worker.Phone,
		&worker.Gender,
		&worker.PhotoKey,
		&worker.Status,
		&worker.CreatedAt,
		&worker.UpdatedAt,
		&worker.DeletedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, repo.ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return worker, nil
}

// =========================================
// LIST
// =========================================

func (r *Repository) List(
	ctx context.Context,
	organizationID string,
) ([]*entity.GarmentWorker, error) {

	query := `
		SELECT
			id,
			organization_id,
			name,
			COALESCE(phone, ''),
			COALESCE(gender, ''),
			COALESCE(photo_key, ''),
			status,
			created_at,
			updated_at,
			deleted_at
		FROM garment_workers
		WHERE organization_id = $1
		  AND deleted_at IS NULL
		ORDER BY created_at DESC
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

	workers := make([]*entity.GarmentWorker, 0)

	for rows.Next() {
		worker := &entity.GarmentWorker{}

		if err := rows.Scan(
			&worker.ID,
			&worker.OrganizationID,
			&worker.Name,
			&worker.Phone,
			&worker.Gender,
			&worker.PhotoKey,
			&worker.Status,
			&worker.CreatedAt,
			&worker.UpdatedAt,
			&worker.DeletedAt,
		); err != nil {
			return nil, err
		}

		workers = append(workers, worker)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return workers, nil
}

// =========================================
// UPDATE
// =========================================

func (r *Repository) Update(
	ctx context.Context,
	organizationID string,
	worker *entity.GarmentWorker,
) error {

	query := `
		UPDATE garment_workers
		SET
			name = COALESCE(NULLIF($3, ''), name),
			phone = COALESCE(NULLIF($4, ''), phone),
			gender = COALESCE(NULLIF($5, ''), gender),
			photo_key = COALESCE(NULLIF($6, ''), photo_key),
			status = COALESCE(NULLIF($7, ''), status),
			updated_at = NOW()
		WHERE id = $1
		  AND organization_id = $2
		  AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		worker.ID,
		organizationID,
		worker.Name,
		worker.Phone,
		worker.Gender,
		worker.PhotoKey,
		string(worker.Status),
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
// =========================================
// SOFT DELETE
// =========================================

func (r *Repository) Delete(
	ctx context.Context,
	organizationID string,
	id string,
) error {

	query := `
		UPDATE garment_workers
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
