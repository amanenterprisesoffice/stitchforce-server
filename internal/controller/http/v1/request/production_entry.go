
package request

import "time"

type CreateProductionEntry struct {
    BundleID         string    `json:"bundle_id" validate:"required"`
    WorkerID         string    `json:"worker_id" validate:"required"`
    StyleOperationID string    `json:"style_operation_id" validate:"required"`
    ProductionDate   time.Time `json:"production_date" validate:"required"`
    Remarks          string    `json:"remarks" validate:"omitempty,max=1000"`
}

type UpdateProductionEntry struct {
    BundleID       string    `json:"bundle_id" validate:"required"`
    WorkerID       string    `json:"worker_id" validate:"required"`
    Quantity       int       `json:"quantity" validate:"required,gt=0"`
    ProductionDate time.Time `json:"production_date" validate:"required"`
    Remarks        string    `json:"remarks" validate:"omitempty,max=1000"`
}
