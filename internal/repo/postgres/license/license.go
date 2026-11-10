package license

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"stitchforce-server/internal/entity"
	"stitchforce-server/internal/repo"
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) repo.LicenseRepository {
	return &Repository{db: db}
}

func (r *Repository) Create(
	ctx context.Context,
	license *entity.License,
) error {
	query := `
		INSERT INTO licenses (
			license_key_hash,
			organization_id,
			duration_days,
			status
		)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	return r.db.QueryRowContext(
		ctx,
		query,
		license.LicenseKeyHash,
		license.OrganizationID,
		license.DurationDays,
		license.Status,
	).Scan(
		&license.ID,
		&license.CreatedAt,
	)
}

func (r *Repository) Activate(
	ctx context.Context,
	licenseKeyHash string,
	organizationID string,
) (*entity.License, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var license entity.License

	// 1. Lock and retrieve the license.
	err = tx.QueryRowContext(
		ctx,
		`
			SELECT
				id,
				organization_id,
				duration_days,
				status
			FROM licenses
			WHERE license_key_hash = $1
			FOR UPDATE
		`,
		licenseKeyHash,
	).Scan(
		&license.ID,
		&license.OrganizationID,
		&license.DurationDays,
		&license.Status,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("license not found")
		}
		return nil, err
	}

	// 2. Check license status.
	switch license.Status {
	case "activated":
		return nil, errors.New("license already activated")
	case "revoked":
		return nil, errors.New("license has been revoked")
	case "unused":
		// Continue.
	default:
		return nil, errors.New("invalid license status")
	}

	// 3. Verify organization ownership.
	if license.OrganizationID == nil ||
		*license.OrganizationID != organizationID {
		return nil, errors.New("license does not belong to this organization")
	}

	// 4. Lock and retrieve organization expiry and status.
	var (
		currentExpiry sql.NullTime
		orgStatus     string
	)

	err = tx.QueryRowContext(
		ctx,
		`
			SELECT
				license_expires_at,
				license_status
			FROM organizations
			WHERE id = $1
			FOR UPDATE
		`,
		organizationID,
	).Scan(
		&currentExpiry,
		&orgStatus,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("organization not found")
		}
		return nil, err
	}

	if orgStatus == "suspended" {
		return nil, errors.New("organization is suspended")
	}

	// 5. Calculate new expiry.
	var newExpiry time.Time

	err = tx.QueryRowContext(
		ctx,
		`
			SELECT
				CASE
					WHEN $1::timestamptz > NOW()
					THEN $1::timestamptz
					ELSE NOW()
				END + ($2 * INTERVAL '1 day')
		`,
		nullTimeValue(currentExpiry),
		license.DurationDays,
	).Scan(&newExpiry)

	if err != nil {
		return nil, err
	}

	// 6. Activate the license.
	err = tx.QueryRowContext(
		ctx,
		`
			UPDATE licenses
			SET
				status = 'activated',
				activated_at = NOW()
			WHERE id = $1
			RETURNING activated_at
		`,
		license.ID,
	).Scan(&license.ActivatedAt)

	if err != nil {
		return nil, err
	}

	// 7. Update organization license.
	_, err = tx.ExecContext(
		ctx,
		`
			UPDATE organizations
			SET
				license_status = 'active',
				license_expires_at = $1,
				updated_at = NOW()
			WHERE id = $2
		`,
		newExpiry,
		organizationID,
	)

	if err != nil {
		return nil, err
	}

	// 8. Commit transaction.
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	license.Status = "activated"

	return &license, nil
}

func nullTimeValue(value sql.NullTime) interface{} {
	if !value.Valid {
		return nil
	}

	return value.Time
}

var _ = fmt.Sprintf