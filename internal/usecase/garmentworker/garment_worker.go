package garmentworker

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"stitchforce-server/internal/entity"
	"stitchforce-server/internal/repo"
	"stitchforce-server/internal/storage"
	"stitchforce-server/internal/usecase"
	"stitchforce-server/internal/usecase/input"
	"stitchforce-server/internal/usecase/output"
	ucutils "stitchforce-server/internal/usecase/utils"
	"stitchforce-server/pkg/logger"
	"stitchforce-server/pkg/utils"
)

const (
	WorkerStatusActive   = "active"
	WorkerStatusInactive = "inactive"

	maxPhotoSize = 5 * 1024 * 1024
)

type UseCase struct {
	repository repo.GarmentWorkerRepository
	logger     *logger.Logger
	storage    storage.Storage
}

func New(
	repository repo.GarmentWorkerRepository,
	logger *logger.Logger,
	storage storage.Storage,
) usecase.GarmentWorkerUseCase {
	return &UseCase{
		repository: repository,
		logger:     logger,
		storage:    storage,
	}
}

func (u *UseCase) Create(
	ctx context.Context,
	organizationID string,
	input *input.CreateGarmentWorkerInput,
) error {

	organizationID, err := ucutils.ValidateOrganizationID(organizationID)
	if err != nil {
		return err
	}

	if input == nil {
		return errors.New("garment worker input is required")
	}

	worker := &entity.GarmentWorker{
		OrganizationID: organizationID,
		Name:           input.Name,
		Phone:          input.Phone,
		Gender:         input.Gender,
		Status:         WorkerStatusActive,
	}

	if err := ucutils.ValidateWorker(worker); err != nil {
		return err
	}

	// Validate photo before creating the database record.
	if input.Photo != nil {
		if input.Photo.Size <= 0 {
			return errors.New("photo must not be empty")
		}

		if input.Photo.Size > maxPhotoSize {
			return errors.New("photo must not exceed 5MB")
		}

		switch input.Photo.ContentType {
		case "image/jpeg", "image/png", "image/webp":
		default:
			return errors.New("unsupported photo format")
		}
	}

	u.logger.InfoContext(
		ctx,
		"creating garment worker",
		"organization_id", organizationID,
		"name", worker.Name,
	)

	if err := u.repository.Create(ctx, organizationID, worker); err != nil {
		return fmt.Errorf("create garment worker: %w", err)
	}

	// Profile photo is optional.
	if input.Photo == nil {
		return nil
	}

	photoID, err := utils.NewID()
	if err != nil {
		return fmt.Errorf("generate photo ID: %w", err)
	}

	photoKey := fmt.Sprintf(
		"organizations/%s/workers/%s/profile/%s",
		organizationID,
		worker.ID,
		photoID,
	)

	if err := u.storage.Upload(
		ctx,
		photoKey,
		input.Photo.Reader,
		input.Photo.Size,
		input.Photo.ContentType,
	); err != nil {

		u.rollbackWorkerCreation(ctx, organizationID, worker.ID)

		return fmt.Errorf("upload worker photo: %w", err)
	}

	worker.PhotoKey = photoKey

	if err := u.repository.Update(ctx, organizationID, worker); err != nil {

		if deleteErr := u.storage.Delete(ctx, photoKey); deleteErr != nil {
			u.logger.ErrorContext(
				ctx,
				"failed to cleanup uploaded worker photo",
				"photo_key", photoKey,
				"error", deleteErr,
			)
		}

		u.rollbackWorkerCreation(ctx, organizationID, worker.ID)

		return fmt.Errorf("save worker photo key: %w", err)
	}

	return nil
}

func (u *UseCase) GetByID(
	ctx context.Context,
	organizationID string,
	workerID string,
) (*output.GarmentWorker, error) {

	organizationID, err := ucutils.ValidateOrganizationID(organizationID)
	if err != nil {
		return nil, err
	}

	workerID = strings.TrimSpace(workerID)

	if workerID == "" {
		return nil, errors.New("garment worker id is required")
	}

	worker, err := u.repository.GetByID(
		ctx,
		organizationID,
		workerID,
	)
	if err != nil {
		return nil, err
	}

	return u.toOutput(ctx, worker)
}

