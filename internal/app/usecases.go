package app

import (
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
	"stitchforce-server/pkg/logger"
)

type UseCases struct {
	Users           usecase.UserUseCase
	GarmentWorkers  usecase.GarmentWorkerUseCase
	Organization    usecase.OrganizationUseCase
	License         usecase.LicenseUseCase
	Style           usecase.StyleUseCase
	StyleOperations usecase.StyleOperationUseCase
	Size            usecase.SizeUseCase
	Bundle          usecase.BundleUseCase
	ProductionEntry usecase.ProductionEntryUseCase
	WorkerPayment usecase.WorkerPaymentUseCase
}

func NewUseCases(
	repositories Repositories,
	log *logger.Logger,
	storage storage.Storage,
) UseCases {
	return UseCases{
		Users: useruc.New(
			repositories.Users,
			log,
		),
		GarmentWorkers: garmentworkeruc.New(
			repositories.GarmentWorkers,
			log,
			storage,
		),
		Organization:    organizationuc.New(repositories.Organization, log),
		Style:           styleuc.New(repositories.Style, log),
		StyleOperations: styleopsuc.New(repositories.StyleOperations, log),
		Size:            sizeuc.New(repositories.Sizes, log),
		Bundle:          bundleuc.New(repositories.Bundles, storage, log),
		ProductionEntry: productionentryuc.New(repositories.ProductionEntry, repositories.Bundles, repositories.GarmentWorkers, repositories.StyleOperations, repositories.Style, storage, log),
		License:         licenseuc.New(repositories.Licenses, log),
		WorkerPayment:   workerpaymentuc.New(repositories.WorkerPayments, repositories.GarmentWorkers, storage, log),
	}
}
