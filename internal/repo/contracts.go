package repo

import (
	"context"
	"errors"

	"stitchforce-server/internal/entity"
	"stitchforce-server/internal/usecase/input"
	"stitchforce-server/internal/usecase/output"
)

var ErrNotFound = errors.New("not found")

type OrganizationRepository interface {
	Create(context.Context, *entity.Organization) error
	GetByID(context.Context, string) (*entity.Organization, error)
	List(context.Context) ([]*entity.Organization, error)
	Update(context.Context, *entity.Organization) error
	Delete(context.Context, string) error
	GetDashboardOverview(
		ctx context.Context,
		organizationID string,
	) (*output.DashboardOverview, error)
	GetDashboardTrend(
		ctx context.Context,
		organizationID string,
		filter input.DashboardFilter,
	) (*output.OrganizationDashboard, error)
}

type UserRepository interface {
	Create(context.Context, *entity.User) error
	GetByID(context.Context, string) (*entity.User, error)
	GetByEmail(context.Context, string) (*entity.User, error)
	List(context.Context) ([]*entity.User, error)
	Update(context.Context, *entity.User) error
	Delete(context.Context, string) error
	RegisterOrganization(
		context.Context,
		*entity.Organization,
		*entity.User,
	) error
}

type StyleRepository interface {
	Create(context.Context, string, *entity.Style) error
	GetByID(context.Context, string, string) (*entity.Style, error)
	List(context.Context, string, input.StyleFilter) ([]*entity.Style, error)
	Update(context.Context, string, *entity.Style) error
	Delete(context.Context, string, string) error
}

type StyleOperationRepository interface {
	Create(context.Context, string, string, *entity.StyleOperation) error

	GetByID(
		context.Context,
		string,
		string,
		string,
	) (*entity.StyleOperation, error)

	List(
		context.Context,
		string, // organizationID
		string, // styleID
		input.StyleOperationFilter,
	) ([]*entity.StyleOperation, error)

	Update(context.Context, string, string, *entity.StyleOperation) error
	Delete(context.Context, string, string, string) error
}

type SizeRepository interface {
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

type BundleRepository interface {
	Create(
		context.Context,
		string, // organizationID
		*entity.Bundle,
	) error

	GetByID(
		context.Context,
		string,
		string,
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

type GarmentWorkerRepository interface {
	Create(
		context.Context,
		string, // organizationID
		*entity.GarmentWorker,
	) error

	GetByID(
		context.Context,
		string, // organizationID
		string, // workerID
	) (*entity.GarmentWorker, error)

	List(
		context.Context,
		string, // organizationID
		input.GarmentWorkerFilter,
	) ([]*entity.GarmentWorker, error)

	Update(
		context.Context,
		string, // organizationID
		*entity.GarmentWorker,
	) error

	Delete(
		context.Context,
		string, // organizationID
		string, // workerID
	) error
}

type ProductionEntryRepository interface {
	Create(
		context.Context,
		string, // organizationID
		*entity.ProductionEntry,
	) error

	GetByID(
		context.Context,
		string, // organizationID
		string, // productionEntryID
	) (*entity.ProductionEntry, error)

	List(
		context.Context,
		string, // organizationID
		input.ProductionEntryFilter,
	) ([]*output.ProductionEntry, error)

	UpdateWithBundleLimit(
		context.Context,
		string, // organizationID
		*entity.ProductionEntry,
	) error

	Delete(
		context.Context,
		string, // organizationID
		string, // productionEntryID
	) error
}

type WorkerPaymentRepository interface {
	CreateWithAllocations(
		context.Context,
		string, // organizationID
		*entity.WorkerPayment,
		input.ProductionEntryFilter,
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
	) ([]*output.WorkerPayment, error)

	GetSummary(
		context.Context,
		string, // organizationID
		input.WorkerPaymentSummaryFilter, // workerID
	) (*output.WorkerPaymentSummary, error)

	ListWorkerBalances(
		ctx context.Context,
		organizationID string,
		filter input.WorkerPaymentBalanceFilter,
	) ([]*output.WorkerPaymentBalance, error)

	Void(
		context.Context,
		string, // organizationID
		string, // paymentID
	) error
}
