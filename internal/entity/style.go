package entity

import "time"

type Style struct {
    ID             string    `json:"id" db:"id"`
    OrganizationID string    `json:"organization_id" db:"organization_id"`
    Name           string    `json:"name" db:"name"`
    StyleNumber    string    `json:"style_number" db:"style_number"`
    Description    string    `json:"description" db:"description"`
    CreatedAt      time.Time `json:"created_at" db:"created_at"`
    UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}