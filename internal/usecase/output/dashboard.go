package output

type OrganizationDashboard struct {
    Overview     DashboardOverview     `json:"overview"`
    Today        DashboardPeriodStats  `json:"today"`
    ThisWeek     DashboardWeeklyStats  `json:"this_week"`
    ThisMonth    DashboardPeriodStats  `json:"this_month"`
    BundleStatus DashboardBundleStatus `json:"bundle_status"`
}

type DashboardOverview struct {
    TotalWorkers           int   `json:"total_workers"`
    ActiveWorkers          int   `json:"active_workers"`
    InactiveWorkers        int   `json:"inactive_workers"`

    TotalStyles            int   `json:"total_styles"`
    TotalBundles           int   `json:"total_bundles"`
    TotalProductionEntries int   `json:"total_production_entries"`

    TotalEarned            int64 `json:"total_earned"`
    TotalPaid              int64 `json:"total_paid"`
    TotalPending           int64 `json:"total_pending"`
}

type DashboardPeriodStats struct {
    ProductionQuantity int   `json:"production_quantity"`
    ProductionEntries  int   `json:"production_entries"`

    PaymentCount       int   `json:"payment_count"`
    PaymentAmount      int64 `json:"payment_amount"`
}

type DashboardWeeklyStats struct {
    ProductionQuantity int   `json:"production_quantity"`
    ProductionEntries  int   `json:"production_entries"`

    PaymentCount       int   `json:"payment_count"`
    PaymentAmount      int64 `json:"payment_amount"`

    Daily []DashboardDailyStats `json:"daily"`
}

type DashboardDailyStats struct {
    Date string `json:"date"`

    ProductionQuantity int   `json:"production_quantity"`
    ProductionEntries  int   `json:"production_entries"`

    PaymentCount       int   `json:"payment_count"`
    PaymentAmount      int64 `json:"payment_amount"`
}

type DashboardBundleStatus struct {
    Created    int `json:"created"`
    Ready      int `json:"ready"`
    InProgress int `json:"in_progress"`
    Completed  int `json:"completed"`
    Cancelled  int `json:"cancelled"`
}