package organization

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"stitchforce-server/internal/entity"
	"stitchforce-server/internal/repo"
	"stitchforce-server/internal/usecase/input"
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

func (r *Repository) GetDashboardOverview(
	ctx context.Context,
	organizationID string,
) (*output.DashboardOverview, error) {

	overview := &output.DashboardOverview{}

	query := `
		SELECT
			-- Lifetime workers
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

			-- Lifetime styles and bundles
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

			-- Lifetime production entries
			(
				SELECT COUNT(*)
				FROM production_entries
				WHERE organization_id = $1
				  AND deleted_at IS NULL
			),

			-- Lifetime earned
			(
				SELECT COALESCE(SUM(total_amount), 0)
				FROM production_entries
				WHERE organization_id = $1
				  AND deleted_at IS NULL
			),

			-- Lifetime paid
			(
				SELECT COALESCE(SUM(amount), 0)
				FROM worker_payments
				WHERE organization_id = $1
				  AND status = 'completed'
			),

			-- Today's production quantity
			(
				SELECT COALESCE(SUM(quantity), 0)
				FROM production_entries
				WHERE organization_id = $1
				  AND deleted_at IS NULL
				  AND production_date::date = CURRENT_DATE
			),

			-- Today's production entries
			(
				SELECT COUNT(*)
				FROM production_entries
				WHERE organization_id = $1
				  AND deleted_at IS NULL
				  AND production_date::date = CURRENT_DATE
			),

			-- Today's production amount
			(
				SELECT COALESCE(SUM(total_amount), 0)
				FROM production_entries
				WHERE organization_id = $1
				  AND deleted_at IS NULL
				  AND production_date::date = CURRENT_DATE
			),

			-- Today's payment count
			(
				SELECT COUNT(*)
				FROM worker_payments
				WHERE organization_id = $1
				  AND status = 'completed'
				  AND payment_date::date = CURRENT_DATE
			),

			-- Today's payment amount
			(
				SELECT COALESCE(SUM(amount), 0)
				FROM worker_payments
				WHERE organization_id = $1
				  AND status = 'completed'
				  AND payment_date::date = CURRENT_DATE
			)
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		organizationID,
	).Scan(
		&overview.TotalWorkers,
		&overview.ActiveWorkers,
		&overview.InactiveWorkers,
		&overview.TotalStyles,
		&overview.TotalBundles,
		&overview.TotalProductionEntries,
		&overview.TotalEarned,
		&overview.TotalPaid,
		&overview.Today.ProductionQuantity,
		&overview.Today.ProductionEntries,
		&overview.Today.ProductionAmount,
		&overview.Today.PaymentCount,
		&overview.Today.PaymentAmount,
	)

	if err != nil {
		return nil, fmt.Errorf("get dashboard overview: %w", err)
	}

	overview.TotalPending = overview.TotalEarned - overview.TotalPaid

	// BUNDLE STATUS
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
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int

		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("scan bundle status: %w", err)
		}

		switch entity.BundleStatus(strings.TrimSpace(status)) {
		case entity.BundleStatusCreated:
			overview.BundleStatus.Created = count
		case entity.BundleStatusReady:
			overview.BundleStatus.Ready = count
		case entity.BundleStatusInProgress:
			overview.BundleStatus.InProgress = count
		case entity.BundleStatusCompleted:
			overview.BundleStatus.Completed = count
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bundle statuses: %w", err)
	}

	return overview, nil
}

func (r *Repository) GetDashboardTrend(
	ctx context.Context,
	organizationID string,
	filter input.DashboardFilter,
) (*output.OrganizationDashboard, error) {

	location, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		return nil, fmt.Errorf("load dashboard timezone: %w", err)
	}

	fromDate := time.Date(
		filter.FromDate.Year(),
		filter.FromDate.Month(),
		filter.FromDate.Day(),
		0, 0, 0, 0,
		location,
	)

	toDate := time.Date(
		filter.ToDate.Year(),
		filter.ToDate.Month(),
		filter.ToDate.Day(),
		0, 0, 0, 0,
		location,
	)

	if toDate.Before(fromDate) {
		return nil, fmt.Errorf("to_date cannot be before from_date")
	}

	switch filter.GroupBy {
	case input.DashboardGroupDaily,
		input.DashboardGroupWeekly,
		input.DashboardGroupHalfMonthly,
		input.DashboardGroupMonthly,
		input.DashboardGroupQuarterly:
	default:
		return nil, fmt.Errorf("invalid dashboard group_by")
	}

	toDateExclusive := toDate.AddDate(0, 0, 1)

	dashboard := &output.OrganizationDashboard{
		Trend: output.DashboardTrend{
			FromDate: fromDate.Format("2006-01-02"),
			ToDate:   toDate.Format("2006-01-02"),
			GroupBy:  filter.GroupBy,
			Buckets:  make([]output.DashboardTrendBucket, 0),
		},
	}

	// 1. INITIALIZE TREND BUCKETS
	bucketIndexes := make(map[string]int)

	for current := fromDate; !current.After(toDate); {
		bucketStart := dashboardBucketStart(current, filter.GroupBy)
		bucketEnd := dashboardBucketEnd(bucketStart, filter.GroupBy)

		// Clip bucket dates to the selected date range.
		actualStart := bucketStart
		if actualStart.Before(fromDate) {
			actualStart = fromDate
		}

		actualEnd := bucketEnd
		if actualEnd.After(toDate) {
			actualEnd = toDate
		}

		key := bucketStart.Format("2006-01-02")

		index := len(dashboard.Trend.Buckets)

		bucketIndexes[key] = index

		dashboard.Trend.Buckets = append(
			dashboard.Trend.Buckets,
			output.DashboardTrendBucket{
				Label:    dashboardBucketLabel(bucketStart, bucketEnd, filter.GroupBy),
				FromDate: actualStart.Format("2006-01-02"),
				ToDate:   actualEnd.Format("2006-01-02"),
			},
		)

		current = bucketEnd.AddDate(0, 0, 1)
	}

	// 2. PRODUCTION STATISTICS
	productionQuery := `
		SELECT
			TO_CHAR(
				production_date AT TIME ZONE 'Asia/Kolkata',
				'YYYY-MM-DD'
			),
			COUNT(*),
			COALESCE(SUM(quantity), 0),
			COALESCE(SUM(total_amount), 0)

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
		productionQuery,
		organizationID,
		fromDate,
		toDateExclusive,
	)
	if err != nil {
		return nil, fmt.Errorf("get dashboard production: %w", err)
	}

	for productionRows.Next() {
		var (
			date     string
			count    int
			quantity int
			amount   int64
		)

		if err := productionRows.Scan(
			&date,
			&count,
			&quantity,
			&amount,
		); err != nil {
			productionRows.Close()
			return nil, fmt.Errorf("scan dashboard production: %w", err)
		}

		parsedDate, err := time.ParseInLocation(
			"2006-01-02",
			date,
			location,
		)
		if err != nil {
			productionRows.Close()
			return nil, fmt.Errorf("parse production date: %w", err)
		}

		bucketStart := dashboardBucketStart(parsedDate, filter.GroupBy)
		key := bucketStart.Format("2006-01-02")

		index, exists := bucketIndexes[key]
		if !exists {
			continue
		}

		bucket := &dashboard.Trend.Buckets[index]

		bucket.ProductionEntries += count
		bucket.ProductionQuantity += quantity
		bucket.ProductionAmount += amount
	}

	if err := productionRows.Err(); err != nil {
		productionRows.Close()
		return nil, fmt.Errorf("iterate dashboard production: %w", err)
	}

	productionRows.Close()

	// 3. PAYMENT STATISTICS

	paymentQuery := `
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
		paymentQuery,
		organizationID,
		fromDate,
		toDateExclusive,
	)
	if err != nil {
		return nil, fmt.Errorf("get dashboard payments: %w", err)
	}

	for paymentRows.Next() {
		var (
			date   string
			count  int
			amount int64
		)

		if err := paymentRows.Scan(
			&date,
			&count,
			&amount,
		); err != nil {
			paymentRows.Close()
			return nil, fmt.Errorf("scan dashboard payment: %w", err)
		}

		parsedDate, err := time.ParseInLocation(
			"2006-01-02",
			date,
			location,
		)
		if err != nil {
			paymentRows.Close()
			return nil, fmt.Errorf("parse payment date: %w", err)
		}

		bucketStart := dashboardBucketStart(parsedDate, filter.GroupBy)
		key := bucketStart.Format("2006-01-02")

		index, exists := bucketIndexes[key]
		if !exists {
			continue
		}

		bucket := &dashboard.Trend.Buckets[index]

		bucket.PaymentCount += count
		bucket.PaymentAmount += amount
	}

	if err := paymentRows.Err(); err != nil {
		paymentRows.Close()
		return nil, fmt.Errorf("iterate dashboard payments: %w", err)
	}

	paymentRows.Close()

	// 4. CALCULATE SELECTED PERIOD TOTALS
	for _, bucket := range dashboard.Trend.Buckets {
		dashboard.Period.ProductionQuantity += bucket.ProductionQuantity
		dashboard.Period.ProductionEntries += bucket.ProductionEntries
		dashboard.Period.ProductionAmount += bucket.ProductionAmount

		dashboard.Period.PaymentCount += bucket.PaymentCount
		dashboard.Period.PaymentAmount += bucket.PaymentAmount
	}

	return dashboard, nil
}

// Helpers
func dashboardBucketStart(date time.Time, groupBy input.DashboardGroup) time.Time {
	year, month, day := date.Date()
	location := date.Location()

	switch groupBy {
	case input.DashboardGroupDaily:
		return time.Date(year, month, day, 0, 0, 0, 0, location)

	case input.DashboardGroupWeekly:
		daysSinceMonday := (int(date.Weekday()) + 6) % 7

		return time.Date(
			year, month, day-daysSinceMonday,
			0, 0, 0, 0,
			location,
		)

	case input.DashboardGroupHalfMonthly:
		if day <= 15 {
			return time.Date(year, month, 1, 0, 0, 0, 0, location)
		}

		return time.Date(year, month, 16, 0, 0, 0, 0, location)

	case input.DashboardGroupMonthly:
		return time.Date(year, month, 1, 0, 0, 0, 0, location)

	case input.DashboardGroupQuarterly:
		quarterMonth := time.Month(((int(month)-1)/3)*3 + 1)

		return time.Date(
			year, quarterMonth, 1,
			0, 0, 0, 0,
			location,
		)
	}

	return date
}

func dashboardBucketEnd(start time.Time, groupBy input.DashboardGroup) time.Time {
	switch groupBy {
	case input.DashboardGroupDaily:
		return start

	case input.DashboardGroupWeekly:
		return start.AddDate(0, 0, 6)

	case input.DashboardGroupHalfMonthly:
		if start.Day() == 1 {
			return start.AddDate(0, 0, 14)
		}

		return time.Date(
			start.Year(),
			start.Month()+1,
			0,
			0, 0, 0, 0,
			start.Location(),
		)

	case input.DashboardGroupMonthly:
		return time.Date(
			start.Year(),
			start.Month()+1,
			0,
			0, 0, 0, 0,
			start.Location(),
		)

	case input.DashboardGroupQuarterly:
		return start.AddDate(0, 3, 0).AddDate(0, 0, -1)
	}

	return start
}

func dashboardBucketLabel(
	start time.Time,
	end time.Time,
	groupBy input.DashboardGroup,
) string {
	switch groupBy {
	case input.DashboardGroupDaily:
		return start.Format("Jan 2")

	case input.DashboardGroupWeekly, input.DashboardGroupHalfMonthly:
		return start.Format("Jan 2") + "–" + end.Format("Jan 2")

	case input.DashboardGroupMonthly:
		return start.Format("Jan 2006")

	case input.DashboardGroupQuarterly:
		quarter := (int(start.Month())-1)/3 + 1
		return fmt.Sprintf("Q%d %d", quarter, start.Year())
	}

	return start.Format("Jan 2")
}
