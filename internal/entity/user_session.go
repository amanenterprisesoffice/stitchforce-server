package entity

import "time"

type UserSession struct {
    ID             string     `json:"id" db:"id"`
    UserID         string     `json:"user_id" db:"user_id"`
    OrganizationID string     `json:"organization_id" db:"organization_id"`
    DeviceName     string     `json:"device_name" db:"device_name"`
    IPAddress      string     `json:"ip_address" db:"ip_address"`
    UserAgent      string     `json:"user_agent" db:"user_agent"`
    ExpiresAt      time.Time  `json:"expires_at" db:"expires_at"`
    RevokedAt      *time.Time `json:"revoked_at,omitempty" db:"revoked_at"`
    CreatedAt      time.Time  `json:"created_at" db:"created_at"`
    LastUsedAt     time.Time  `json:"last_used_at" db:"last_used_at"`
}