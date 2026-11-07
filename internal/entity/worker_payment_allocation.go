package entity

import "time"

type WorkerPaymentAllocation struct {
	ID                string    `json:"id" db:"id"`
	OrganizationID    string    `json:"organization_id" db:"organization_id"`
	WorkerPaymentID   string    `json:"worker_payment_id" db:"worker_payment_id"`
	ProductionEntryID string    `json:"production_entry_id" db:"production_entry_id"`
	Amount            int64     `json:"amount" db:"amount"`

	CreatedAt         time.Time `json:"created_at" db:"created_at"`
}