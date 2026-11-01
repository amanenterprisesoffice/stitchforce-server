package app

import (
	"stitchforce-server/internal/repo"
	garmentworkerrepo "stitchforce-server/internal/repo/persistent/garmentworker"
	organizationrepo "stitchforce-server/internal/repo/persistent/organization"
	stylerepo "stitchforce-server/internal/repo/persistent/style"
	styleopsrepo "stitchforce-server/internal/repo/persistent/styleoperation"
	sizerepo "stitchforce-server/internal/repo/persistent/size"
	userrepo "stitchforce-server/internal/repo/persistent/user"
	"stitchforce-server/pkg/postgres"
)

type Repositories struct {
	Users           repo.UserRepository
	GarmentWorkers  repo.GarmentWorkerRepository
	Organization    repo.OrganizationRepository
	Style           repo.StyleRepository
	StyleOperations repo.StyleOperationRepository
	Sizes           repo.SizeRepository
}

func NewRepositories(database *postgres.Client) Repositories {
	return Repositories{
		Users:           userrepo.New(database.DB),
		GarmentWorkers:  garmentworkerrepo.New(database.DB),
		Organization:    organizationrepo.New(database.DB),
		Style:           stylerepo.New(database.DB),
		StyleOperations: styleopsrepo.New(database.DB),
		Sizes: sizerepo.New(database.DB),
	}
}
