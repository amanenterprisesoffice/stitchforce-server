package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
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

func New(db *sql.DB) repo.ReportRepository {
	return &Repository{db: db}
}

const reportColumns = `
	id,
	organization_id,
	type,
	year,
	month,
	file_key,
	file_size,
	record_counts,
	archived_at,
	expires_at,
	created_at,
	cleanup_completed,
	expiration_reminder_sent_at
`

func scanReport(scanner interface {
	Scan(dest ...any) error
}) (*entity.Report, error) {
	report := &entity.Report{}
	var recordCounts []byte

	err := scanner.Scan(
		&report.ID,
		&report.OrganizationID,
		&report.Type,
		&report.Year,
		&report.Month,
		&report.FileKey,
		&report.FileSize,
		&recordCounts,
		&report.ArchivedAt,
		&report.ExpiresAt,
		&report.CreatedAt,
		&report.CleanupCompleted,
		&report.ExpirationReminderSentAt,
	)
	if err != nil {
		return nil, err
	}

	if len(recordCounts) > 0 {
		if err := json.Unmarshal(recordCounts, &report.RecordCounts); err != nil {
			return nil, fmt.Errorf("decode report record counts: %w", err)
		}
	}

	if report.RecordCounts == nil {
		report.RecordCounts = make(map[string]int)
	}

	return report, nil
}

func (r *Repository) Create(
	ctx context.Context,
	report *entity.Report,
) error {
	if report == nil {
		return errors.New("report is required")
	}

	recordCounts, err := json.Marshal(report.RecordCounts)
	if err != nil {
		return fmt.Errorf("marshal report record counts: %w", err)
	}

	query := `
		INSERT INTO reports (
			id,
			organization_id,
			type,
			year,
			month,
			file_key,
			file_size,
			record_counts,
			archived_at,
			expires_at,
			cleanup_completed
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at
	`

	err = r.db.QueryRowContext(
		ctx,
		query,
		report.ID,
		report.OrganizationID,
		report.Type,
		report.Year,
		report.Month,
		report.FileKey,
		report.FileSize,
		recordCounts,
		report.ArchivedAt,
		report.ExpiresAt,
		report.CleanupCompleted,
	).Scan(&report.CreatedAt)

	if err != nil {
		return fmt.Errorf("create report: %w", err)
	}

	return nil
}

func (r *Repository) List(
	ctx context.Context,
	organizationID string,
	filter input.ReportFilter,
) ([]*entity.Report, error) {
	query := `
		SELECT ` + reportColumns + `
		FROM reports
		WHERE organization_id = $1
	`

	args := []interface{}{organizationID}
	argIndex := 2

	if filter.Year != nil {
		query += fmt.Sprintf(" AND year = $%d", argIndex)
		args = append(args, *filter.Year)
		argIndex++
	}

	if filter.Month != nil {
		query += fmt.Sprintf(" AND month = $%d", argIndex)
		args = append(args, *filter.Month)
		argIndex++
	}

	query += `
		ORDER BY year DESC, month DESC, archived_at DESC
	`

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list reports: %w", err)
	}
	defer rows.Close()

	reports := make([]*entity.Report, 0)

	for rows.Next() {
		report, err := scanReport(rows)
		if err != nil {
			return nil, fmt.Errorf("scan report: %w", err)
		}

		reports = append(reports, report)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate reports: %w", err)
	}

	return reports, nil
}

func (r *Repository) GetByID(
	ctx context.Context,
	organizationID string,
	reportID string,
) (*entity.Report, error) {
	query := `
		SELECT ` + reportColumns + `
		FROM reports
		WHERE organization_id = $1
		  AND id = $2
	`

	report, err := scanReport(
		r.db.QueryRowContext(ctx, query, organizationID, reportID),
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repo.ErrNotFound
		}

		return nil, fmt.Errorf("get report: %w", err)
	}

	return report, nil
}

func (r *Repository) ListExpired(
	ctx context.Context,
	now time.Time,
) ([]*entity.Report, error) {
	query := `
		SELECT ` + reportColumns + `
		FROM reports
		WHERE expires_at <= $1
		ORDER BY expires_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, now)
	if err != nil {
		return nil, fmt.Errorf("list expired reports: %w", err)
	}
	defer rows.Close()

	reports := make([]*entity.Report, 0)

	for rows.Next() {
		report, err := scanReport(rows)
		if err != nil {
			return nil, fmt.Errorf("scan expired report: %w", err)
		}

		reports = append(reports, report)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate expired reports: %w", err)
	}

	return reports, nil
}

func (r *Repository) ListExpiring(
	ctx context.Context,
	now time.Time,
) ([]*entity.Report, error) {
	query := `
		SELECT ` + reportColumns + `
		FROM reports
		WHERE expires_at <= $1::TIMESTAMPTZ + INTERVAL '3 days'
		  AND expires_at > $1::TIMESTAMPTZ
		  AND expiration_reminder_sent_at IS NULL
		ORDER BY expires_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, now)
	if err != nil {
		return nil, fmt.Errorf("list expiring reports: %w", err)
	}
	defer rows.Close()

	reports := make([]*entity.Report, 0)

	for rows.Next() {
		report, err := scanReport(rows)
		if err != nil {
			return nil, fmt.Errorf("scan expiring report: %w", err)
		}

		reports = append(reports, report)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate expiring reports: %w", err)
	}

	return reports, nil
}

func (r *Repository) Delete(
	ctx context.Context,
	organizationID string,
	reportID string,
) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM reports
		WHERE organization_id = $1
		  AND id = $2
	`,
		organizationID,
		reportID,
	)
	if err != nil {
		return fmt.Errorf("delete report: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get deleted report rows affected: %w", err)
	}

	if affected == 0 {
		return repo.ErrNotFound
	}

	return nil
}

func (r *Repository) MarkCleanupCompleted(
	ctx context.Context,
	organizationID string,
	reportID string,
) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE reports
		SET cleanup_completed = TRUE
		WHERE organization_id = $1
		  AND id = $2
	`,
		organizationID,
		reportID,
	)
	if err != nil {
		return fmt.Errorf("mark report cleanup completed: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get updated report rows affected: %w", err)
	}

	if affected == 0 {
		return repo.ErrNotFound
	}

	return nil
}

func (r *Repository) MarkExpirationReminderSent(
    ctx context.Context,
    organizationID string,
    reportID string,
) error {
    result, err := r.db.ExecContext(ctx, `
        UPDATE reports
        SET expiration_reminder_sent_at = CURRENT_TIMESTAMP
        WHERE organization_id = $1
          AND id = $2
          AND expiration_reminder_sent_at IS NULL
    `, organizationID, reportID)
    if err != nil {
        return fmt.Errorf("mark expiration reminder sent: %w", err)
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

var _ repo.ReportRepository = (*Repository)(nil)