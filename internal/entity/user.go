package entity

import "time"

type User struct {
	ID             string   `json:"id" db:"id" form:"id"`
	OrganizationID string   `json:"organization_id" db:"organization_id" form:"organization_id"`
	Name           string   `json:"name" db:"name" form:"name"`
	Email          string   `json:"email" db:"email" form:"email"`
	Role           UserRole `json:"role" db:"role" form:"role"`
	PhotoKey       string   `json:"-" db:"photo_key"`

	// Never expose password hashes in API responses.
	PasswordHash string `json:"-" db:"password_hash" form:"-"`

	EmailVerifiedAt              *time.Time `json:"email_verified_at,omitempty" db:"email_verified_at"`
	EmailVerificationTokenHash   string     `json:"-" db:"email_verification_token_hash"`
	EmailVerificationTokenExpiry *time.Time `json:"-" db:"email_verification_expires_at"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type UserRole string

const (
	UserRoleOrgAdmin UserRole = "org_admin"
	UserRoleAdmin    UserRole = "admin"
	UserRoleStaff    UserRole = "staff"
)

func (r UserRole) IsValid() bool {
	switch r {
	case UserRoleOrgAdmin, UserRoleAdmin, UserRoleStaff:
		return true
	default:
		return false
	}
}
