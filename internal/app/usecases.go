package app

import (
	"stitchforce-server/config"
	"stitchforce-server/internal/storage"
	"stitchforce-server/internal/usecase"
	bundleuc "stitchforce-server/internal/usecase/bundle"
	garmentworkeruc "stitchforce-server/internal/usecase/garmentworker"
	licenseuc "stitchforce-server/internal/usecase/license"
	organizationuc "stitchforce-server/internal/usecase/organization"
	productionentryuc "stitchforce-server/internal/usecase/productionentry"
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
	GarmentWorkers    usecase.GarmentWorkerUseCase
	Organizations     usecase.OrganizationUseCase
	License           usecase.LicenseUseCase
	Styles            usecase.StyleUseCase
	StyleOperations   usecase.StyleOperationUseCase
	Size              usecase.SizeUseCase
	Bundle            usecase.BundleUseCase
	ProductionEntries usecase.ProductionEntryUseCase
	WorkerPayments    usecase.WorkerPaymentUseCase
}

func NewUseCases(
	repositories Repositories,
	log *logger.Logger,
	storage storage.Storage,
	emailer email.Sender,
	cfg config.Config,
) UseCases {
	return UseCases{
		Users: useruc.New(
			repositories.Users,
			log,
			storage,
			emailer,
			cfg.Frontend,
		),
		GarmentWorkers: garmentworkeruc.New(
			repositories.GarmentWorkers,
			storage,
			log,
		),
		Organizations: organizationuc.New(repositories.Organizations,
			repositories.Invitations,
			repositories.Users,
			storage,
			emailer,
			cfg.Frontend,
			log),
		Styles:            styleuc.New(repositories.Styles, log),
		StyleOperations:   styleopsuc.New(repositories.StyleOperations, log),
		Size:              sizeuc.New(repositories.Sizes, log),
		Bundle:            bundleuc.New(repositories.Bundles, storage, log),
		ProductionEntries: productionentryuc.New(repositories.ProductionEntries, repositories.Bundles, repositories.GarmentWorkers, repositories.StyleOperations, repositories.Styles, storage, log),
		License:           licenseuc.New(repositories.Licenses, log),
		WorkerPayments:    workerpaymentuc.New(repositories.WorkerPayments, repositories.GarmentWorkers, storage, log),
	}
}
