
package user

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

func New(db *sql.DB) repo.UserRepository {
	return &Repository{db: db}
}

// RegisterOrganization creates the organization and its first admin
// atomically. Both inserts succeed, or neither does.
func (r *Repository) RegisterOrganization(
	ctx context.Context,
	org *entity.Organization,
	u *entity.User,
) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO organizations (
			id, name, license_status, license_expires_at
		)
		VALUES ($1, $2, $3, $4)
	`,
		org.ID,
		org.Name,
		org.LicenseStatus,
		org.LicenseExpiresAt,
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO users (
			id, organization_id, name, email, password_hash, role
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`,
		u.ID,
		org.ID,
		u.Name,
		u.Email,
		u.PasswordHash,
		u.Role,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *Repository) Create(
	ctx context.Context,
	u *entity.User,
) error {
	return r.db.QueryRowContext(ctx, `
		INSERT INTO users (
			id, organization_id, name, email, password_hash, role
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`,
		u.ID,
		u.OrganizationID,
		u.Name,
		u.Email,
		u.PasswordHash,
		u.Role,
	).Scan(&u.ID)
}

func (r *Repository) GetByID(
	ctx context.Context,
	id string,
) (*entity.User, error) {
	return r.getUser(ctx, `
		SELECT id, organization_id, name, email, role, password_hash
		FROM users WHERE id = $1
	`, id)
}

func (r *Repository) GetByEmail(
	ctx context.Context,
	email string,
) (*entity.User, error) {
	return r.getUser(ctx, `
		SELECT id, organization_id, name, email, role, password_hash
		FROM users WHERE email = $1
	`, email)
}

func (r *Repository) getUser(
	ctx context.Context,
	query string,
	arg any,
) (*entity.User, error) {
	u := &entity.User{}

	err := r.db.QueryRowContext(ctx, query, arg).Scan(
		&u.ID,
		&u.OrganizationID,
		&u.Name,
		&u.Email,
		&u.Role,
		&u.PasswordHash,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repo.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (r *Repository) List(
	ctx context.Context,
) ([]*entity.User, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, organization_id, name, email, role, password_hash
		FROM users
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]*entity.User, 0)

	for rows.Next() {
		u := &entity.User{}
		if err := rows.Scan(
			&u.ID,
			&u.OrganizationID,
			&u.Name,
			&u.Email,
			&u.Role,
			&u.PasswordHash,
		); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *Repository) Update(
	ctx context.Context,
	u *entity.User,
) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE users
		SET name = $2, email = $3, role = $4
		WHERE id = $1
	`,
		u.ID, u.Name, u.Email, u.Role,
	)
	if err != nil {
		return err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return repo.ErrNotFound
	}

	return nil
}

func (r *Repository) Delete(
	ctx context.Context,
	id string,
) error {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM users WHERE id = $1`, id,
	)
	if err != nil {
		return err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return repo.ErrNotFound
	}

	return nil
}
