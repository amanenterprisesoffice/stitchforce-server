package output

import (
	"stitchforce-server/internal/entity"
	"time"
)

type User struct {
	ID              string          `json:"id"`
	OrganizationID  string          `json:"organization_id"`
	Name            string          `json:"name"`
	Email           string          `json:"email"`
	Role            entity.UserRole `json:"role"`
	PhotoURL        *string         `json:"photo_url,omitempty"`
	EmailVerifiedAt *time.Time      `json:"email_verified_at,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}
