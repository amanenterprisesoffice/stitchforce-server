package repo

import (
	"context"
	"errors"

	"stitchforce-server/internal/entity"
)

var ErrNotFound = errors.New("not found")

type OrganizationRepository interface {
	Create(context.Context, *entity.Organization) error
	GetByID(context.Context, string) (*entity.Organization, error)
	List(context.Context) ([]*entity.Organization, error)
	Update(context.Context, *entity.Organization) error
	Delete(context.Context, string) error
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

// type GarmentWorkerRepository interface {
// 	Create(context.Context, *entity.GarmentWorker) error
// 	GetByID(context.Context, string) (*entity.GarmentWorker, error)
// 	List(context.Context) ([]*entity.GarmentWorker, error)
// 	Update(context.Context, *entity.GarmentWorker) error
// 	Delete(context.Context, string) error
// }

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
