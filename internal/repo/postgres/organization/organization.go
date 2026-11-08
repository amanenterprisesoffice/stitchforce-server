package organization

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"stitchforce-server/internal/entity"
	"stitchforce-server/internal/repo"
	"stitchforce-server/internal/usecase/output"
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

func (r *Repository) GetDashboard(
	ctx context.Context,
	organizationID string,
) (*output.OrganizationDashboard, error) {
	dashboard := &output.OrganizationDashboard{
		ThisWeek: output.DashboardWeeklyStats{
			Daily: make([]output.DashboardDailyStats, 0, 7),
		},
	}

	location, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		return nil, fmt.Errorf("load dashboard timezone: %w", err)
	}

	now := time.Now().In(location)

	todayStart := time.Date(
		now.Year(), now.Month(), now.Day(),
		0, 0, 0, 0, location,
	)
	tomorrow := todayStart.AddDate(0, 0, 1)

	monthStart := time.Date(
		now.Year(), now.Month(), 1,
		0, 0, 0, 0, location,
	)
	nextMonth := monthStart.AddDate(0, 1, 0)

	// Monday is the first day of the week.
	daysSinceMonday := (int(todayStart.Weekday()) + 6) % 7

	weekStart := todayStart.AddDate(0, 0, -daysSinceMonday)
	nextWeek := weekStart.AddDate(0, 0, 7)

	// --------------------------------------------------
	// 1. LIFETIME OVERVIEW
	// --------------------------------------------------

	overviewQuery := `
        SELECT
            (
                SELECT COUNT(*)
                FROM garment_workers
                WHERE organization_id = $1
                  AND deleted_at IS NULL
            ),

            (
                SELECT COUNT(*)
                FROM garment_workers
                WHERE organization_id = $1
                  AND deleted_at IS NULL
                  AND status = 'active'
            ),

            (
                SELECT COUNT(*)
                FROM garment_workers
                WHERE organization_id = $1
                  AND deleted_at IS NULL
                  AND status = 'inactive'
            ),

            (
                SELECT COUNT(*)
                FROM styles
                WHERE organization_id = $1
            ),

            (
                SELECT COUNT(*)
                FROM bundles
                WHERE organization_id = $1
                  AND deleted_at IS NULL
            ),

            (
                SELECT COUNT(*)
                FROM production_entries
                WHERE organization_id = $1
                  AND deleted_at IS NULL
            ),

            (
                SELECT COALESCE(SUM(total_amount), 0)
                FROM production_entries
                WHERE organization_id = $1
                  AND deleted_at IS NULL
            ),

            (
                SELECT COALESCE(SUM(amount), 0)
                FROM worker_payments
                WHERE organization_id = $1
                  AND status = 'completed'
            )
    `

	err = r.db.QueryRowContext(
		ctx,
		overviewQuery,
		organizationID,
	).Scan(
		&dashboard.Overview.TotalWorkers,
		&dashboard.Overview.ActiveWorkers,
		&dashboard.Overview.InactiveWorkers,
		&dashboard.Overview.TotalStyles,
		&dashboard.Overview.TotalBundles,
		&dashboard.Overview.TotalProductionEntries,
		&dashboard.Overview.TotalEarned,
		&dashboard.Overview.TotalPaid,
	)
	if err != nil {
		return nil, fmt.Errorf("get dashboard overview: %w", err)
	}

	dashboard.Overview.TotalPending =
		dashboard.Overview.TotalEarned -
			dashboard.Overview.TotalPaid

	// --------------------------------------------------
	// 2. BUNDLE STATUS
	// --------------------------------------------------

	bundleQuery := `
        SELECT
            status,
            COUNT(*)
        FROM bundles
        WHERE organization_id = $1
          AND deleted_at IS NULL
        GROUP BY status
    `

	rows, err := r.db.QueryContext(
		ctx,
		bundleQuery,
		organizationID,
	)
	if err != nil {
		return nil, fmt.Errorf("get dashboard bundle statuses: %w", err)
	}

	for rows.Next() {
		var status string
		var count int

		if err := rows.Scan(&status, &count); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan bundle status: %w", err)
		}

		switch status {
		case "created":
			dashboard.BundleStatus.Created = count
		case "ready":
			dashboard.BundleStatus.Ready = count
		case "in_progress":
			dashboard.BundleStatus.InProgress = count
		case "completed":
			dashboard.BundleStatus.Completed = count
		case "cancelled":
			dashboard.BundleStatus.Cancelled = count
		}
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("iterate bundle statuses: %w", err)
	}

	rows.Close()

	// --------------------------------------------------
	// 3. TODAY AND THIS MONTH PRODUCTION
	// --------------------------------------------------

	productionPeriodQuery := `
        SELECT
            COUNT(*) FILTER (
                WHERE production_date >= $2
                  AND production_date < $3
            ),

            COALESCE(SUM(quantity) FILTER (
                WHERE production_date >= $2
                  AND production_date < $3
            ), 0),

            COUNT(*) FILTER (
                WHERE production_date >= $4
                  AND production_date < $5
            ),

            COALESCE(SUM(quantity) FILTER (
                WHERE production_date >= $4
                  AND production_date < $5
            ), 0)

        FROM production_entries
        WHERE organization_id = $1
          AND deleted_at IS NULL
          AND (
              (production_date >= $2 AND production_date < $3)
              OR
              (production_date >= $4 AND production_date < $5)
          )
    `

	err = r.db.QueryRowContext(
		ctx,
		productionPeriodQuery,
		organizationID,
		todayStart,
		tomorrow,
		monthStart,
		nextMonth,
	).Scan(
		&dashboard.Today.ProductionEntries,
		&dashboard.Today.ProductionQuantity,
		&dashboard.ThisMonth.ProductionEntries,
		&dashboard.ThisMonth.ProductionQuantity,
	)
	if err != nil {
		return nil, fmt.Errorf("get dashboard production periods: %w", err)
	}

	// --------------------------------------------------
	// 4. TODAY AND THIS MONTH PAYMENTS
	// --------------------------------------------------

	paymentPeriodQuery := `
        SELECT
            COUNT(*) FILTER (
                WHERE payment_date >= $2
                  AND payment_date < $3
            ),

            COALESCE(SUM(amount) FILTER (
                WHERE payment_date >= $2
                  AND payment_date < $3
            ), 0),

            COUNT(*) FILTER (
                WHERE payment_date >= $4
                  AND payment_date < $5
            ),

            COALESCE(SUM(amount) FILTER (
                WHERE payment_date >= $4
                  AND payment_date < $5
            ), 0)

        FROM worker_payments
        WHERE organization_id = $1
          AND status = 'completed'
          AND (
              (payment_date >= $2 AND payment_date < $3)
              OR
              (payment_date >= $4 AND payment_date < $5)
          )
    `

	err = r.db.QueryRowContext(
		ctx,
		paymentPeriodQuery,
		organizationID,
		todayStart,
		tomorrow,
		monthStart,
		nextMonth,
	).Scan(
		&dashboard.Today.PaymentCount,
		&dashboard.Today.PaymentAmount,
		&dashboard.ThisMonth.PaymentCount,
		&dashboard.ThisMonth.PaymentAmount,
	)
	if err != nil {
		return nil, fmt.Errorf("get dashboard payment periods: %w", err)
	}

	// --------------------------------------------------
	// 5. INITIALIZE ALL SEVEN DAYS
	// --------------------------------------------------

	for i := 0; i < 7; i++ {
		date := weekStart.AddDate(0, 0, i)

		dashboard.ThisWeek.Daily = append(
			dashboard.ThisWeek.Daily,
			output.DashboardDailyStats{
				Date: date.Format("2006-01-02"),
			},
		)
	}

	// --------------------------------------------------
	// 6. WEEKLY PRODUCTION
	// --------------------------------------------------

	weeklyProductionQuery := `
        SELECT
            TO_CHAR(
                production_date AT TIME ZONE 'Asia/Kolkata',
                'YYYY-MM-DD'
            ),
            COUNT(*),
            COALESCE(SUM(quantity), 0)

        FROM production_entries
        WHERE organization_id = $1
          AND deleted_at IS NULL
          AND production_date >= $2
          AND production_date < $3

        GROUP BY 1
        ORDER BY 1
    `

	productionRows, err := r.db.QueryContext(
		ctx,
		weeklyProductionQuery,
		organizationID,
		weekStart,
		nextWeek,
	)
	if err != nil {
		return nil, fmt.Errorf("get weekly production: %w", err)
	}

	for productionRows.Next() {
		var date string
		var count int
		var quantity int

		if err := productionRows.Scan(
			&date,
			&count,
			&quantity,
		); err != nil {
			productionRows.Close()
			return nil, fmt.Errorf("scan weekly production: %w", err)
		}

		index := -1

		for i := range dashboard.ThisWeek.Daily {
			if dashboard.ThisWeek.Daily[i].Date == date {
				index = i
				break
			}
		}

		if index >= 0 {
			dashboard.ThisWeek.Daily[index].ProductionEntries = count
			dashboard.ThisWeek.Daily[index].ProductionQuantity = quantity
		}

		dashboard.ThisWeek.ProductionEntries += count
		dashboard.ThisWeek.ProductionQuantity += quantity
	}

	if err := productionRows.Err(); err != nil {
		productionRows.Close()
		return nil, fmt.Errorf("iterate weekly production: %w", err)
	}

	productionRows.Close()

	// --------------------------------------------------
	// 7. WEEKLY PAYMENTS
	// --------------------------------------------------

	weeklyPaymentQuery := `
        SELECT
            TO_CHAR(
                payment_date AT TIME ZONE 'Asia/Kolkata',
                'YYYY-MM-DD'
            ),
            COUNT(*),
            COALESCE(SUM(amount), 0)

        FROM worker_payments
        WHERE organization_id = $1
          AND status = 'completed'
          AND payment_date >= $2
          AND payment_date < $3

        GROUP BY 1
        ORDER BY 1
    `

	paymentRows, err := r.db.QueryContext(
		ctx,
		weeklyPaymentQuery,
		organizationID,
		weekStart,
		nextWeek,
	)
	if err != nil {
		return nil, fmt.Errorf("get weekly payments: %w", err)
	}

	for paymentRows.Next() {
		var date string
		var count int
		var amount int64

		if err := paymentRows.Scan(
			&date,
			&count,
			&amount,
		); err != nil {
			paymentRows.Close()
			return nil, fmt.Errorf("scan weekly payments: %w", err)
		}

		for i := range dashboard.ThisWeek.Daily {
			if dashboard.ThisWeek.Daily[i].Date == date {
				dashboard.ThisWeek.Daily[i].PaymentCount = count
				dashboard.ThisWeek.Daily[i].PaymentAmount = amount
				break
			}
		}

		dashboard.ThisWeek.PaymentCount += count
		dashboard.ThisWeek.PaymentAmount += amount
	}

	if err := paymentRows.Err(); err != nil {
		paymentRows.Close()
		return nil, fmt.Errorf("iterate weekly payments: %w", err)
	}

	paymentRows.Close()

	return dashboard, nil
}
