package input

import "time"

type StyleFilter struct {
    Search string

    Limit  int
    Offset int
}

type BundleFilter struct {
    Search string
    StyleID string
    SizeID  string
    Status  string

    CreatedFrom *time.Time
    CreatedTo   *time.Time

    Limit  int
    Offset int
}

type ProductionEntryFilter struct {
    Search string
    BundleID        string
    WorkerID        string
    StyleID         string
    StyleOperationID string

    FromDate *time.Time
    ToDate   *time.Time

    Limit  int
    Offset int
}

type GarmentWorkerFilter struct {
    Status string
    Search string

    Limit  int
    Offset int
}

type StyleOperationFilter struct {
	Search string

	Limit  int
	Offset int
}