package request

type Register struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type Login struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}


type UpdateUser struct {
	Name  *string `form:"name" json:"name" binding:"omitempty,min=2,max=100"`
	Email *string `form:"email" json:"email" binding:"omitempty,email"`
}