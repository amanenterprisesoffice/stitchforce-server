package output

import "time"

type PaymentAllocation struct {
	ID                string
	OrganizationID    string
	WorkerPaymentID   string
	ProductionEntryID string
	Amount            int64
	CreatedAt         time.Time
}


// ArchiveOrganization contains only the organization fields needed by the
// monthly archive generator.
type ArchiveOrganization struct {
	ID        string    `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}


type ArchivePeriod struct {
    Year  int `json:"year" db:"year"`
    Month int `json:"month" db:"month"`
}