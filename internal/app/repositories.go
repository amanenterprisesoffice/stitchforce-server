package app

import (
	"stitchforce-server/internal/repo"
	garmentworkerrepo "stitchforce-server/internal/repo/persistent/garmentworker"
	userrepo "stitchforce-server/internal/repo/persistent/user"
	"stitchforce-server/pkg/postgres"
)

type Repositories struct {
	Users          repo.UserRepository
	GarmentWorkers repo.GarmentWorkerRepository
}

func NewRepositories(database *postgres.Client) Repositories {
	return Repositories{
		Users:          userrepo.New(database.DB),
		GarmentWorkers: garmentworkerrepo.New(database.DB),
	}
}