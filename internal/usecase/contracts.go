package usecase

import (
	"context"

	"stitchforce-server/internal/entity"
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

type GarmentWorkerUseCase interface {
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
