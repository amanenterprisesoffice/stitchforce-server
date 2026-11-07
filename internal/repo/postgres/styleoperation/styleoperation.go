package styleoperation

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

func New(db *sql.DB) repo.StyleOperationRepository {
	return &Repository{db: db}
}

func (r *Repository) Create(
	ctx context.Context,
	organizationID string,
	styleID string,
	operation *entity.StyleOperation,
) error {
	query := `
		INSERT INTO style_operations (
			id,
			organization_id,
			style_id,
			name,
			rate_per_piece
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	return r.db.QueryRowContext(
		ctx,
		query,
		operation.ID,
		organizationID,
		styleID,
		operation.Name,
		operation.RatePerPiece,
	).Scan(&operation.ID)
}

func (r *Repository) GetByID(
	ctx context.Context,
	organizationID string,
	styleID string,
	id string,
) (*entity.StyleOperation, error) {
	query := `
		SELECT
			id,
			organization_id,
			style_id,
			name,
			rate_per_piece,
			created_at,
			updated_at
		FROM style_operations
		WHERE id = $1
		  AND organization_id = $2
		  AND style_id = $3
	`

	operation := &entity.StyleOperation{}

	err := r.db.QueryRowContext(
		ctx,
		query,
		id,
		organizationID,
		styleID,
	).Scan(
		&operation.ID,
		&operation.OrganizationID,
		&operation.StyleID,
		&operation.Name,
		&operation.RatePerPiece,
		&operation.CreatedAt,
		&operation.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, repo.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return operation, nil
}

func (r *Repository) List(
	ctx context.Context,
	organizationID string,
	styleID string,
	filter input.StyleOperationFilter,
) ([]*entity.StyleOperation, error) {
	query := `
		SELECT
			id,
			organization_id,
			style_id,
			name,
			rate_per_piece,
			created_at,
			updated_at
		FROM style_operations
		WHERE organization_id = $1
		  AND style_id = $2
	`

	args := []any{organizationID, styleID}
	argPosition := 3

	if filter.Search != "" {
		query += fmt.Sprintf(" AND name ILIKE $%d", argPosition)
		args = append(args, "%"+filter.Search+"%")
		argPosition++
	}

	query += " ORDER BY created_at ASC"

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

	operations := make([]*entity.StyleOperation, 0)

	for rows.Next() {
		operation := &entity.StyleOperation{}

		if err := rows.Scan(
			&operation.ID,
			&operation.OrganizationID,
			&operation.StyleID,
			&operation.Name,
			&operation.RatePerPiece,
			&operation.CreatedAt,
			&operation.UpdatedAt,
		); err != nil {
			return nil, err
		}

		operations = append(operations, operation)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return operations, nil
}

func (r *Repository) Update(
	ctx context.Context,
	organizationID string,
	styleID string,
	operation *entity.StyleOperation,
) error {
	query := `
		UPDATE style_operations
		SET name = $4,
		    rate_per_piece = $5,
		    updated_at = NOW()
		WHERE id = $1
		  AND organization_id = $2
		  AND style_id = $3
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		operation.ID,
		organizationID,
		styleID,
		operation.Name,
		operation.RatePerPiece,
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
	styleID string,
	id string,
) error {
	query := `
		DELETE FROM style_operations
		WHERE id = $1
		  AND organization_id = $2
		  AND style_id = $3
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		id,
		organizationID,
		styleID,
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
