package output

import "time"

type Organization struct {
	ID               string     `json:"id"`
	Name             string     `json:"name"`
	AddressLine1     string     `json:"address_line_1"`
	AddressLine2     string     `json:"address_line_2"`
	City             string     `json:"city"`
	State            string     `json:"state"`
	PostalCode       string     `json:"postal_code"`
	Country          string     `json:"country"`
	PhotoURL         *string    `json:"photo_url,omitempty"`
	LicenseStatus    string     `json:"license_status"`
	LicenseExpiresAt *time.Time `json:"license_expires_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}