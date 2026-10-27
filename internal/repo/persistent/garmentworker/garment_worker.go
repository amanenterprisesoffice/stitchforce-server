
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

func (r *Repository) Create(
	ctx context.Context,
	organizationID string,
	worker *entity.GarmentWorker,
) error {
	query := `
		INSERT INTO garment_workers (
			organization_id, name, phone, gender
		)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	return r.db.QueryRowContext(
		ctx,
		query,
		organizationID,
		worker.Name,
		worker.Phone,
		worker.Gender,
	).Scan(&worker.ID)
}

func (r *Repository) GetByID(
	ctx context.Context,
	organizationID string,
	id string,
) (*entity.GarmentWorker, error) {
	query := `
		SELECT id, organization_id, name, phone, gender
		FROM garment_workers
		WHERE id = $1 AND organization_id = $2
	`

	worker := &entity.GarmentWorker{}

	err := r.db.QueryRowContext(
		ctx, query, id, organizationID,
	).Scan(
		&worker.ID,
		&worker.OrganizationID,
		&worker.Name,
		&worker.Phone,
		&worker.Gender,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repo.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return worker, nil
}

func (r *Repository) List(
	ctx context.Context,
	organizationID string,
) ([]*entity.GarmentWorker, error) {
	query := `
		SELECT id, organization_id, name, phone, gender
		FROM garment_workers
		WHERE organization_id = $1
		ORDER BY id
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID)
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

func (r *Repository) Update(
	ctx context.Context,
	organizationID string,
	worker *entity.GarmentWorker,
) error {
	query := `
		UPDATE garment_workers
		SET name = $3,
		    phone = $4,
		    gender = $5
		WHERE id = $1 AND organization_id = $2
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		worker.ID,
		organizationID,
		worker.Name,
		worker.Phone,
		worker.Gender,
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

func (r *Repository) Delete(
	ctx context.Context,
	organizationID string,
	id string,
) error {
	result, err := r.db.ExecContext(
		ctx,
		`DELETE FROM garment_workers
		 WHERE id = $1 AND organization_id = $2`,
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
