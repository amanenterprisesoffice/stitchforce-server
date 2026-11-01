
package entity

import "time"

type ProductionEntry struct {
	ID             string `json:"id" db:"id"`

	OrganizationID string `json:"organization_id" db:"organization_id"`

	// Production references
	WorkerID       string `json:"worker_id" db:"worker_id"`
	StyleID        string `json:"style_id" db:"style_id"`
	BundleID       string `json:"bundle_id" db:"bundle_id"`
	OperationID    string `json:"operation_id" db:"operation_id"`

	// Production details
	CompletedPieces int64 `json:"completed_pieces" db:"completed_pieces"`

	// Historical rate snapshot.
	// Stored in paise to avoid floating-point calculations.
	RatePerPiece int64 `json:"rate_per_piece" db:"rate_per_piece"`

	// CompletedPieces * RatePerPiece
	TotalEarnings int64 `json:"total_earnings" db:"total_earnings"`

	ProductionDate time.Time `json:"production_date" db:"production_date"`

	Notes string `json:"notes" db:"notes"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
