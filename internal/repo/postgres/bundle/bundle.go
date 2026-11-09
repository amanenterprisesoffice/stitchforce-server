package bundle

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"stitchforce-server/internal/entity"
	"stitchforce-server/internal/repo"
	"stitchforce-server/internal/usecase/input"
	"stitchforce-server/internal/usecase/output"
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) repo.BundleRepository {
	return &Repository{db: db}
}

// CREATE

func (r *Repository) Create(
	ctx context.Context,
	organizationID string,
	bundle *entity.Bundle,
) error {
	query := `
        INSERT INTO bundles (
            id,
            organization_id,
            bundle_number,
            style_id,
            size_id,
            quantity,
            attributes,
            status,
            remarks
        )
        VALUES (
            $1, $2, $3, $4, $5, $6,
            COALESCE($7::jsonb, '{}'::jsonb),
            $8, $9
        )
        RETURNING created_at, updated_at
    `

	return r.db.QueryRowContext(
		ctx,
		query,
		bundle.ID,
		organizationID,
		bundle.BundleNumber,
		bundle.StyleID,
		bundle.SizeID,
		bundle.Quantity,
		[]byte(bundle.Attributes),
		bundle.Status,
		bundle.Remarks,
	).Scan(
		&bundle.CreatedAt,
		&bundle.UpdatedAt,
	)
}

// SCAN

func scanBundle(
	row interface {
		Scan(...any) error
	},
) (*entity.Bundle, error) {
	bundle := &entity.Bundle{}

	var attributes []byte

	err := row.Scan(
		&bundle.ID,
		&bundle.OrganizationID,
		&bundle.BundleNumber,
		&bundle.StyleID,
		&bundle.SizeID,
		&bundle.Quantity,
		&attributes,
		&bundle.Status,
		&bundle.Remarks,
		&bundle.CreatedAt,
		&bundle.UpdatedAt,
		&bundle.DeletedAt,
	)
	if err != nil {
		return nil, err
	}

	bundle.Attributes = attributes

	return bundle, nil
}

// SELECT COLUMNS

const selectColumns = `
    id,
    organization_id,
    bundle_number,
    style_id,
    size_id,
    quantity,
    attributes,
    status,
    remarks,
    created_at,
    updated_at,
    deleted_at
`

// GET BY ID

