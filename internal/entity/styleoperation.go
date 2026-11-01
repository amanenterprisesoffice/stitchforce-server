package entity

import "time"

type StyleOperation struct {
    ID             string    `json:"id" db:"id"`
    OrganizationID string    `json:"organization_id" db:"organization_id"`
    StyleID        string    `json:"style_id" db:"style_id"`
    Name           string    `json:"name" db:"name"`
    RatePerPiece   int64     `json:"rate_per_piece" db:"rate_per_piece"`
    CreatedAt      time.Time `json:"created_at" db:"created_at"`
    UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}