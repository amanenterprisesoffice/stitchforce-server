package size

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

func New(db *sql.DB) repo.SizeRepository {
	return &Repository{db: db}
}

// CREATE
func (r *Repository) Create(
	ctx context.Context,
	organizationID string,
	size *entity.Size,
) error {
	query := `
        INSERT INTO sizes (
            id,
            organization_id,
            name
        )
        VALUES ($1, $2, $3)
        RETURNING created_at, updated_at
    `

	return r.db.QueryRowContext(
		ctx,
		query,
		size.ID,
		organizationID,
		size.Name,
	).Scan(
		&size.CreatedAt,
		&size.UpdatedAt,
	)
}

// SCAN

func scanSize(
	row interface {
		Scan(...any) error
	},
) (*entity.Size, error) {
	size := &entity.Size{}

	err := row.Scan(
		&size.ID,
		&size.OrganizationID,
		&size.Name,
		&size.CreatedAt,
		&size.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return size, nil
}

// SELECT COLUMNS

const selectColumns = `
    id,
    organization_id,
    name,
    created_at,
    updated_at
`

// GET BY ID

func (r *Repository) GetByID(
	ctx context.Context,
	organizationID string,
	sizeID string,
) (*entity.Size, error) {
	size, err := scanSize(r.db.QueryRowContext(
		ctx,
		`SELECT `+selectColumns+`
         FROM sizes
         WHERE id = $1
         AND organization_id = $2`,
		sizeID,
		organizationID,
	))

	if errors.Is(err, sql.ErrNoRows) {
		return nil, repo.ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return size, nil
}

// LIST

func (r *Repository) List(
	ctx context.Context,
	organizationID string,
) ([]*entity.Size, error) {
	query := `
        SELECT ` + selectColumns + `
        FROM sizes
        WHERE organization_id = $1
        ORDER BY created_at ASC
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

	sizes := make([]*entity.Size, 0)

	for rows.Next() {
		size, err := scanSize(rows)
		if err != nil {
			return nil, err
		}

		sizes = append(sizes, size)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sizes, nil
}

// UPDATE

func (r *Repository) Update(
	ctx context.Context,
	organizationID string,
	size *entity.Size,
) error {
	query := `
        UPDATE sizes
        SET name = $1,
            updated_at = NOW()
        WHERE id = $2
        AND organization_id = $3
    `

	result, err := r.db.ExecContext(
		ctx,
		query,
		size.Name,
		size.ID,
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

// DELETE

func (r *Repository) Delete(
	ctx context.Context,
	organizationID string,
	sizeID string,
) error {
	query := `
        DELETE FROM sizes
        WHERE id = $1
        AND organization_id = $2
    `

	result, err := r.db.ExecContext(
		ctx,
		query,
		sizeID,
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
