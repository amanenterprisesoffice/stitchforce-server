package input

import (
    "time"

    "stitchforce-server/internal/entity"
)

type CreateWorkerPaymentInput struct {
    WorkerID      string              `json:"worker_id"`
    FromDate      time.Time           `json:"from_date"`
    ToDate        time.Time           `json:"to_date"`
    PaymentDate   time.Time           `json:"payment_date"`
    PaymentMethod entity.PaymentMethod `json:"payment_method"`
    Reference     string              `json:"reference"`
    Remarks       string              `json:"remarks"`
}