package output

import "stitchforce-server/internal/usecase/input"

type OrganizationDashboard struct {
	Period DashboardPeriodStats `json:"period"`
	Trend  DashboardTrend       `json:"trend"`
}

type DashboardOverview struct {
	TotalWorkers    int `json:"total_workers"`
	ActiveWorkers   int `json:"active_workers"`
	InactiveWorkers int `json:"inactive_workers"`

	TotalStyles            int `json:"total_styles"`
	TotalBundles           int `json:"total_bundles"`
	TotalProductionEntries int `json:"total_production_entries"`

	TotalEarned  int64                 `json:"total_earned"`
	TotalPaid    int64                 `json:"total_paid"`
	TotalPending int64                 `json:"total_pending"`
	BundleStatus DashboardBundleStatus `json:"bundle_status"`
	Today        DashboardPeriodStats  `json:"today"`
}

// Aggregated statistics for the selected date range.
type DashboardPeriodStats struct {
	ProductionQuantity int   `json:"production_quantity"`
	ProductionEntries  int   `json:"production_entries"`
	ProductionAmount   int64 `json:"production_amount"`

	PaymentCount  int   `json:"payment_count"`
	PaymentAmount int64 `json:"payment_amount"`
}

type DashboardTrend struct {
	FromDate string               `json:"from_date"`
	ToDate   string               `json:"to_date"`
	GroupBy  input.DashboardGroup `json:"group_by"`

	Buckets []DashboardTrendBucket `json:"buckets"`
}

type DashboardTrendBucket struct {
	Label    string `json:"label"`
	FromDate string `json:"from_date"`
	ToDate   string `json:"to_date"`

	ProductionQuantity int   `json:"production_quantity"`
	ProductionEntries  int   `json:"production_entries"`
	ProductionAmount   int64 `json:"production_amount"`

	PaymentCount  int   `json:"payment_count"`
	PaymentAmount int64 `json:"payment_amount"`
}

type DashboardBundleStatus struct {
	Created    int `json:"created"`
	Ready      int `json:"ready"`
	InProgress int `json:"in_progress"`
	Completed  int `json:"completed"`
}
