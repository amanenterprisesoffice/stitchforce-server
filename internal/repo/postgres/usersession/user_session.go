package postgres

import (
	"context"
	"database/sql"
	"errors"

	"stitchforce-server/internal/entity"
	"stitchforce-server/internal/repo"
)

type SessionRepository struct {
	db *sql.DB
}

func New(db *sql.DB) repo.UserSessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) CreateWithLimit(
	ctx context.Context,
	session *entity.UserSession,
	maxSessions int,
) error {
	if maxSessions < 1 {
		return errors.New("max sessions must be at least 1")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Serialize logins for this user.
	var userID string
	err = tx.QueryRowContext(
		ctx,
		`SELECT id FROM users WHERE id = $1 FOR UPDATE`,
		session.UserID,
	).Scan(&userID)
	if err != nil {
		return err
	}

	// Keep room for the new session.
	keep := maxSessions - 1

	_, err = tx.ExecContext(
		ctx,
		`
		UPDATE user_sessions
		SET revoked_at = NOW()
		WHERE id IN (
			SELECT id
			FROM user_sessions
			WHERE user_id = $1
			  AND revoked_at IS NULL
			  AND expires_at > NOW()
			ORDER BY created_at DESC, id DESC
			OFFSET $2
		)
		`,
		session.UserID,
		keep,
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		`
		INSERT INTO user_sessions (
			id,
			user_id,
			organization_id,
			device_name,
			ip_address,
			user_agent,
			expires_at,
			created_at,
			last_used_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		`,
		session.ID,
		session.UserID,
		session.OrganizationID,
		session.DeviceName,
		session.IPAddress,
		session.UserAgent,
		session.ExpiresAt,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *SessionRepository) GetByID(
	ctx context.Context,
	id string,
) (*entity.UserSession, error) {
	session, err := scanUserSession(
		r.db.QueryRowContext(
			ctx,
			`
			SELECT
				id,
				user_id,
				organization_id,
				device_name,
				ip_address,
				user_agent,
				expires_at,
				revoked_at,
				created_at,
				last_used_at
			FROM user_sessions
			WHERE id = $1
			`,
			id,
		),
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repo.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (r *SessionRepository) GetActive(
	ctx context.Context,
	id string,
) (*entity.UserSession, error) {
	session, err := scanUserSession(
		r.db.QueryRowContext(
			ctx,
			`
			SELECT
				id,
				user_id,
				organization_id,
				device_name,
				ip_address,
				user_agent,
				expires_at,
				revoked_at,
				created_at,
				last_used_at
			FROM user_sessions
			WHERE id = $1
			  AND revoked_at IS NULL
			  AND expires_at > NOW()
			`,
			id,
		),
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repo.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (r *SessionRepository) Revoke(
	ctx context.Context,
	id string,
) error {
	_, err := r.db.ExecContext(
		ctx,
		`
		UPDATE user_sessions
		SET revoked_at = NOW()
		WHERE id = $1
		  AND revoked_at IS NULL
		`,
		id,
	)
	return err
}

func (r *SessionRepository) RevokeAll(
	ctx context.Context,
	userID string,
) error {
	_, err := r.db.ExecContext(
		ctx,
		`
		UPDATE user_sessions
		SET revoked_at = NOW()
		WHERE user_id = $1
		  AND revoked_at IS NULL
		`,
		userID,
	)
	return err
}

func (r *SessionRepository) RevokeOrganizationSessions(
	ctx context.Context,
	organizationID string,
) error {
	_, err := r.db.ExecContext(
		ctx,
		`
		UPDATE user_sessions
		SET revoked_at = NOW()
		WHERE organization_id = $1
		  AND revoked_at IS NULL
		`,
		organizationID,
	)
	return err
}

func scanUserSession(
	row interface {
		Scan(dest ...any) error
	},
) (*entity.UserSession, error) {
	var session entity.UserSession
	var revokedAt sql.NullTime

	err := row.Scan(
		&session.ID,
		&session.UserID,
		&session.OrganizationID,
		&session.DeviceName,
		&session.IPAddress,
		&session.UserAgent,
		&session.ExpiresAt,
		&revokedAt,
		&session.CreatedAt,
		&session.LastUsedAt,
	)
	if err != nil {
		return nil, err
	}

	if revokedAt.Valid {
		session.RevokedAt = &revokedAt.Time
	}

	return &session, nil
}

// Optional helper if you need to update last activity.
func (r *SessionRepository) UpdateLastUsed(
	ctx context.Context,
	id string,
) error {
	_, err := r.db.ExecContext(
		ctx,
		`
		UPDATE user_sessions
		SET last_used_at = NOW()
		WHERE id = $1
		  AND revoked_at IS NULL
		`,
		id,
	)
	return err
}
