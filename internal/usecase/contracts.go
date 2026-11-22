package usecase

import (
	"context"
	"time"

	"stitchforce-server/internal/entity"
	"stitchforce-server/internal/usecase/input"
	"stitchforce-server/internal/usecase/output"
)

type LicenseUseCase interface {
	Generate(
		ctx context.Context,
		organizationID string,
		durationDays int,
	) (*entity.GeneratedLicense, error)
	Activate(
		ctx context.Context,
		organizationID string,
		licenseKey string,
	) (*entity.License, error)
}

type OrganizationUseCase interface {
	Create(context.Context, *entity.Organization) error
	GetByID(context.Context, string) (*output.Organization, error)
	GetByUserAndOrganization(
		ctx context.Context,
		userID string,
		organizationID string,
	) (*output.Organization, error)
	List(context.Context) ([]*output.Organization, error)
	Update(context.Context, string, *input.UpdateOrganizationInput) error
	Delete(context.Context, string) error
	Invite(
		ctx context.Context,
		organizationID string,
		email string,
		role entity.UserRole,
	) error

	AcceptInvitation(
		ctx context.Context,
		rawToken string,
		name string,
		password string,
	) (*entity.User, error)

	ListInvitations(ctx context.Context, organizationID string) ([]*entity.OrganizationInvitation, error)
	ResendInvitation(ctx context.Context, organizationID, invitationID string) error
	CancelInvitation(ctx context.Context, organizationID, invitationID string) error

	GetDashboardOverview(
		context.Context,
		string, // organizationID
	) (*output.DashboardOverview, error)
	GetDashboardTrend(
		context.Context,
		string, // organizationID
		input.DashboardFilter,
	) (*output.OrganizationDashboard, error)

	SendDueDashboardAlerts(
		ctx context.Context,
		now time.Time,
	) (int, error)
}

type UserUseCase interface {
	Register(context.Context, string, string, string) (*entity.User, error)
	Login(
		ctx context.Context,
		email, password string,
		deviceName string,
		ipAddress string,
		userAgent string,
	) (*output.User, *entity.UserSession, error)
	RevokeSession(ctx context.Context, sessionID string) error
	GetActiveSession(
		ctx context.Context,
		sessionID string,
	) (*entity.UserSession, error)

	GetByID(context.Context, string) (*output.User, error)

	List(
		context.Context,
		string,
		input.UserFilter,
	) ([]*output.User, error)

	Update(
		context.Context,
		string,
		*input.UpdateUserInput,
	) error

	Delete(context.Context, string) error
	DeleteMe(context.Context, string, string) error

	// Authentication
	VerifyEmail(context.Context, string) error
	SendEmailVerification(context.Context, *entity.User) error

	ForgotPassword(context.Context, string) error
	ResetPassword(context.Context, string, string) error
}

type StyleUseCase interface {
	Create(
		context.Context,
		string, // organizationID
		*entity.Style,
	) error

	GetByID(
		context.Context,
		string, // organizationID
		string, // styleID
	) (*entity.Style, error)

	List(
		context.Context,
		string, // organizationID
		input.StyleFilter,
	) ([]*entity.Style, error)

	Update(
		context.Context,
		string, // organizationID
		*entity.Style,
	) error

	Delete(
		context.Context,
		string, // organizationID
		string, // styleID
	) error
}

type StyleOperationUseCase interface {
	Create(
		context.Context,
		string, // organizationID
		string, // styleID
		*entity.StyleOperation,
	) error

	GetByID(
		context.Context,
		string, // organizationID
		string, // styleID
		string, // operationID
	) (*entity.StyleOperation, error)

	List(
		context.Context,
		string, // organizationID
		string, // styleID
		input.StyleOperationFilter,
	) ([]*entity.StyleOperation, error)

	Update(
		context.Context,
		string, // organizationID
		string, // styleID
		*entity.StyleOperation,
	) error

	Delete(
		context.Context,
		string, // organizationID
		string, // styleID
		string, // operationID
	) error
}

