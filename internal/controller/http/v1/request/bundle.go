package request

import (
	"encoding/json"
	"stitchforce-server/internal/entity"
)

type CreateBundle struct {
	BundleNumber string              `json:"bundle_number" validate:"required,min=1,max=100"`
	StyleID      string              `json:"style_id" validate:"required"`
	SizeID       string              `json:"size_id" validate:"required"`
	Quantity     int                 `json:"quantity" validate:"required,gt=0"`
	Attributes   json.RawMessage     `json:"attributes" validate:"omitempty"`
	Status       entity.BundleStatus `json:"status" validate:"omitempty,oneof=created ready"`
	Remarks      string              `json:"remarks" validate:"omitempty,max=1000"`
}

type UpdateBundle struct {
    BundleNumber *string              `json:"bundle_number" validate:"omitempty,min=1,max=100"`
    SizeID       *string              `json:"size_id" validate:"omitempty"`
    Quantity     *int                 `json:"quantity" validate:"omitempty,gt=0"`
    Attributes   *json.RawMessage     `json:"attributes" validate:"omitempty"`
    Status       *entity.BundleStatus `json:"status" validate:"omitempty,oneof=created ready"`
    Remarks      *string              `json:"remarks" validate:"omitempty,max=1000"`
}