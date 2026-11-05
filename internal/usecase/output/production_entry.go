package output

import "time"

type ProductionEntry struct {
	ID               string `json:"id"`
	OrganizationID   string `json:"organization_id"`
	BundleID         string `json:"bundle_id"`
	WorkerID         string `json:"worker_id"`
	StyleOperationID string `json:"style_operation_id"`

	Quantity     int   `json:"quantity"`
	RatePerPiece int64 `json:"rate_per_piece"`
	TotalAmount  int64 `json:"total_amount"`

	ProductionDate time.Time `json:"production_date"`
	Remarks        string    `json:"remarks"`

	Worker    ProductionEntryWorker    `json:"worker"`
	Bundle    ProductionEntryBundle    `json:"bundle"`
	Style     ProductionEntryStyle     `json:"style"`
	Operation ProductionEntryOperation `json:"operation"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProductionEntryWorker struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	PhotoURL string `json:"photo_url,omitempty"`
}

type ProductionEntryBundle struct {
	ID           string `json:"id"`
	BundleNumber string `json:"bundle_number"`
	Quantity     int    `json:"quantity"`
	Status       string `json:"status"`
}

type ProductionEntryStyle struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	StyleNumber string `json:"style_number"`
}

type ProductionEntryOperation struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	RatePerPiece int64  `json:"rate_per_piece"`
}