func (r *Repository) GetByID(
	ctx context.Context,
	organizationID string,
	bundleID string,
) (*output.BundleDetail, error) {

	bundle, err := scanBundle(r.db.QueryRowContext(
		ctx,
		`SELECT `+selectColumns+`
         FROM bundles
         WHERE id = $1
         AND organization_id = $2
         AND deleted_at IS NULL`,
		bundleID,
		organizationID,
	))

	if errors.Is(err, sql.ErrNoRows) {
		return nil, repo.ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	detail := &output.BundleDetail{
		ID:             bundle.ID,
		OrganizationID: bundle.OrganizationID,
		BundleNumber:   bundle.BundleNumber,
		StyleID:        bundle.StyleID,
		SizeID:         bundle.SizeID,
		Quantity:       bundle.Quantity,
		Attributes:     bundle.Attributes,
		Status:         bundle.Status,
		Remarks:        bundle.Remarks,
		CreatedAt:      bundle.CreatedAt,
		UpdatedAt:      bundle.UpdatedAt,
		Operations:     make([]output.BundleOperationProgress, 0),
	}

	query := `
        SELECT
            so.id,
            so.name,
            so.rate_per_piece,

            gw.id,
            gw.name,
            gw.phone,
            gw.photo_key,

            COALESCE(SUM(pe.quantity), 0) AS completed_quantity

        FROM style_operations so

        LEFT JOIN production_entries pe
            ON pe.style_operation_id = so.id
            AND pe.bundle_id = $1
            AND pe.organization_id = $2

        LEFT JOIN garment_workers gw
            ON gw.id = pe.worker_id
            AND gw.organization_id = $2

        WHERE so.style_id = $3
            AND so.organization_id = $2

        GROUP BY
            so.id,
            so.name,
            so.rate_per_piece,
            so.created_at,
            gw.id,
            gw.name,
            gw.phone,
            gw.photo_key

        ORDER BY so.created_at ASC
    `

	rows, err := r.db.QueryContext(
		ctx,
		query,
		bundleID,
		organizationID,
		bundle.StyleID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	operationIndex := make(map[string]int)

	for rows.Next() {
		var (
			operationID   string
			operationName string
			rate          int64

			workerID    sql.NullString
			workerName  sql.NullString
			workerPhone sql.NullString
			photoKey    sql.NullString

			completedQuantity int
		)

		if err := rows.Scan(
			&operationID,
			&operationName,
			&rate,
			&workerID,
			&workerName,
			&workerPhone,
			&photoKey,
			&completedQuantity,
		); err != nil {
			return nil, err
		}

		operation := output.BundleOperationProgress{
			ID:                operationID,
			Name:              operationName,
			RatePerPiece:      rate,
			CompletedQuantity: completedQuantity,
			PendingQuantity:   bundle.Quantity - completedQuantity,
		}

		if workerID.Valid {
			operation.Worker = &output.SlimWorker{
				ID:       workerID.String,
				Name:     workerName.String,
				Phone:    workerPhone.String,
				PhotoKey: photoKey.String,
			}
		}

		// One operation should have one worker.
		if _, exists := operationIndex[operationID]; exists {
			return nil, fmt.Errorf(
				"multiple workers found for operation %s in bundle %s",
				operationID,
				bundleID,
			)
		}

		operationIndex[operationID] = len(detail.Operations)
		detail.Operations = append(detail.Operations, operation)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return detail, nil
}

// LIST

func (r *Repository) List(
	ctx context.Context,
	organizationID string,
	filter input.BundleFilter,
) ([]*entity.Bundle, error) {

	query := `
		SELECT ` + selectColumns + `
		FROM bundles
		WHERE organization_id = $1
		  AND deleted_at IS NULL
	`

	args := []any{organizationID}
	argPosition := 2

	if filter.Search != "" {
		query += fmt.Sprintf(" AND bundle_number ILIKE $%d", argPosition)
		args = append(args, "%"+filter.Search+"%")
		argPosition++
	}

	if filter.StyleID != "" {
		query += fmt.Sprintf(" AND style_id = $%d", argPosition)
		args = append(args, filter.StyleID)
		argPosition++
	}

	if filter.SizeID != "" {
		query += fmt.Sprintf(" AND size_id = $%d", argPosition)
		args = append(args, filter.SizeID)
		argPosition++
	}

	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argPosition)
		args = append(args, filter.Status)
		argPosition++
	}

	if filter.CreatedFrom != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argPosition)
		args = append(args, *filter.CreatedFrom)
		argPosition++
	}

	if filter.CreatedTo != nil {
		query += fmt.Sprintf(" AND created_at < $%d", argPosition)
		args = append(args, *filter.CreatedTo)
		argPosition++
	}

	query += " ORDER BY created_at DESC"

	// Pagination.
	limit := filter.Limit
	offset := filter.Offset

	query += fmt.Sprintf(
		" LIMIT $%d OFFSET $%d",
		argPosition,
		argPosition+1,
	)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(
		ctx,
		query,
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bundles := make([]*entity.Bundle, 0)

	for rows.Next() {
		bundle, err := scanBundle(rows)
		if err != nil {
			return nil, err
		}

		bundles = append(bundles, bundle)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return bundles, nil
}

// UPDATE

func (r *Repository) Update(
	ctx context.Context,
	organizationID string,
	bundleID string,
	input *input.UpdateBundleInput,
) error {
	query := `
		UPDATE bundles
		SET
			bundle_number = COALESCE($1, bundle_number),
			size_id = COALESCE($2, size_id),
			quantity = COALESCE($3, quantity),
			attributes = COALESCE($4::jsonb, attributes),
			status = COALESCE($5, status),
			remarks = COALESCE($6, remarks),
			updated_at = NOW()
		WHERE id = $7
		AND organization_id = $8
		AND deleted_at IS NULL
	`

	var attributes any
	if input.Attributes != nil {
		attributes = []byte(*input.Attributes)
	}

	result, err := r.db.ExecContext(
		ctx,
		query,
		input.BundleNumber,
		input.SizeID,
		input.Quantity,
		attributes,
		input.Status,
		input.Remarks,
		bundleID,
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

// DELETE (SOFT DELETE)

func (r *Repository) Delete(
	ctx context.Context,
	organizationID string,
	bundleID string,
) error {
	query := `
        UPDATE bundles
        SET
            deleted_at = NOW(),
            updated_at = NOW()
        WHERE id = $1
        AND organization_id = $2
        AND deleted_at IS NULL
    `

	result, err := r.db.ExecContext(
		ctx,
		query,
		bundleID,
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
