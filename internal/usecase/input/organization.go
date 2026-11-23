package input

import "stitchforce-server/internal/entity"

type UpdateOrganizationInput struct {
	Name                  *string
	AddressLine1          *string
	AddressLine2          *string
	City                  *string
	State                 *string
	PostalCode            *string
	Country               *string
	Photo                 *UploadFile
	DashboardAlertEnabled *bool
	DashboardAlertEmail   *string
	DashboardAlertSchedule *entity.DashboardAlertSchedule

	WorkerPaymentCycle *entity.WorkerPaymentCycle
}