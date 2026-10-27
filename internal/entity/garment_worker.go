package entity

type GarmentWorker struct {
	ID             string `json:"id" db:"id" form:"id"`
	OrganizationID string `json:"organization_id" db:"organization_id" form:"organization_id"`
	Name           string `json:"name" db:"name" form:"name"`
	Phone          string `json:"phone" db:"phone" form:"phone"`
	Gender         string `json:"gender" db:"gender" form:"gender"`
}