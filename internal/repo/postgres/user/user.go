package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"stitchforce-server/internal/entity"
	"stitchforce-server/internal/repo"
	"stitchforce-server/internal/usecase/input"
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) repo.UserRepository {
	return &Repository{db: db}
}

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
			id, organization_id, name, email, password_hash, role, photo_key
		)
		VALUES ($1, $2, $3, $4, $5, $6, '')
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
			id, organization_id, name, email, password_hash, role, photo_key
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`,
		u.ID,
		u.OrganizationID,
		u.Name,
		u.Email,
		u.PasswordHash,
		u.Role,
		u.PhotoKey,
	).Scan(&u.ID)
}

func (r *Repository) GetByID(
	ctx context.Context,
	id string,
) (*entity.User, error) {
	return r.getUser(ctx, `
		SELECT
			id,
			organization_id,
			name,
			email,
			role,
			password_hash,
			COALESCE(photo_key, ''),
			email_verified_at,
			created_at,
			updated_at
		FROM users
		WHERE id = $1
	`, id)
}

func (r *Repository) GetByEmail(
	ctx context.Context,
	email string,
) (*entity.User, error) {
	return r.getUser(ctx, `
		SELECT
			id,
			organization_id,
			name,
			email,
			role,
			password_hash,
			COALESCE(photo_key, ''),
			email_verified_at,
			created_at,
			updated_at
		FROM users
		WHERE email = $1
	`, email)
}

func (r *Repository) getUser(
	ctx context.Context,
	query string,
	args ...any,
) (*entity.User, error) {
	u := &entity.User{}

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&u.ID,
		&u.OrganizationID,
		&u.Name,
		&u.Email,
		&u.Role,
		&u.PasswordHash,
		&u.PhotoKey,
		&u.EmailVerifiedAt,
		&u.CreatedAt,
		&u.UpdatedAt,
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
	organizationID string,
	filter input.UserFilter,
) ([]*entity.User, error) {
	query := `
		SELECT
			id,
			organization_id,
			name,
			email,
			role,
			password_hash,
			COALESCE(photo_key, ''),
			email_verified_at,
			created_at,
			updated_at
		FROM users
		WHERE organization_id = $1
	`

	args := []any{organizationID}
	argPosition := 2

	if filter.Role != "" {
		query += fmt.Sprintf(" AND role = $%d", argPosition)
		args = append(args, filter.Role)
		argPosition++
	}

	if filter.Search != "" {
		query += fmt.Sprintf(
			" AND (name ILIKE $%d OR email ILIKE $%d)",
			argPosition,
			argPosition,
		)
		args = append(args, "%"+filter.Search+"%")
		argPosition++
	}

	query += " ORDER BY created_at DESC"

	query += fmt.Sprintf(
		" LIMIT $%d OFFSET $%d",
		argPosition,
		argPosition+1,
	)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
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
			&u.PhotoKey,
			&u.EmailVerifiedAt,
			&u.CreatedAt,
			&u.UpdatedAt,
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
		SET
			name = COALESCE(NULLIF($2, ''), name),
			email = COALESCE(NULLIF($3, ''), email),
			role = COALESCE(NULLIF($4, ''), role),
			photo_key = COALESCE(NULLIF($5, ''), photo_key),
			updated_at = NOW()
		WHERE id = $1
	`,
		u.ID,
		u.Name,
		u.Email,
		u.Role,
		u.PhotoKey,
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
	id string,
) error {
	result, err := r.db.ExecContext(
		ctx,
		`DELETE FROM users WHERE id = $1`,
		id,
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

func (r *Repository) CreateEmailVerificationToken(
	ctx context.Context,
	userID string,
	tokenHash string,
	expiresAt time.Time,
) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE users
		SET
			email_verification_token_hash = $2,
			email_verification_expires_at = $3,
			updated_at = NOW()
		WHERE id = $1
		  AND email_verified_at IS NULL
	`,
		userID,
		tokenHash,
		expiresAt,
	)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *Repository) VerifyEmail(
	ctx context.Context,
	tokenHash string,
) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE users
		SET
			email_verified_at = NOW(),
			email_verification_token_hash = NULL,
			email_verification_expires_at = NULL,
			updated_at = NOW()
		WHERE email_verification_token_hash = $1
		  AND email_verification_expires_at > NOW()
		  AND email_verified_at IS NULL
	`, tokenHash)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}