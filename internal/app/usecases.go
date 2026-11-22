package app

import (
	"stitchforce-server/config"
	"stitchforce-server/internal/storage"
	"stitchforce-server/internal/usecase"
	bundleuc "stitchforce-server/internal/usecase/bundle"
	garmentworkeruc "stitchforce-server/internal/usecase/garmentworker"
	licenseuc "stitchforce-server/internal/usecase/license"
	notificationuc "stitchforce-server/internal/usecase/notification"
	organizationuc "stitchforce-server/internal/usecase/organization"
	productionentryuc "stitchforce-server/internal/usecase/productionentry"
	reportuc "stitchforce-server/internal/usecase/report"
	sizeuc "stitchforce-server/internal/usecase/size"
	styleuc "stitchforce-server/internal/usecase/style"
	styleopsuc "stitchforce-server/internal/usecase/styleoperation"
	useruc "stitchforce-server/internal/usecase/user"
	workerpaymentuc "stitchforce-server/internal/usecase/workerpayment"
	"stitchforce-server/pkg/email"
	"stitchforce-server/pkg/logger"
)

type UseCases struct {
	Users             usecase.UserUseCase
	Notifications     usecase.NotificationUseCase
	GarmentWorkers    usecase.GarmentWorkerUseCase
	Organizations     usecase.OrganizationUseCase
	License           usecase.LicenseUseCase
	Styles            usecase.StyleUseCase
	StyleOperations   usecase.StyleOperationUseCase
	Size              usecase.SizeUseCase
	Bundle            usecase.BundleUseCase
	ProductionEntries usecase.ProductionEntryUseCase
	WorkerPayments    usecase.WorkerPaymentUseCase
	Reports           usecase.ReportUseCase
}

func NewUseCases(
	repositories Repositories,
	log *logger.Logger,
	storage storage.Storage,
	emailer email.Sender,
	cfg config.Config,
) UseCases {
	users := useruc.New(
		repositories.Users,
		repositories.Sessions,
		log,
		storage,
		emailer,
		cfg.Frontend,
	)

	notifications := notificationuc.New(
		repositories.Notifications,
		log,
	)

	organizations := organizationuc.New(
		repositories.Organizations,
		repositories.Invitations,
		repositories.Users,
		storage,
		emailer,
		cfg.Frontend,
		log,
	)

	reports := reportuc.New(
		repositories.Reports,
		repositories.Archives,
		users,
		notifications,
		emailer,
		storage,
		cfg.Frontend,
		log,
	)

	return UseCases{
		Users:             users,
		Notifications:     notifications,
		Organizations:     organizations,
		Reports:           reports,
		GarmentWorkers:    garmentworkeruc.New(repositories.GarmentWorkers, storage, log),
		Styles:            styleuc.New(repositories.Styles, log),
		StyleOperations:   styleopsuc.New(repositories.StyleOperations, log),
		Size:              sizeuc.New(repositories.Sizes, log),
		Bundle:            bundleuc.New(repositories.Bundles, storage, log),
		ProductionEntries: productionentryuc.New(repositories.ProductionEntries, repositories.Bundles, repositories.GarmentWorkers, repositories.StyleOperations, repositories.Styles, storage, log),
		License:           licenseuc.New(repositories.Licenses, log),
		WorkerPayments:    workerpaymentuc.New(repositories.WorkerPayments, repositories.GarmentWorkers, storage, log),
	}
}
