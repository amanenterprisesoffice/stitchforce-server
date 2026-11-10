package entity

import "time"

type License struct {
	ID             string     `json:"id" db:"id"`
	LicenseKeyHash string     `json:"-" db:"license_key_hash"`
	DurationDays   int        `json:"duration_days" db:"duration_days"`
	Status         string     `json:"status" db:"status"`
	OrganizationID *string    `json:"organization_id,omitempty" db:"organization_id"`
	ActivatedAt    *time.Time `json:"activated_at,omitempty" db:"activated_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

// GeneratedLicense is returned only when generating a new license.
// The original key is never stored in the database.
type GeneratedLicense struct {
	License    *License `json:"license"`
	LicenseKey string   `json:"license_key"`
}