package output

import (
	"stitchforce-server/internal/entity"
	"time"
)

type Organization struct {
	ID               string     `json:"id"`
	Name             string     `json:"name"`
	AddressLine1     string     `json:"address_line_1"`
	AddressLine2     string     `json:"address_line_2"`
	City             string     `json:"city"`
	State            string     `json:"state"`
	PostalCode       string     `json:"postal_code"`
	Country          string     `json:"country"`
	PhotoURL         *string    `json:"photo_url,omitempty"`
	LicenseStatus    string     `json:"license_status"`
	LicenseExpiresAt *time.Time `json:"license_expires_at"`

	DashboardAlertEnabled    bool                          `json:"dashboard_alert_enabled" db:"dashboard_alert_enabled"`
	DashboardAlertEmail      string                        `json:"dashboard_alert_email" db:"dashboard_alert_email"`
	DashboardAlertSchedule   entity.DashboardAlertSchedule `json:"dashboard_alert_schedule" db:"dashboard_alert_schedule"`
	DashboardAlertLastSentAt *time.Time                    `json:"dashboard_alert_last_sent_at" db:"dashboard_alert_last_sent_at"`

	WorkerPaymentCycle *entity.WorkerPaymentCycle `json:"worker_payment_cycle,omitempty" db:"worker_payment_cycle"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
