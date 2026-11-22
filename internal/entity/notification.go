package entity

import "time"

type NotificationType string

const (
	NotificationLicenseExpiry NotificationType = "license_expiry"
	NotificationReportReady   NotificationType = "report_ready"
	NotificationProduction    NotificationType = "production"
	NotificationPayment       NotificationType = "payment"
	NotificationSecurity      NotificationType = "security"
	NotificationSystem        NotificationType = "system"
)

type Notification struct {
	ID             string            `json:"id" db:"id"`
	OrganizationID string            `json:"organization_id" db:"organization_id"`
	UserID         string            `json:"user_id" db:"user_id"`
	Type           NotificationType  `json:"type" db:"type"`
	Title          string            `json:"title" db:"title"`
	Message        string            `json:"message" db:"message"`
	Data           map[string]any    `json:"data" db:"data"`
	ActionURL      string            `json:"action_url" db:"action_url"`
	ReadAt         *time.Time        `json:"read_at,omitempty" db:"read_at"`
	CreatedAt      time.Time         `json:"created_at" db:"created_at"`
}