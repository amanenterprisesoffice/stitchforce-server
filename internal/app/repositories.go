package app

import (
	"stitchforce-server/internal/repo"
	bundlerepo "stitchforce-server/internal/repo/persistent/bundle"
	productionentryrepo "stitchforce-server/internal/repo/persistent/productionentry"
	garmentworkerrepo "stitchforce-server/internal/repo/persistent/garmentworker"
	organizationrepo "stitchforce-server/internal/repo/persistent/organization"
	sizerepo "stitchforce-server/internal/repo/persistent/size"
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
	Sizes           repo.SizeRepository
	Bundles          repo.BundleRepository
	ProductionEntry  repo.ProductionEntryRepository
}

func NewRepositories(database *postgres.Client) Repositories {
	return Repositories{
		Users:           userrepo.New(database.DB),
		GarmentWorkers:  garmentworkerrepo.New(database.DB),
		Organization:    organizationrepo.New(database.DB),
		Style:           stylerepo.New(database.DB),
		StyleOperations: styleopsrepo.New(database.DB),
		Sizes:           sizerepo.New(database.DB),
		Bundles:          bundlerepo.New(database.DB),
		ProductionEntry:  productionentryrepo.New(database.DB),
	}
}
