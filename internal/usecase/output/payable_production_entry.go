package output

import "time"

type PayableProductionEntry struct {
	ProductionEntryID string                   `json:"production_entry_id"`
	WorkerId          string                   `json:"worker_id"`
	Bundle            ProductionEntryBundle    `json:"bundle"`
	Style             ProductionEntryStyle     `json:"style"`
	Operation         ProductionEntryOperation `json:"operation"`

	ProductionDate time.Time `json:"production_date"`

	Quantity      int   `json:"quantity"`
	RatePerPiece  int64 `json:"rate_per_piece"`
	TotalAmount   int64 `json:"total_amount"`
	PaidAmount    int64 `json:"paid_amount"`
	PendingAmount int64 `json:"pending_amount"`
}

type WorkerPaymentSummary struct {
	TotalEarned  int64 `json:"total_earned"`
	TotalPaid    int64 `json:"total_paid"`
	TotalPending int64 `json:"total_pending"`
}
