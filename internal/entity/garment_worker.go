
package entity

import "time"

type GarmentWorker struct {
	ID             string     `json:"id" db:"id"`
	OrganizationID string     `json:"organization_id" db:"organization_id"`

	Name           string     `json:"name" db:"name"`
	Phone          string     `json:"phone" db:"phone"`
	Gender         string     `json:"gender" db:"gender"`
	PhotoKey       string     `json:"photo_key" db:"photo_key"`

	Status         WorkerStatus     `json:"status" db:"status"` // active, inactive

	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

type WorkerStatus string

const (
	WorkerStatusActive   WorkerStatus = "active"
	WorkerStatusInactive WorkerStatus = "inactive"
)
