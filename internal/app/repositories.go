package app

import (
	"stitchforce-server/internal/repo"
	bundlerepo "stitchforce-server/internal/repo/postgres/bundle"
	garmentworkerrepo "stitchforce-server/internal/repo/postgres/garmentworker"
	licenserepo "stitchforce-server/internal/repo/postgres/license"
	organizationrepo "stitchforce-server/internal/repo/postgres/organization"
	organizationinvitationrepo "stitchforce-server/internal/repo/postgres/organizationinvitation"
	productionentryrepo "stitchforce-server/internal/repo/postgres/productionentry"
	sizerepo "stitchforce-server/internal/repo/postgres/size"
	stylerepo "stitchforce-server/internal/repo/postgres/style"
	styleopsrepo "stitchforce-server/internal/repo/postgres/styleoperation"
	userrepo "stitchforce-server/internal/repo/postgres/user"
	workerpaymentrepo "stitchforce-server/internal/repo/postgres/workerpayment"
	"stitchforce-server/pkg/postgres"
)

type Repositories struct {
	Users           repo.UserRepository
	GarmentWorkers  repo.GarmentWorkerRepository
	Organizations   repo.OrganizationRepository
	Invitations     repo.OrganizationInvitationRepository
	Licenses        repo.LicenseRepository
	Styles          repo.StyleRepository
	StyleOperations repo.StyleOperationRepository
	Sizes           repo.SizeRepository
	Bundles         repo.BundleRepository
	ProductionEntries repo.ProductionEntryRepository
	WorkerPayments  repo.WorkerPaymentRepository
}

func NewRepositories(database *postgres.Client) Repositories {
	return Repositories{
		Users:           userrepo.New(database.DB),
		GarmentWorkers:  garmentworkerrepo.New(database.DB),
		Organizations:    organizationrepo.New(database.DB),
		Invitations:     organizationinvitationrepo.New(database.DB),
		Licenses:        licenserepo.New(database.DB),
		Styles:           stylerepo.New(database.DB),
		StyleOperations: styleopsrepo.New(database.DB),
		Sizes:           sizerepo.New(database.DB),
		Bundles:         bundlerepo.New(database.DB),
		ProductionEntries: productionentryrepo.New(database.DB),
		WorkerPayments:  workerpaymentrepo.New(database.DB),
	}
}
