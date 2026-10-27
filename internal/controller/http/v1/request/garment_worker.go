package request

type CreateGarmentWorker struct {
	Name   string `json:"name" validate:"required,min=2,max=200"`
	Phone  string `json:"phone" validate:"required,min=10,max=15"`
	Gender string `json:"gender" validate:"required,oneof=male female other"`
}

type UpdateGarmentWorker struct {
	Name   string `json:"name" validate:"required,min=2,max=200"`
	Phone  string `json:"phone" validate:"required,min=10,max=15"`
	Gender string `json:"gender" validate:"required,oneof=male female other"`
}