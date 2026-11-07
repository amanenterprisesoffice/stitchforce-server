package entity

import "time"

type PaymentMethod string

const (
	PaymentMethodCash   PaymentMethod = "cash"
	PaymentMethodBank   PaymentMethod = "bank"
	PaymentMethodUPI    PaymentMethod = "upi"
	PaymentMethodCheque PaymentMethod = "cheque"
)

func (p PaymentMethod) Validate() bool {
	switch p {
	case PaymentMethodCash, PaymentMethodBank, PaymentMethodUPI, PaymentMethodCheque:
		return true
	default:
		return false
	}
}

type PaymentStatus string

const (
	PaymentStatusCompleted PaymentStatus = "completed"
	PaymentStatusVoided    PaymentStatus = "voided"
)

func (p PaymentStatus) Validate() bool {
	switch p {
	case PaymentStatusCompleted, PaymentStatusVoided:
		return true
	default:
		return false
	}
}

type WorkerPayment struct {
	ID             string `json:"id" db:"id"`
	OrganizationID string `json:"organization_id" db:"organization_id"`
	WorkerID       string `json:"worker_id" db:"worker_id"`

	Amount        int64         `json:"amount" db:"amount"`
	PaymentDate   time.Time     `json:"payment_date" db:"payment_date"`
	PaymentMethod PaymentMethod `json:"payment_method" db:"payment_method"`
	Reference     string        `json:"reference" db:"reference"`
	Remarks       string        `json:"remarks" db:"remarks"`
	Status        PaymentStatus `json:"status" db:"status"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
