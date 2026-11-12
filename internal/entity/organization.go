
package entity

import "time"

type Organization struct {
	ID               string     `json:"id" db:"id"`
	Name             string     `json:"name" db:"name"`
	AddressLine1     string     `json:"address_line_1" db:"address_line_1"`
	AddressLine2     string     `json:"address_line_2" db:"address_line_2"`
	City             string     `json:"city" db:"city"`
	State            string     `json:"state" db:"state"`
	PostalCode        string     `json:"postal_code" db:"postal_code"`
	Country           string     `json:"country" db:"country"`
	PhotoKey          string     `json:"-" db:"photo_key"`
	LicenseStatus     string     `json:"license_status" db:"license_status"`
	LicenseExpiresAt  *time.Time `json:"license_expires_at" db:"license_expires_at"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
}