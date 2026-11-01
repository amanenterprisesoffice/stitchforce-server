
package request

type CreateStyleOperation struct {
    Name         string `json:"name" validate:"required,min=2,max=200"`
    RatePerPiece int64  `json:"rate_per_piece" validate:"gte=0"`
}

type UpdateStyleOperation struct {
    Name         string `json:"name" validate:"required,min=2,max=200"`
    RatePerPiece int64  `json:"rate_per_piece" validate:"gte=0"`
}
