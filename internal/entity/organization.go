package entity

import (
	"errors"
	"fmt"
	"time"
)

type Organization struct {
	ID                       string                 `json:"id" db:"id"`
	Name                     string                 `json:"name" db:"name"`
	AddressLine1             string                 `json:"address_line_1" db:"address_line_1"`
	AddressLine2             string                 `json:"address_line_2" db:"address_line_2"`
	City                     string                 `json:"city" db:"city"`
	State                    string                 `json:"state" db:"state"`
	PostalCode               string                 `json:"postal_code" db:"postal_code"`
	Country                  string                 `json:"country" db:"country"`
	PhotoKey                 string                 `json:"-" db:"photo_key"`
	WorkerPaymentCycle       WorkerPaymentCycle     `json:"worker_payment_cycle" db:"worker_payment_cycle"`
	
	DashboardAlertEnabled    bool                   `json:"dashboard_alert_enabled" db:"dashboard_alert_enabled"`
	DashboardAlertEmail      string                 `json:"dashboard_alert_email" db:"dashboard_alert_email"`
	DashboardAlertSchedule   DashboardAlertSchedule `json:"dashboard_alert_schedule" db:"dashboard_alert_schedule"`
	DashboardAlertLastSentAt *time.Time             `json:"dashboard_alert_last_sent_at" db:"dashboard_alert_last_sent_at"`

	LicenseStatus    string     `json:"license_status" db:"license_status"`
	LicenseExpiresAt *time.Time `json:"license_expires_at" db:"license_expires_at"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type WorkerPaymentCycle string

const (
	WorkerPaymentCycleHalfMonthly WorkerPaymentCycle = "half_monthly"
	WorkerPaymentCycleMonthly     WorkerPaymentCycle = "monthly"
	WorkerPaymentCycleBiMonthly   WorkerPaymentCycle = "bi_monthly"
	WorkerPaymentCycleQuarterly   WorkerPaymentCycle = "quarterly"
	WorkerPaymentCycleHalfYearly  WorkerPaymentCycle = "half_yearly"
	WorkerPaymentCycleYearly      WorkerPaymentCycle = "yearly"
)

func (c WorkerPaymentCycle) Validate() error {
	switch c {
	case WorkerPaymentCycleHalfMonthly,
		WorkerPaymentCycleMonthly,
		WorkerPaymentCycleBiMonthly,
		WorkerPaymentCycleQuarterly,
		WorkerPaymentCycleHalfYearly,
		WorkerPaymentCycleYearly:
		return nil
	default:
		return errors.New("invalid worker payment cycle")
	}
}

type Schedule string

const (
	Daily   Schedule = "daily"
	Weekly  Schedule = "weekly"
	Monthly Schedule = "monthly"
)

type Weekday string

const (
	Monday    Weekday = "monday"
	Tuesday   Weekday = "tuesday"
	Wednesday Weekday = "wednesday"
	Thursday  Weekday = "thursday"
	Friday    Weekday = "friday"
	Saturday  Weekday = "saturday"
	Sunday    Weekday = "sunday"
)

type DashboardAlertSchedule struct {
	Frequency Schedule `json:"frequency"`

	// 24-hour format: HH:mm
	Time string `json:"time"`

	// Required for weekly schedules.
	DayOfWeek Weekday `json:"day_of_week,omitempty"`

	// Required for monthly schedules (1-31).
	DayOfMonth int `json:"day_of_month,omitempty"`

	// IANA timezone, e.g. Asia/Kolkata.
	Timezone string `json:"timezone"`
}

func (s DashboardAlertSchedule) Validate() error {
	switch s.Frequency {
	case Daily:
	case Weekly:
		switch s.DayOfWeek {
		case Monday, Tuesday, Wednesday, Thursday, Friday, Saturday, Sunday:
		default:
			return fmt.Errorf("invalid day of week")
		}
	case Monthly:
		if s.DayOfMonth < 1 || s.DayOfMonth > 31 {
			return fmt.Errorf("day of month must be between 1 and 31")
		}
	default:
		return fmt.Errorf("invalid schedule frequency: %s", s.Frequency)
	}

	if _, err := time.Parse("15:04", s.Time); err != nil {
		return fmt.Errorf("invalid schedule time: expected HH:mm")
	}

	if _, err := time.LoadLocation(s.Timezone); err != nil {
		return fmt.Errorf("invalid timezone: %s", s.Timezone)
	}

	return nil
}
