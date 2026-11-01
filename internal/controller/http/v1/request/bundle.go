
package request

import "encoding/json"

type CreateBundle struct {
    BundleNumber string          `form:"bundle_number" json:"bundle_number" binding:"required,min=1,max=100"`
    StyleID      string          `form:"style_id" json:"style_id" binding:"required"`
    SizeID       string          `form:"size_id" json:"size_id" binding:"required"`
    Quantity     int             `form:"quantity" json:"quantity" binding:"required,gt=0"`
    Attributes   json.RawMessage `form:"attributes" json:"attributes"`
    Status       string          `form:"status" json:"status" binding:"omitempty,oneof=created ready"`
    Remarks      string          `form:"remarks" json:"remarks" binding:"omitempty,max=1000"`
}

type UpdateBundle struct {
    BundleNumber string          `form:"bundle_number" json:"bundle_number" binding:"required,min=1,max=100"`
    StyleID      string          `form:"style_id" json:"style_id" binding:"required"`
    SizeID       string          `form:"size_id" json:"size_id" binding:"required"`
    Quantity     int             `form:"quantity" json:"quantity" binding:"required,gt=0"`
    Attributes   json.RawMessage `form:"attributes" json:"attributes"`
    Status       string          `form:"status" json:"status" binding:"omitempty,oneof=created ready"`
    Remarks      string          `form:"remarks" json:"remarks" binding:"omitempty,max=1000"`
}
