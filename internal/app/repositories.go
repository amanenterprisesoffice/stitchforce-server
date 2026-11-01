package app

import (
	"stitchforce-server/internal/repo"
	garmentworkerrepo "stitchforce-server/internal/repo/persistent/garmentworker"
	organizationrepo "stitchforce-server/internal/repo/persistent/organization"
	stylerepo "stitchforce-server/internal/repo/persistent/style"
	styleopsrepo "stitchforce-server/internal/repo/persistent/styleoperation"
	userrepo "stitchforce-server/internal/repo/persistent/user"
	"stitchforce-server/pkg/postgres"
)

type Repositories struct {
	Users           repo.UserRepository
	GarmentWorkers  repo.GarmentWorkerRepository
	Organization    repo.OrganizationRepository
	Style           repo.StyleRepository
	StyleOperations repo.StyleOperationRepository
}

func NewRepositories(database *postgres.Client) Repositories {
	return Repositories{
		Users:           userrepo.New(database.DB),
		GarmentWorkers:  garmentworkerrepo.New(database.DB),
		Organization:    organizationrepo.New(database.DB),
		Style:           stylerepo.New(database.DB),
		StyleOperations: styleopsrepo.New(database.DB),
	}
}
