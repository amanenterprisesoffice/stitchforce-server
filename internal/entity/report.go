package entity

import "time"

type ReportType string

const (
	ReportMonthlyArchive ReportType = "monthly_archive"
)

type Report struct {
	ID                       string         `json:"id" db:"id"`
	OrganizationID           string         `json:"organization_id" db:"organization_id"`
	Type                     ReportType     `json:"type" db:"type"`
	Year                     int            `json:"year" db:"year"`
	Month                    int            `json:"month" db:"month"`
	FileKey                  string         `json:"-" db:"file_key"`
	FileSize                 int64          `json:"file_size" db:"file_size"`
	RecordCounts             map[string]int `json:"record_counts" db:"record_counts"`
	ArchivedAt               time.Time      `json:"archived_at" db:"archived_at"`
	ExpiresAt                time.Time      `json:"expires_at" db:"expires_at"`
	CreatedAt                time.Time      `json:"created_at" db:"created_at"`
	CleanupCompleted         bool           `json:"-" db:"cleanup_completed"`
	ExpirationReminderSentAt *time.Time      `json:"-" db:"expiration_reminder_sent_at"`
}
