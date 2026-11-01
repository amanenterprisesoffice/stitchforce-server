package entity

import (
    "encoding/json"
    "time"
)

type Bundle struct {
    ID             string `json:"id" db:"id"`
    OrganizationID string `json:"organization_id" db:"organization_id"`

    BundleNumber   string `json:"bundle_number" db:"bundle_number"`

    StyleID        string `json:"style_id" db:"style_id"`
    SizeID         string `json:"size_id" db:"size_id"`

    Quantity       int    `json:"quantity" db:"quantity"`

    // Flexible attributes: shade, color, fabric lot, etc.
    Attributes     json.RawMessage `json:"attributes" db:"attributes"`

    Status         string `json:"status" db:"status"`
    Remarks        string `json:"remarks" db:"remarks"`

    CreatedAt      time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
    DeletedAt      *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}