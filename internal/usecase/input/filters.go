package input

import (
	"stitchforce-server/internal/entity"
	"time"
)
type DashboardGroup string
const (
	DashboardGroupDaily       = "daily"
	DashboardGroupWeekly      = "weekly"
	DashboardGroupHalfMonthly = "half_monthly"
	DashboardGroupMonthly     = "monthly"
	DashboardGroupQuarterly   = "quarterly"
)

func (g DashboardGroup) IsValid() bool {
	switch g {
	case DashboardGroupDaily, DashboardGroupWeekly, DashboardGroupHalfMonthly, DashboardGroupMonthly, DashboardGroupQuarterly:
		return true
	default:
		return false
	}
}

type DashboardFilter struct {
	FromDate time.Time
	ToDate   time.Time
	GroupBy  DashboardGroup
}

type StyleFilter struct {
	Search string

	Limit  int
	Offset int
}

type BundleFilter struct {
	Search  string
	StyleID string
	SizeID  string
	Status  string

	CreatedFrom *time.Time
	CreatedTo   *time.Time

	Limit  int
	Offset int
}

type ProductionEntryFilter struct {
	Search           string
	BundleID         string
	WorkerID         string
	StyleID          string
	StyleOperationID string

	PaymentStatus entity.ProductionEntryPaymentStatus // all, payable, pending, partially_paid, paid

	FromDate *time.Time
	ToDate   *time.Time

	Limit  int
	Offset int
}

type GarmentWorkerFilter struct {
	Search string

	Status entity.WorkerStatus
	Limit  int
	Offset int
}

type StyleOperationFilter struct {
	Search string

	Limit  int
	Offset int
}

type WorkerPaymentFilter struct {
	WorkerID string
	Status   entity.PaymentStatus
	FromDate *time.Time
	ToDate   *time.Time
	Limit    int
	Offset   int
}

type WorkerPaymentBalanceFilter struct {
	Search string
	Limit  int
	Offset int
}

type WorkerPaymentSummaryFilter struct {
	WorkerID string
	Status   entity.PaymentStatus

	FromDate *time.Time
	ToDate   *time.Time
}
