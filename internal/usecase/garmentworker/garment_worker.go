
package garmentworker

import (
	"context"
	"errors"
	"strings"

	"stitchforce-server/internal/entity"
	"stitchforce-server/internal/repo"
	"stitchforce-server/internal/usecase"
	"stitchforce-server/pkg/logger"
)

type UseCase struct {
	repository repo.GarmentWorkerRepository
	logger     *logger.Logger
}

func New(
	repository repo.GarmentWorkerRepository,
	logger *logger.Logger,
) usecase.GarmentWorkerUseCase {
	return &UseCase{
		repository: repository,
		logger:     logger,
	}
}

func validateOrganizationID(organizationID string) (string, error) {
	organizationID = strings.TrimSpace(organizationID)

	if organizationID == "" {
		return "", errors.New("organization id is required")
	}

	return organizationID, nil
}

func validateWorker(worker *entity.GarmentWorker) error {
	if worker == nil {
		return errors.New("garment worker is required")
	}

	worker.Name = strings.TrimSpace(worker.Name)
	worker.Phone = strings.TrimSpace(worker.Phone)
	worker.Gender = strings.ToLower(strings.TrimSpace(worker.Gender))

	if worker.Name == "" {
		return errors.New("name is required")
	}

	if worker.Phone == "" {
		return errors.New("phone is required")
	}

	switch worker.Gender {
	case "male", "female", "other":
	default:
		return errors.New("gender must be male, female, or other")
	}

	return nil
}

func (u *UseCase) Create(
	ctx context.Context,
	organizationID string,
	worker *entity.GarmentWorker,
) error {
	var err error

	organizationID, err = validateOrganizationID(organizationID)
	if err != nil {
		return err
	}

	if err := validateWorker(worker); err != nil {
		return err
	}

	worker.OrganizationID = organizationID

	u.logger.InfoContext(
		ctx,
		"creating garment worker",
		"organization_id", organizationID,
		"name", worker.Name,
	)

	return u.repository.Create(ctx, organizationID, worker)
}

func (u *UseCase) GetByID(
	ctx context.Context,
	organizationID string,
	workerID string,
) (*entity.GarmentWorker, error) {
	var err error

	organizationID, err = validateOrganizationID(organizationID)
	if err != nil {
		return nil, err
	}

	workerID = strings.TrimSpace(workerID)
	if workerID == "" {
		return nil, errors.New("garment worker id is required")
	}

	return u.repository.GetByID(ctx, organizationID, workerID)
}

func (u *UseCase) List(
	ctx context.Context,
	organizationID string,
) ([]*entity.GarmentWorker, error) {
	organizationID, err := validateOrganizationID(organizationID)
	if err != nil {
		return nil, err
	}

	return u.repository.List(ctx, organizationID)
}

func (u *UseCase) Update(
	ctx context.Context,
	organizationID string,
	worker *entity.GarmentWorker,
) error {
	var err error

	organizationID, err = validateOrganizationID(organizationID)
	if err != nil {
		return err
	}

	if worker == nil {
		return errors.New("garment worker is required")
	}

	worker.ID = strings.TrimSpace(worker.ID)

	if worker.ID == "" {
		return errors.New("garment worker id is required")
	}

	if err := validateWorker(worker); err != nil {
		return err
	}

	worker.OrganizationID = organizationID

	u.logger.InfoContext(
		ctx,
		"updating garment worker",
		"organization_id", organizationID,
		"id", worker.ID,
	)

	return u.repository.Update(ctx, organizationID, worker)
}

func (u *UseCase) Delete(
	ctx context.Context,
	organizationID string,
	workerID string,
) error {
	var err error

	organizationID, err = validateOrganizationID(organizationID)
	if err != nil {
		return err
	}

	workerID = strings.TrimSpace(workerID)
	if workerID == "" {
		return errors.New("garment worker id is required")
	}

	u.logger.InfoContext(
		ctx,
		"deleting garment worker",
		"organization_id", organizationID,
		"id", workerID,
	)

	return u.repository.Delete(ctx, organizationID, workerID)
}
