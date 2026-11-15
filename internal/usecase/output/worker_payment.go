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
	Worker          SlimWorker            `json:"worker" db:"worker"`
}

type SlimWorkerPayment struct {
	ID            string               `json:"id"`
	WorkerID      string               `json:"worker_id"`
	Amount        int64                `json:"amount"`
	PaymentDate   time.Time            `json:"payment_date"`
	PaymentMethod entity.PaymentMethod `json:"payment_method"`
	Reference     string               `json:"reference"`
	Remarks       string               `json:"remarks"`
	Status        entity.PaymentStatus `json:"status"`
	CreatedAt     time.Time            `json:"created_at"`
	Worker        SlimWorker           `json:"worker"`
}

type WorkerPaymentListSummary struct {
	TotalPayments     int   `json:"total_payments"`
	TotalAmount       int64 `json:"total_amount"`
	CompletedPayments int   `json:"completed_payments"`
	CompletedAmount   int64 `json:"completed_amount"`
}

type WorkerPaymentList struct {
	Summary  WorkerPaymentListSummary `json:"summary"`
	Payments []*SlimWorkerPayment     `json:"payments"`
}

type PaidProductionEntry struct {
	ProductionEntry
	AllocatedAmount int64 `json:"allocated_amount"`
}

type WorkerPaymentSummary struct {
	TotalEarned  int64 `json:"total_earned"`
	TotalPaid    int64 `json:"total_paid"`
	TotalPending int64 `json:"total_pending"`
}

type WorkerPaymentBalance struct {
	Worker   SlimWorker `json:"worker"`
	PhotoURL string     `json:"photo_url,omitempty"`

	TotalEarned  int64 `json:"total_earned"`
	TotalPaid    int64 `json:"total_paid"`
	TotalPending int64 `json:"total_pending"`

	PendingEntries int `json:"pending_entries"`
}

type WorkerPaymentBalanceList struct {
	Summary WorkerPaymentBalanceSummary `json:"summary"`
	Workers []*WorkerPaymentBalance     `json:"workers"`
}

type WorkerPaymentBalanceSummary struct {
	TotalEarned  int64 `json:"total_earned"`
	TotalPaid    int64 `json:"total_paid"`
	TotalPending int64 `json:"total_pending"`

	PendingEntries int `json:"pending_entries"`
	TotalWorkers   int `json:"total_workers"`
}
