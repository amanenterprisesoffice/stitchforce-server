package output

import (
	"stitchforce-server/internal/entity"
	"time"
)

type WorkerPayment struct {
	ID             string `json:"id" db:"id"`
	OrganizationID string `json:"organization_id" db:"organization_id"`
	WorkerID       string `json:"worker_id" db:"worker_id"`

	Amount        int64                `json:"amount" db:"amount"`
	PaymentDate   time.Time            `json:"payment_date" db:"payment_date"`
	PaymentMethod entity.PaymentMethod `json:"payment_method" db:"payment_method"`
	Reference     string               `json:"reference" db:"reference"`
	Remarks       string               `json:"remarks" db:"remarks"`
	Status        entity.PaymentStatus `json:"status" db:"status"`

	CreatedAt       time.Time             `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at" db:"updated_at"`
	ProductionEntry []PaidProductionEntry `json:"production_entry" db:"production_entry"`
	Worker          ProductionEntryWorker `json:"worker" db:"worker"`
}

type PaidProductionEntry struct {
	ProductionEntry
	AllocatedAmount int64 `json:"allocated_amount"`
}

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
