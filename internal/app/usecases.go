package app

import (
	"stitchforce-server/internal/usecase"
	garmentworkeruc "stitchforce-server/internal/usecase/garmentworker"
	useruc "stitchforce-server/internal/usecase/user"
	"stitchforce-server/pkg/logger"
)

type UseCases struct {
	Users          usecase.UserUseCase
	GarmentWorkers usecase.GarmentWorkerUseCase
}

func NewUseCases(
	repositories Repositories,
	log *logger.Logger,
) UseCases {
	return UseCases{
		Users: useruc.New(
			repositories.Users,
			log,
		),
		GarmentWorkers: garmentworkeruc.New(
			repositories.GarmentWorkers,
			log,
		),
	}
}
