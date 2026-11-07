package request

import "stitchforce-server/internal/entity"

type CreateWorkerPayment struct {
	WorkerID      string               `json:"worker_id" binding:"required"`
	FromDate      string               `json:"from_date" binding:"required"`
	ToDate        string               `json:"to_date" binding:"required"`
	PaymentDate   string               `json:"payment_date" binding:"required"`
	PaymentMethod entity.PaymentMethod `json:"payment_method" binding:"required,oneof=cash bank upi cheque"`
	Reference     string               `json:"reference" binding:"omitempty,max=200"`
	Remarks       string               `json:"remarks" binding:"omitempty,max=1000"`
}