func (u *UseCase) List(
	ctx context.Context,
	organizationID string,
) ([]*output.GarmentWorker, error) {

	organizationID, err := ucutils.ValidateOrganizationID(organizationID)
	if err != nil {
		return nil, err
	}

	workers, err := u.repository.List(ctx, organizationID)
	if err != nil {
		return nil, err
	}

	results := make([]*output.GarmentWorker, 0, len(workers))

	for _, worker := range workers {
		result, err := u.toOutput(ctx, worker)
		if err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	return results, nil
}

func (u *UseCase) Update(
	ctx context.Context,
	organizationID string,
	worker *entity.GarmentWorker,
) error {

	organizationID, err := ucutils.ValidateOrganizationID(organizationID)
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

	worker.OrganizationID = organizationID

	if err := ucutils.ValidateWorker(worker); err != nil {
		return err
	}

	if err := validateWorkerStatus(worker.Status); err != nil {
		return err
	}

	u.logger.InfoContext(
		ctx,
		"updating garment worker",
		"organization_id", organizationID,
		"worker_id", worker.ID,
	)

	return u.repository.Update(ctx, organizationID, worker)
}

// DEACTIVATE / REACTIVATE
// SetStatus allows administrators to activate or deactivate workers.
// Deleted workers cannot be activated through this method.
func (u *UseCase) SetStatus(
	ctx context.Context,
	organizationID string,
	workerID string,
	status string,
) error {

	organizationID, err := ucutils.ValidateOrganizationID(organizationID)
	if err != nil {
		return err
	}

	workerID = strings.TrimSpace(workerID)

	if workerID == "" {
		return errors.New("garment worker id is required")
	}

	status = strings.ToLower(strings.TrimSpace(status))

	if err := validateWorkerStatus(status); err != nil {
		return err
	}

	worker, err := u.repository.GetByID(
		ctx,
		organizationID,
		workerID,
	)
	if err != nil {
		return err
	}

	worker.Status = status

	worker.UpdatedAt = time.Now().UTC()

	if err := u.repository.Update(ctx, organizationID, worker); err != nil {
		return fmt.Errorf("update worker status: %w", err)
	}

	u.logger.InfoContext(
		ctx,
		"garment worker status updated",
		"organization_id", organizationID,
		"worker_id", workerID,
		"status", status,
	)

	return nil
}

// SOFT DELETE
func (u *UseCase) Delete(
	ctx context.Context,
	organizationID string,
	workerID string,
) error {

	organizationID, err := ucutils.ValidateOrganizationID(organizationID)
	if err != nil {
		return err
	}

	workerID = strings.TrimSpace(workerID)

	if workerID == "" {
		return errors.New("garment worker id is required")
	}

	// Verify worker exists and belongs to organization.
	_, err = u.repository.GetByID(
		ctx,
		organizationID,
		workerID,
	)
	if err != nil {
		return err
	}

	// Do NOT physically delete the worker.
	// Do NOT delete the profile photo.
	// Historical production and payment records must remain accessible.
	if err := u.repository.Delete(
		ctx,
		organizationID,
		workerID,
	); err != nil {
		return fmt.Errorf("soft delete garment worker: %w", err)
	}

	u.logger.InfoContext(
		ctx,
		"garment worker soft deleted",
		"organization_id", organizationID,
		"worker_id", workerID,
	)

	return nil
}

func validateWorkerStatus(status string) error {
	status = strings.ToLower(strings.TrimSpace(status))

	switch status {
	case WorkerStatusActive, WorkerStatusInactive:
		return nil
	default:
		return errors.New("worker status must be active or inactive")
	}
}

// ROLLBACK
func (u *UseCase) rollbackWorkerCreation(
	ctx context.Context,
	organizationID string,
	workerID string,
) {
	if err := u.repository.Delete(
		ctx,
		organizationID,
		workerID,
	); err != nil {
		u.logger.ErrorContext(
			ctx,
			"failed to rollback worker creation",
			"worker_id", workerID,
			"error", err,
		)
	}
}

// ENTITY TO OUTPUT
func (u *UseCase) toOutput(
	ctx context.Context,
	worker *entity.GarmentWorker,
) (*output.GarmentWorker, error) {

	if worker == nil {
		return nil, errors.New("garment worker not found")
	}

	result := &output.GarmentWorker{
		ID:             worker.ID,
		OrganizationID: worker.OrganizationID,
		Name:           worker.Name,
		Phone:          worker.Phone,
		Gender:         worker.Gender,
		Status:         worker.Status,
		CreatedAt:      worker.CreatedAt,
		UpdatedAt:      worker.UpdatedAt,
		DeletedAt:      worker.DeletedAt,
	}

	if strings.TrimSpace(worker.PhotoKey) != "" {
		photoURL, err := u.storage.SignedURL(
			ctx,
			worker.PhotoKey,
			15*time.Minute,
		)
		if err != nil {
			return nil, fmt.Errorf("generate worker photo URL: %w", err)
		}

		result.PhotoURL = &photoURL
	}

	return result, nil
}
