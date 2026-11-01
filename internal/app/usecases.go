package app

import (
	"stitchforce-server/internal/storage"
	"stitchforce-server/internal/usecase"
	garmentworkeruc "stitchforce-server/internal/usecase/garmentworker"
	organizationuc "stitchforce-server/internal/usecase/organization"
	styleuc "stitchforce-server/internal/usecase/style"
	styleopsuc "stitchforce-server/internal/usecase/styleoperation"
	useruc "stitchforce-server/internal/usecase/user"
	"stitchforce-server/pkg/logger"
)

type UseCases struct {
	Users           usecase.UserUseCase
	GarmentWorkers  usecase.GarmentWorkerUseCase
	Organization    usecase.OrganizationUseCase
	Style           usecase.StyleUseCase
	StyleOperations usecase.StyleOperationUseCase
}

func NewUseCases(
	repositories Repositories,
	log *logger.Logger,
	storage storage.Storage,
) UseCases {
	return UseCases{
		Users: useruc.New(
			repositories.Users,
			log,
		),
		GarmentWorkers: garmentworkeruc.New(
			repositories.GarmentWorkers,
			log,
			storage,
		),
		Organization:    organizationuc.New(repositories.Organization, log),
		Style:           styleuc.New(repositories.Style, log),
		StyleOperations: styleopsuc.New(repositories.StyleOperations, log),
	}
}
