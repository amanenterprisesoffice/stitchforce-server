
package request

type CreateStyle struct {
    Name        string `json:"name" validate:"required,min=2,max=200"`
    StyleNumber string `json:"style_number" validate:"required,min=1,max=100"`
    Description string `json:"description" validate:"omitempty,max=1000"`
}

type UpdateStyle struct {
    Name        string `json:"name" validate:"required,min=2,max=200"`
    StyleNumber string `json:"style_number" validate:"required,min=1,max=100"`
    Description string `json:"description" validate:"omitempty,max=1000"`
}
