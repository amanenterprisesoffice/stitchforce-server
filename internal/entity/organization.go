
package entity

import "time"

type Organization struct {
    ID               string     `json:"id" db:"id"`
    Name             string     `json:"name" db:"name"`
    LicenseStatus    string     `json:"license_status" db:"license_status"`
    LicenseExpiresAt *time.Time `json:"license_expires_at" db:"license_expires_at"`
    CreatedAt        time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}
