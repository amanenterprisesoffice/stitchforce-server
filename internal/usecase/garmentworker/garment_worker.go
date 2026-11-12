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
	MaxPhotoSize       = 5 * 1024 * 1024
	MaxPaginationLimit = 600
)

type UseCase struct {
	repository repo.GarmentWorkerRepository
	storage    storage.Storage
	logger     *logger.Logger
}

func New(
	repository repo.GarmentWorkerRepository,
	storage storage.Storage,
	logger *logger.Logger,
) usecase.GarmentWorkerUseCase {
	return &UseCase{
		repository: repository,
		storage:    storage,
		logger:     logger,
	}
}

func (u *UseCase) Create(
	ctx context.Context,
	organizationID string,
	input *input.CreateGarmentWorkerInput,
) error {
	ok := ucutils.ValidateID(organizationID)
	if !ok {
		return errors.New("invalid organization ID")
	}

	if input == nil {
		return errors.New("garment worker input is required")
	}

	id, err := utils.NewID()
	if err != nil {
		return err
	}

	worker := &entity.GarmentWorker{
		ID:             id,
		OrganizationID: organizationID,
		Name:           input.Name,
		Phone:          input.Phone,
		Gender:         input.Gender,
		Status:         entity.WorkerStatusActive,
		PhotoKey:       "",
	}

	if err := ucutils.ValidateWorker(worker); err != nil {
		return err
	}

	// Validate photo before performing any operations.
	if input.Photo != nil {
		if input.Photo.Size <= 0 {
			return errors.New("photo must not be empty")
		}

		if input.Photo.Size > MaxPhotoSize {
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

	var photoKey string

	// Upload photo before creating the database record.
	if input.Photo != nil {
		photoID, err := utils.NewID()
		if err != nil {
			return fmt.Errorf("generate photo ID: %w", err)
		}

		photoKey = fmt.Sprintf(
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
			return fmt.Errorf("upload worker photo: %w", err)
		}

		worker.PhotoKey = photoKey
	}

	// Create worker with photo key already populated.
	if err := u.repository.Create(ctx, organizationID, worker); err != nil {
		// Cleanup uploaded photo if database creation fails.
		if photoKey != "" {
			if deleteErr := u.storage.Delete(ctx, photoKey); deleteErr != nil {
				u.logger.ErrorContext(
					ctx,
					"failed to cleanup uploaded worker photo",
					"photo_key", photoKey,
					"error", deleteErr,
				)
			}
		}

		return fmt.Errorf("create garment worker: %w", err)
	}

	return nil
}

func (u *UseCase) GetByID(
	ctx context.Context,
	organizationID string,
	workerID string,
) (*output.GarmentWorker, error) {

	ok := ucutils.ValidateID(organizationID)
	if !ok {
		return nil, errors.New("invalid organization ID")
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
	filter input.GarmentWorkerFilter,
) ([]*output.GarmentWorker, error) {

	ok := ucutils.ValidateID(organizationID)
	if !ok {
		return nil, errors.New("invalid organization ID")
	}

	// if status exist then validate
	if filter.Status != "" {
		if !filter.Status.IsValid() {
			return nil, errors.New("invalid worker status")
		}
	}

	if filter.Limit <= 0 {
		filter.Limit = 20
	}

	if filter.Limit > MaxPaginationLimit {
		filter.Limit = MaxPaginationLimit
	}

	if filter.Offset < 0 {
		filter.Offset = 0
	}

	workers, err := u.repository.List(ctx, organizationID, filter)
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
	workerID string,
	input *input.UpdateGarmentWorkerInput,
) error {
	if !ucutils.ValidateID(organizationID) {
		return errors.New("invalid organization ID")
	}

	workerID = strings.TrimSpace(workerID)
	if workerID == "" {
		return errors.New("garment worker id is required")
	}

	if input == nil {
		return errors.New("garment worker input is required")
	}

	existing, err := u.repository.GetByID(ctx, organizationID, workerID)
	if err != nil {
		return err
	}

	oldPhotoKey := existing.PhotoKey

	// Apply only provided fields.
	if name := strings.TrimSpace(input.Name); name != "" {
		existing.Name = name
	}

	if phone := strings.TrimSpace(input.Phone); phone != "" {
		existing.Phone = phone
	}

	if gender := strings.TrimSpace(input.Gender); gender != "" {
		existing.Gender = gender
	}

	if input.Status != "" {
		if !input.Status.IsValid() {
			return errors.New("invalid worker status")
		}
		existing.Status = input.Status
	}

	var newPhotoKey string

	if input.Photo != nil {
		if input.Photo.Size <= 0 {
			return errors.New("photo must not be empty")
		}

		if input.Photo.Size > MaxPhotoSize {
			return errors.New("photo must not exceed 5MB")
		}

		switch input.Photo.ContentType {
		case "image/jpeg", "image/png", "image/webp":
		default:
			return errors.New("unsupported photo format")
		}

		photoID, err := utils.NewID()
		if err != nil {
			return err
		}

		newPhotoKey = fmt.Sprintf(
			"organizations/%s/workers/%s/profile/%s",
			organizationID,
			workerID,
			photoID,
		)

		if err := u.storage.Upload(
			ctx,
			newPhotoKey,
			input.Photo.Reader,
			input.Photo.Size,
			input.Photo.ContentType,
		); err != nil {
			return fmt.Errorf("upload worker photo: %w", err)
		}

		existing.PhotoKey = newPhotoKey
	}

	if err := ucutils.ValidateWorker(existing); err != nil {
		if newPhotoKey != "" {
			_ = u.storage.Delete(ctx, newPhotoKey)
		}
		return err
	}

	u.logger.InfoContext(
		ctx,
		"updating garment worker",
		"organization_id", organizationID,
		"worker_id", workerID,
	)

	if err := u.repository.Update(ctx, organizationID, existing); err != nil {
		if newPhotoKey != "" {
			if deleteErr := u.storage.Delete(ctx, newPhotoKey); deleteErr != nil {
				u.logger.ErrorContext(ctx, "failed to cleanup uploaded worker photo",
					"photo_key", newPhotoKey,
					"error", deleteErr,
				)
			}
		}
		return err
	}

	// Delete old photo only after successful database update.
	if newPhotoKey != "" && oldPhotoKey != "" {
		if err := u.storage.Delete(ctx, oldPhotoKey); err != nil {
			u.logger.ErrorContext(ctx, "failed to delete previous worker photo",
				"photo_key", oldPhotoKey,
				"error", err,
			)
		}
	}

	return nil
}

// SOFT DELETE
func (u *UseCase) Delete(
	ctx context.Context,
	organizationID string,
	workerID string,
) error {

	ok := ucutils.ValidateID(organizationID)
	if !ok {
		return errors.New("invalid organization ID")
	}

	workerID = strings.TrimSpace(workerID)

	if workerID == "" {
		return errors.New("garment worker id is required")
	}

	// Verify worker exists and belongs to organization.
	_, err := u.repository.GetByID(
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
