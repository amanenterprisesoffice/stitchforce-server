package style

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

func New(db *sql.DB) repo.StyleRepository {
	return &Repository{db: db}
}

func (r *Repository) Create(
	ctx context.Context,
	organizationID string,
	style *entity.Style,
) error {
	query := `
		INSERT INTO styles (
			id,
			organization_id,
			name,
			style_number,
			description
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	return r.db.QueryRowContext(
		ctx,
		query,
		style.ID,
		organizationID,
		style.Name,
		style.StyleNumber,
		style.Description,
	).Scan(&style.ID)
}

func (r *Repository) GetByID(
	ctx context.Context,
	organizationID string,
	id string,
) (*entity.Style, error) {
	query := `
		SELECT
			id,
			organization_id,
			name,
			style_number,
			description,
			created_at,
			updated_at
		FROM styles
		WHERE id = $1
		  AND organization_id = $2
	`

	style := &entity.Style{}

	err := r.db.QueryRowContext(
		ctx,
		query,
		id,
		organizationID,
	).Scan(
		&style.ID,
		&style.OrganizationID,
		&style.Name,
		&style.StyleNumber,
		&style.Description,
		&style.CreatedAt,
		&style.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, repo.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return style, nil
}

func (r *Repository) List(
	ctx context.Context,
	organizationID string,
	filter input.StyleFilter,
) ([]*entity.Style, error) {
	query := `
		SELECT
			id,
			organization_id,
			name,
			style_number,
			description,
			created_at,
			updated_at
		FROM styles
		WHERE organization_id = $1
	`

	args := []any{organizationID}
	argPosition := 2

	if filter.Search != "" {
		query += fmt.Sprintf(`
			AND (
				name ILIKE $%d
				OR style_number ILIKE $%d
			)
		`, argPosition, argPosition)

		args = append(args, "%"+filter.Search+"%")
		argPosition++
	}

	query += " ORDER BY created_at DESC"

	limit := filter.Limit
	offset := filter.Offset
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

	styles := make([]*entity.Style, 0)

	for rows.Next() {
		style := &entity.Style{}

		if err := rows.Scan(
			&style.ID,
			&style.OrganizationID,
			&style.Name,
			&style.StyleNumber,
			&style.Description,
			&style.CreatedAt,
			&style.UpdatedAt,
		); err != nil {
			return nil, err
		}

		styles = append(styles, style)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return styles, nil
}

func (r *Repository) Update(
	ctx context.Context,
	organizationID string,
	style *entity.Style,
) error {
	query := `
		UPDATE styles
		SET name = $3,
		    style_number = $4,
		    description = $5,
		    updated_at = NOW()
		WHERE id = $1
		  AND organization_id = $2
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		style.ID,
		organizationID,
		style.Name,
		style.StyleNumber,
		style.Description,
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
	query := `
		DELETE FROM styles
		WHERE id = $1
		  AND organization_id = $2
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
