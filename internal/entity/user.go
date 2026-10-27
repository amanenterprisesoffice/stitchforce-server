package entity

type User struct {
	ID             string `json:"id" db:"id" form:"id"`
	OrganizationID string `json:"organization_id" db:"organization_id" form:"organization_id"`
	Name           string `json:"name" db:"name" form:"name"`
	Email          string `json:"email" db:"email" form:"email"`
	Role           string `json:"role" db:"role" form:"role"`

	// Never expose password hashes in API responses.
	PasswordHash string `json:"-" db:"password_hash" form:"-"`
}