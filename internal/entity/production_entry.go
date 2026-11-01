package entity

import "time"

type ProductionEntry struct {
	ID             string `json:"id" db:"id"`
	OrganizationID string `json:"organization_id" db:"organization_id"`

	BundleID        string `json:"bundle_id" db:"bundle_id"`
	WorkerID        string `json:"worker_id" db:"worker_id"`
	StyleOperationID string `json:"style_operation_id" db:"style_operation_id"`

	Quantity     int   `json:"quantity" db:"quantity"`
	RatePerPiece int64 `json:"rate_per_piece" db:"rate_per_piece"`

	ProductionDate time.Time `json:"production_date" db:"production_date"`
	Remarks        string    `json:"remarks" db:"remarks"`

	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}