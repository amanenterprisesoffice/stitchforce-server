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
	usersessionrepo "stitchforce-server/internal/repo/postgres/usersession"
	notificationrepo "stitchforce-server/internal/repo/postgres/notification"
	workerpaymentrepo "stitchforce-server/internal/repo/postgres/workerpayment"
	reportrepo "stitchforce-server/internal/repo/postgres/report"
	archiverepo "stitchforce-server/internal/repo/postgres/archive"
	"stitchforce-server/pkg/postgres"
)

type Repositories struct {
	Users           repo.UserRepository
	Notifications  repo.NotificationRepository
	Sessions        repo.UserSessionRepository
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
	Reports         repo.ReportRepository
	Archives        repo.ArchiveRepository
}

func NewRepositories(database *postgres.Client) Repositories {
	return Repositories{
		Users:           userrepo.New(database.DB),
		Notifications:   notificationrepo.New(database.DB),
		Sessions:        usersessionrepo.New(database.DB),
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
		Reports:         reportrepo.New(database.DB),
		Archives:        archiverepo.New(database.DB), // Archives repository is not implemented yet
	}
}
