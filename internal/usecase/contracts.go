package usecase

import (
	"context"

	"stitchforce-server/internal/entity"
	"stitchforce-server/internal/usecase/input"
	"stitchforce-server/internal/usecase/output"
)

type OrganizationUseCase interface {
	Create(context.Context, *entity.Organization) error
	GetByID(context.Context, string) (*entity.Organization, error)
	List(context.Context) ([]*entity.Organization, error)
	Update(context.Context, *entity.Organization) error
	Delete(context.Context, string) error
}

type UserUseCase interface {
	Register(context.Context, string, string, string) (*entity.User, error)
	Login(context.Context, string, string) (*entity.User, error)

	GetByID(context.Context, string) (*entity.User, error)
	List(context.Context) ([]*entity.User, error)
	Update(context.Context, *entity.User) error
	Delete(context.Context, string) error
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
	) (*entity.Bundle, error)

	List(
		context.Context,
		string, // organizationID
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
	) ([]*output.GarmentWorker, error)

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

	SetStatus(
		context.Context,
		string, // organizationID
		string, // workerID
		entity.WorkerStatus, // status
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
