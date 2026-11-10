package app

import (
	"stitchforce-server/internal/repo"
	bundlerepo "stitchforce-server/internal/repo/postgres/bundle"
	garmentworkerrepo "stitchforce-server/internal/repo/postgres/garmentworker"
	organizationrepo "stitchforce-server/internal/repo/postgres/organization"
	productionentryrepo "stitchforce-server/internal/repo/postgres/productionentry"
	sizerepo "stitchforce-server/internal/repo/postgres/size"
	stylerepo "stitchforce-server/internal/repo/postgres/style"
	styleopsrepo "stitchforce-server/internal/repo/postgres/styleoperation"
	userrepo "stitchforce-server/internal/repo/postgres/user"
	workerpaymentrepo "stitchforce-server/internal/repo/postgres/workerpayment"
	licenserepo "stitchforce-server/internal/repo/postgres/license"
	"stitchforce-server/pkg/postgres"
)

type Repositories struct {
	Users           repo.UserRepository
	GarmentWorkers  repo.GarmentWorkerRepository
	Organization    repo.OrganizationRepository
	Licenses        repo.LicenseRepository
	Style           repo.StyleRepository
	StyleOperations repo.StyleOperationRepository
	Sizes           repo.SizeRepository
	Bundles         repo.BundleRepository
	ProductionEntry repo.ProductionEntryRepository
	WorkerPayments  repo.WorkerPaymentRepository
}

func NewRepositories(database *postgres.Client) Repositories {
	return Repositories{
		Users:           userrepo.New(database.DB),
		GarmentWorkers:  garmentworkerrepo.New(database.DB),
		Organization:    organizationrepo.New(database.DB),
		Licenses:        licenserepo.New(database.DB),
		Style:           stylerepo.New(database.DB),
		StyleOperations: styleopsrepo.New(database.DB),
		Sizes:           sizerepo.New(database.DB),
		Bundles:         bundlerepo.New(database.DB),
		ProductionEntry: productionentryrepo.New(database.DB),
		WorkerPayments:  workerpaymentrepo.New(database.DB),
	}
}