type SizeUseCase interface {
	Create(
		context.Context,
		string, // organizationID
		*entity.Size,
	) error

	GetByID(
		context.Context,
		string, // organizationID
		string, // sizeID
	) (*entity.Size, error)

	List(
		context.Context,
		string, // organizationID
	) ([]*entity.Size, error)

	Update(
		context.Context,
		string, // organizationID
		*entity.Size,
	) error

	Delete(
		context.Context,
		string, // organizationID
		string, // sizeID
	) error
}

type BundleUseCase interface {
	Create(
		context.Context,
		string, // organizationID
		*entity.Bundle,
	) error

	GetByID(
		context.Context,
		string, // organizationID
		string, // bundleID
	) (*output.BundleDetail, error)

	List(
		context.Context,
		string, // organizationID
		input.BundleFilter,
	) ([]*entity.Bundle, error)

	Update(
		context.Context,
		string, // organizationID
		string, // bundleID
		*input.UpdateBundleInput,
	) error

	Delete(
		context.Context,
		string, // organizationID
		string, // bundleID
	) error
}

type GarmentWorkerUseCase interface {
	Create(context.Context, string, *input.CreateGarmentWorkerInput) error

	GetByID(
		context.Context,
		string, // organizationID
		string, // workerID
	) (*output.GarmentWorker, error)

	List(
		context.Context,
		string, // organizationID
		input.GarmentWorkerFilter,
	) ([]*output.GarmentWorker, error)

	Update(
		context.Context,
		string, // organizationID
		string, // workerID
		*input.UpdateGarmentWorkerInput,
	) error

	Delete(
		context.Context,
		string, // organizationID
		string, // workerID
	) error
}

type ProductionEntryUseCase interface {
	Create(
		context.Context,
		string,
		*entity.ProductionEntry,
	) error

	GetByID(
		context.Context,
		string,
		string,
	) (*output.ProductionEntry, error)

	List(
		context.Context,
		string,
		input.ProductionEntryFilter,
	) ([]*output.ProductionEntry, error)

	Update(
		context.Context,
		string,
		*entity.ProductionEntry,
	) error

	Delete(
		context.Context,
		string,
		string,
	) error
}

type WorkerPaymentUseCase interface {
	Create(
		context.Context,
		string, // organizationID
		*input.CreateWorkerPaymentInput,
	) error

	GetByID(
		context.Context,
		string, // organizationID
		string, // paymentID
	) (*output.WorkerPayment, error)

	List(
		context.Context,
		string, // organizationID
		input.WorkerPaymentFilter,
	) (*output.WorkerPaymentList, error)

	GetSummary(
		context.Context,
		string, // organizationID
		input.WorkerPaymentSummaryFilter,
	) (*output.WorkerPaymentSummary, error)

	ListWorkerBalances(
		context.Context,
		string, // organizationID
		input.WorkerPaymentBalanceFilter,
	) (*output.WorkerPaymentBalanceList, error)

	Void(
		context.Context,
		string, // organizationID
		string, // paymentID
	) error
}

type NotificationUseCase interface {
	Create(ctx context.Context, notification *entity.Notification) error

	CreateBulk(ctx context.Context, notifications []*entity.Notification) error

	List(
		ctx context.Context,
		userID string,
		limit int,
		offset int,
	) ([]*entity.Notification, error)

	GetUnreadCount(ctx context.Context, userID string) (int, error)

	MarkAsRead(ctx context.Context, userID, notificationID string) error

	MarkAllAsRead(ctx context.Context, userID string) error

	Delete(ctx context.Context, userID, notificationID string) error
}

type ReportUseCase interface {
	Create(ctx context.Context, report *entity.Report) error

	List(
		ctx context.Context,
		organizationID string,
		filter input.ReportFilter,
	) ([]*entity.Report, error)

	GetByID(
		ctx context.Context,
		organizationID string,
		reportID string,
	) (*entity.Report, error)

	GetDownloadURL(
		ctx context.Context,
		organizationID string,
		reportID string,
	) (string, error)

	Delete(
		ctx context.Context,
		organizationID string,
		reportID string,
	) error

	GenerateMonthlyReports(
		ctx context.Context,
		now time.Time,
	) (int, error)

	DeleteExpiredReports(
		ctx context.Context,
		now time.Time,
	) (int, error)

	SendExpirationReminders(
		ctx context.Context,
		now time.Time,
	) (int, error)
}
