package request

type UpdateOrganization struct {
	Name         *string `form:"name" json:"name" binding:"omitempty,min=2,max=200"`
	AddressLine1 *string `form:"address_line_1" json:"address_line_1" binding:"omitempty,max=255"`
	AddressLine2 *string `form:"address_line_2" json:"address_line_2" binding:"omitempty,max=255"`
	City         *string `form:"city" json:"city" binding:"omitempty,max=100"`
	State        *string `form:"state" json:"state" binding:"omitempty,max=100"`
	PostalCode   *string `form:"postal_code" json:"postal_code" binding:"omitempty,max=20"`
	Country      *string `form:"country" json:"country" binding:"omitempty,max=100"`
}
	