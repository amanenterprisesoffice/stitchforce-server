package output

import (
	"stitchforce-server/internal/entity"
	"time"
)

type GarmentWorker struct {
	ID             string              `json:"id"`
	OrganizationID string              `json:"organization_id"`
	Name           string              `json:"name"`
	Phone          string              `json:"phone"`
	Gender         string              `json:"gender"`
	PhotoURL       *string             `json:"photo_url,omitempty"`
	Status         entity.WorkerStatus `json:"status"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
	DeletedAt      *time.Time          `json:"deleted_at,omitempty"`
}

type SlimWorker struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	PhotoKey string `json:"-" db:"photo_key"`

	PhotoURL string `json:"photo_url,omitempty"`
}
