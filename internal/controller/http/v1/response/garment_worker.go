package response

import "time"

type GarmentWorkerResponse struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Phone    string  `json:"phone"`
	Gender   string  `json:"gender"`
	PhotoURL *string `json:"photo_url,omitempty"`

	Status string `json:"status" db:"status"` // active, inactive
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}
