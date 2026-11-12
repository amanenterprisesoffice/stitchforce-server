package organizationinvitation

import (
	"context"
	"database/sql"
	"errors"
	"stitchforce-server/internal/entity"
	"stitchforce-server/internal/repo"
	"time"
)

type OrganizationInvitationRepository struct {
	db *sql.DB
}

func New(db *sql.DB) repo.OrganizationInvitationRepository {
	return &OrganizationInvitationRepository{db: db}
}

func (r *OrganizationInvitationRepository) Upsert(
	ctx context.Context,
	invitation *entity.OrganizationInvitation,
) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO organization_invitations (
			id,
			organization_id,
			email,
			role,
			token_hash,
			expires_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (organization_id, email)
		DO UPDATE SET
			role = EXCLUDED.role,
			token_hash = EXCLUDED.token_hash,
			expires_at = EXCLUDED.expires_at,
			accepted_at = NULL
	`,
		invitation.ID,
		invitation.OrganizationID,
		invitation.Email,
		invitation.Role,
		invitation.TokenHash,
		invitation.ExpiresAt,
	)

	return err
}

func (r *OrganizationInvitationRepository) GetByTokenHash(
	ctx context.Context,
	tokenHash string,
) (*entity.OrganizationInvitation, error) {
	var invitation entity.OrganizationInvitation
	var acceptedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, `
		SELECT
			id,
			organization_id,
			email,
			role,
			token_hash,
			expires_at,
			accepted_at,
			created_at
		FROM organization_invitations
		WHERE token_hash = $1
	`,
		tokenHash,
	).Scan(
		&invitation.ID,
		&invitation.OrganizationID,
		&invitation.Email,
		&invitation.Role,
		&invitation.TokenHash,
		&invitation.ExpiresAt,
		&acceptedAt,
		&invitation.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}

	if acceptedAt.Valid {
		invitation.AcceptedAt = &acceptedAt.Time
	}

	return &invitation, nil
}

func (r *OrganizationInvitationRepository) GetByID(
	ctx context.Context,
	organizationID string,
	invitationID string,
) (*entity.OrganizationInvitation, error) {
	var invitation entity.OrganizationInvitation
	var acceptedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, `
		SELECT
			id,
			organization_id,
			email,
			role,
			token_hash,
			expires_at,
			accepted_at,
			created_at
		FROM organization_invitations
		WHERE organization_id = $1 AND id = $2
	`,
		organizationID,
		invitationID,
	).Scan(
		&invitation.ID,
		&invitation.OrganizationID,
		&invitation.Email,
		&invitation.Role,
		&invitation.TokenHash,
		&invitation.ExpiresAt,
		&acceptedAt,
		&invitation.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}

	if acceptedAt.Valid {
		invitation.AcceptedAt = &acceptedAt.Time
	}

	return &invitation, nil
}

func (r *OrganizationInvitationRepository) ListByOrganization(
	ctx context.Context,
	organizationID string,
) ([]*entity.OrganizationInvitation, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			id,
			organization_id,
			email,
			role,
			token_hash,
			expires_at,
			accepted_at,
			created_at
		FROM organization_invitations
		WHERE organization_id = $1
		ORDER BY created_at DESC
	`,
		organizationID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	invitations := make([]*entity.OrganizationInvitation, 0)

	for rows.Next() {
		var invitation entity.OrganizationInvitation
		var acceptedAt sql.NullTime

		if err := rows.Scan(
			&invitation.ID,
			&invitation.OrganizationID,
			&invitation.Email,
			&invitation.Role,
			&invitation.TokenHash,
			&invitation.ExpiresAt,
			&acceptedAt,
			&invitation.CreatedAt,
		); err != nil {
			return nil, err
		}

		if acceptedAt.Valid {
			invitation.AcceptedAt = &acceptedAt.Time
		}

		invitations = append(invitations, &invitation)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return invitations, nil
}

func (r *OrganizationInvitationRepository) Delete(
	ctx context.Context,
	organizationID string,
	invitationID string,
) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM organization_invitations
		WHERE organization_id = $1 AND id = $2
	`,
		organizationID,
		invitationID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return repo.ErrNotFound
	}

	return nil
}

func (r *OrganizationInvitationRepository) Accept(
	ctx context.Context,
	invitationID string,
	u *entity.User,
) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var organizationID string
	var email string
	var role entity.UserRole
	var expiresAt time.Time
	var acceptedAt sql.NullTime

	err = tx.QueryRowContext(ctx, `
		SELECT
			organization_id,
			email,
			role,
			expires_at,
			accepted_at
		FROM organization_invitations
		WHERE id = $1
		FOR UPDATE
	`, invitationID).Scan(
		&organizationID,
		&email,
		&role,
		&expiresAt,
		&acceptedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repo.ErrNotFound
		}
		return err
	}

	if acceptedAt.Valid {
		return errors.New("invitation already accepted")
	}

	if !expiresAt.After(time.Now()) {
		return errors.New("invitation expired")
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO users (
			id,
			organization_id,
			name,
			email,
			password_hash,
			role,
			email_verified_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`,
		u.ID,
		organizationID,
		u.Name,
		email,
		u.PasswordHash,
		role,
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE organization_invitations
		SET accepted_at = NOW()
		WHERE id = $1
	`, invitationID)
	if err != nil {
		return err
	}

	return tx.Commit()
}