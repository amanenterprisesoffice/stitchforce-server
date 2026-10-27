
package organization

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

func New(db *sql.DB) repo.OrganizationRepository {
    return &Repository{db: db}
}

func (r *Repository) Create(
    ctx context.Context,
    org *entity.Organization,
) error {
    query := `
        INSERT INTO organizations (name)
        VALUES ($1)
        RETURNING id, license_status, license_expires_at,
                  created_at, updated_at
    `

    return r.db.QueryRowContext(
        ctx, query, org.Name,
    ).Scan(
        &org.ID,
        &org.LicenseStatus,
        &org.LicenseExpiresAt,
        &org.CreatedAt,
        &org.UpdatedAt,
    )
}

func scanOrganization(
    row interface {
        Scan(...any) error
    },
) (*entity.Organization, error) {
    org := &entity.Organization{}

    err := row.Scan(
        &org.ID,
        &org.Name,
        &org.LicenseStatus,
        &org.LicenseExpiresAt,
        &org.CreatedAt,
        &org.UpdatedAt,
    )
    if err != nil {
        return nil, err
    }

    return org, nil
}

const selectColumns = `
    id, name, license_status, license_expires_at,
    created_at, updated_at
`

func (r *Repository) GetByID(
    ctx context.Context,
    id string,
) (*entity.Organization, error) {
    org, err := scanOrganization(r.db.QueryRowContext(
        ctx,
        `SELECT `+selectColumns+`
         FROM organizations
         WHERE id = $1`,
        id,
    ))

    if errors.Is(err, sql.ErrNoRows) {
        return nil, repo.ErrNotFound
    }
    if err != nil {
        return nil, err
    }

    return org, nil
}

func (r *Repository) List(
    ctx context.Context,
) ([]*entity.Organization, error) {
    rows, err := r.db.QueryContext(
        ctx,
        `SELECT `+selectColumns+`
         FROM organizations
         ORDER BY created_at DESC`,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    organizations := make([]*entity.Organization, 0)

    for rows.Next() {
        org, err := scanOrganization(rows)
        if err != nil {
            return nil, err
        }

        organizations = append(organizations, org)
    }

    if err := rows.Err(); err != nil {
        return nil, err
    }

    return organizations, nil
}

func (r *Repository) Update(
    ctx context.Context,
    org *entity.Organization,
) error {
    result, err := r.db.ExecContext(
        ctx,
        `UPDATE organizations
         SET name = $2,
             updated_at = NOW()
         WHERE id = $1`,
        org.ID,
        org.Name,
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
        `DELETE FROM organizations WHERE id = $1`,
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
