package input

import (
	"encoding/json"
	"stitchforce-server/internal/entity"
)
type UpdateBundleInput struct {
	BundleNumber *string
	SizeID       *string
	Quantity     *int
	Attributes   *json.RawMessage
	Status       *entity.BundleStatus
	Remarks      *string
}
