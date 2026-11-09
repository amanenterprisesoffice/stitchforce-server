package output

import (
	"encoding/json"
	"stitchforce-server/internal/entity"
	"time"
)

type BundleDetail struct {
	ID             string              `json:"id"`
	OrganizationID string              `json:"organization_id"`
	BundleNumber   string              `json:"bundle_number"`
	StyleID        string              `json:"style_id"`
	SizeID         string              `json:"size_id"`
	Quantity       int                 `json:"quantity"`
	Attributes     json.RawMessage     `json:"attributes" db:"attributes"`
	Status         entity.BundleStatus `json:"status"`
	Remarks        string              `json:"remarks"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`

	Operations []BundleOperationProgress `json:"operations"`
}

type BundleOperationProgress struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	RatePerPiece      int64  `json:"rate_per_piece"`
	CompletedQuantity int    `json:"completed_quantity"`
	PendingQuantity   int    `json:"pending_quantity"`

	Worker *SlimWorker `json:"worker"`
}
