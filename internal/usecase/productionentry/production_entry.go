package productionentry

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"stitchforce-server/internal/entity"
	"stitchforce-server/internal/repo"
	"stitchforce-server/internal/usecase"
	"stitchforce-server/internal/usecase/output"
	"stitchforce-server/internal/usecase/utils"
	"stitchforce-server/pkg/logger"
	appUtils "stitchforce-server/pkg/utils"
)

type UseCase struct {
	repository               repo.ProductionEntryRepository
	bundleRepository         repo.BundleRepository
	workerRepository         repo.GarmentWorkerRepository
	styleOperationRepository repo.StyleOperationRepository
	styleRepository          repo.StyleRepository
	logger                   *logger.Logger
}

func New(
	repository repo.ProductionEntryRepository,
	bundleRepository repo.BundleRepository,
	workerRepository repo.GarmentWorkerRepository,
	styleOperationRepository repo.StyleOperationRepository,
	styleRepository repo.StyleRepository,
	logger *logger.Logger,
) usecase.ProductionEntryUseCase {
	return &UseCase{
		repository:               repository,
		bundleRepository:         bundleRepository,
		workerRepository:         workerRepository,
		styleOperationRepository: styleOperationRepository,
		styleRepository:          styleRepository,
		logger:                   logger,
	}
}

// CREATE

func (u *UseCase) Create(
	ctx context.Context,
	organizationID string,
	entry *entity.ProductionEntry,
) error {

	organizationID, err := utils.ValidateOrganizationID(organizationID)
	if err != nil {
		return err
	}

	if entry == nil {
		return errors.New("production entry is required")
	}

	entry.BundleID = strings.TrimSpace(entry.BundleID)
	entry.WorkerID = strings.TrimSpace(entry.WorkerID)
	entry.StyleOperationID = strings.TrimSpace(entry.StyleOperationID)
	entry.Remarks = strings.TrimSpace(entry.Remarks)

	if err := utils.ValidateProductionEntry(entry); err != nil {
		return err
	}

	// Validate bundle.
	bundle, err := u.bundleRepository.GetByID(
		ctx,
		organizationID,
		entry.BundleID,
	)
	if err != nil {
		return fmt.Errorf("get production bundle: %w", err)
	}

	if bundle == nil {
		return errors.New("bundle not found")
	}

	// Validate bundle status.
	if bundle.Status != "ready" && bundle.Status != "in_progress" {
		return errors.New("production can only be recorded for ready or in-progress bundles")
	}

	// Validate worker.
	worker, err := u.workerRepository.GetByID(
		ctx,
		organizationID,
		entry.WorkerID,
	)
	if err != nil {
		return fmt.Errorf("get production worker: %w", err)
	}

	if worker == nil {
		return errors.New("garment worker not found")
	}

	if worker.Status != entity.WorkerStatusActive {
		return errors.New("inactive workers cannot record production")
	}

	// Fetch style operation.
	operation, err := u.styleOperationRepository.GetByID(
		ctx,
		organizationID,
		bundle.StyleID,
		entry.StyleOperationID,
	)
	if err != nil {
		return fmt.Errorf("get style operation: %w", err)
	}

	if operation == nil {
		return errors.New("style operation not found")
	}

	// Ensure operation belongs to the bundle's style.
	if operation.StyleID != bundle.StyleID {
		return errors.New("style operation does not belong to the bundle's style")
	}

	// Snapshot operation rate.
	// Never accept this value from the frontend.
	entry.RatePerPiece = operation.RatePerPiece

	// Generate ID.
	id, err := appUtils.NewID()
	if err != nil {
		return fmt.Errorf("generate production entry ID: %w", err)
	}

	entry.ID = id
	entry.OrganizationID = organizationID

	now := time.Now().UTC()

	entry.CreatedAt = now
	entry.UpdatedAt = now

	u.logger.InfoContext(
		ctx,
		"creating production entry",
		"organization_id", organizationID,
		"bundle_id", entry.BundleID,
		"worker_id", entry.WorkerID,
		"style_operation_id", entry.StyleOperationID,
		"quantity", entry.Quantity,
		"rate_per_piece", entry.RatePerPiece,
	)

	// Repository atomically:
	// 1. Locks the bundle.
	// 2. Verifies bundle quantity.
	// 3. Verifies worker is active.
	// 4. Verifies operation belongs to the bundle's style.
	// 5. Inserts the production entry.
	if err := u.repository.CreateWithBundleLimit(
		ctx,
		organizationID,
		entry,
	); err != nil {
		return fmt.Errorf("create production entry: %w", err)
	}

	return nil
}

// GET BY ID

func (u *UseCase) GetByID(
	ctx context.Context,
	organizationID string,
	entryID string,
) (*output.ProductionEntry, error) {

	organizationID, err := utils.ValidateOrganizationID(organizationID)
	if err != nil {
		return nil, err
	}

	entryID = strings.TrimSpace(entryID)

	if entryID == "" {
		return nil, errors.New("production entry id is required")
	}

	entry, err := u.repository.GetByID(
		ctx,
		organizationID,
		entryID,
	)
	if err != nil {
		return nil, err
	}

	if entry == nil {
		return nil, errors.New("production entry not found")
	}

	result, err := u.buildOutput(
		ctx,
		organizationID,
		entry,
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// LIST

func (u *UseCase) List(
	ctx context.Context,
	organizationID string,
) ([]*output.ProductionEntry, error) {

	organizationID, err := utils.ValidateOrganizationID(organizationID)
	if err != nil {
		return nil, err
	}

	entries, err := u.repository.List(
		ctx,
		organizationID,
	)
	if err != nil {
		return nil, err
	}

	results := make([]*output.ProductionEntry, 0, len(entries))

	for _, entry := range entries {
		result, err := u.buildOutput(
			ctx,
			organizationID,
			entry,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"build production entry output for %s: %w",
				entry.ID,
				err,
			)
		}

		results = append(results, result)
	}

	return results, nil
}

// UPDATE

// UPDATE
func (u *UseCase) Update(
	ctx context.Context,
	organizationID string,
	entry *entity.ProductionEntry,
) error {

	organizationID, err := utils.ValidateOrganizationID(organizationID)
	if err != nil {
		return err
	}

	if entry == nil {
		return errors.New("production entry is required")
	}

	// Normalize input.
	entry.ID = strings.TrimSpace(entry.ID)
	entry.BundleID = strings.TrimSpace(entry.BundleID)
	entry.WorkerID = strings.TrimSpace(entry.WorkerID)
	entry.StyleOperationID = strings.TrimSpace(entry.StyleOperationID)
	entry.Remarks = strings.TrimSpace(entry.Remarks)

	if entry.ID == "" {
		return errors.New("production entry id is required")
	}

	// Validate production entry fields.
	if err := utils.ValidateProductionEntry(entry); err != nil {
		return err
	}

	// Ensure existing entry belongs to this organization.
	existing, err := u.repository.GetByID(
		ctx,
		organizationID,
		entry.ID,
	)
	if err != nil {
		return fmt.Errorf("get existing production entry: %w", err)
	}

	if existing == nil {
		return errors.New("production entry not found")
	}

	// Validate selected bundle.
	bundle, err := u.bundleRepository.GetByID(
		ctx,
		organizationID,
		entry.BundleID,
	)
	if err != nil {
		return fmt.Errorf("get production bundle: %w", err)
	}

	if bundle == nil {
		return errors.New("bundle not found")
	}

	// Production can only be recorded for ready or in-progress bundles.
	if bundle.Status != entity.BundleStatusReady &&
		bundle.Status != entity.BundleStatusInProgress {
		return errors.New(
			"production can only be recorded for ready or in-progress bundles",
		)
	}

	// Validate selected worker.
	worker, err := u.workerRepository.GetByID(
		ctx,
		organizationID,
		entry.WorkerID,
	)
	if err != nil {
		return fmt.Errorf("get production worker: %w", err)
	}

	if worker == nil {
		return errors.New("garment worker not found")
	}

	if worker.Status != entity.WorkerStatusActive {
		return errors.New("inactive workers cannot record production")
	}

	// Validate selected style operation.
	operation, err := u.styleOperationRepository.GetByID(
		ctx,
		organizationID,
		bundle.StyleID,
		entry.StyleOperationID,
	)
	if err != nil {
		return fmt.Errorf("get style operation: %w", err)
	}

	if operation == nil {
		return errors.New("style operation not found")
	}

	// Ensure operation belongs to the selected bundle's style.
	if operation.StyleID != bundle.StyleID {
		return errors.New(
			"style operation does not belong to the bundle's style",
		)
	}

	// Never trust rate_per_piece sent by the frontend.
	// Snapshot the current operation rate.
	entry.RatePerPiece = operation.RatePerPiece

	// Preserve organization ownership.
	entry.OrganizationID = organizationID

	// Preserve original creation timestamp.
	entry.CreatedAt = existing.CreatedAt

	entry.UpdatedAt = time.Now().UTC()

	u.logger.InfoContext(
		ctx,
		"updating production entry",
		"organization_id", organizationID,
		"production_entry_id", entry.ID,
		"bundle_id", entry.BundleID,
		"worker_id", entry.WorkerID,
		"style_operation_id", entry.StyleOperationID,
		"quantity", entry.Quantity,
		"rate_per_piece", entry.RatePerPiece,
	)

	// Repository must atomically:
	// 1. Lock the selected bundle.
	// 2. Exclude the existing entry quantity from the old bundle.
	// 3. Check the quantity limit of the selected bundle.
	// 4. Update the production entry.
	if err := u.repository.UpdateWithBundleLimit(
		ctx,
		organizationID,
		entry,
	); err != nil {
		return fmt.Errorf("update production entry: %w", err)
	}

	return nil
}

// SOFT DELETE

func (u *UseCase) Delete(
	ctx context.Context,
	organizationID string,
	entryID string,
) error {

	organizationID, err := utils.ValidateOrganizationID(organizationID)
	if err != nil {
		return err
	}

	entryID = strings.TrimSpace(entryID)

	if entryID == "" {
		return errors.New("production entry id is required")
	}

	// Verify entry exists and belongs to organization.
	_, err = u.repository.GetByID(
		ctx,
		organizationID,
		entryID,
	)
	if err != nil {
		return err
	}

	// Preserve historical production records.
	// Repository must soft-delete the entry.
	if err := u.repository.Delete(
		ctx,
		organizationID,
		entryID,
	); err != nil {
		return fmt.Errorf("soft delete production entry: %w", err)
	}

	u.logger.InfoContext(
		ctx,
		"production entry soft deleted",
		"organization_id", organizationID,
		"production_entry_id", entryID,
	)

	return nil
}

// Helpers
func (u *UseCase) buildOutput(
	ctx context.Context,
	organizationID string,
	entry *entity.ProductionEntry,
) (*output.ProductionEntry, error) {

	if entry == nil {
		return nil, errors.New("production entry not found")
	}

	// Get bundle.
	bundle, err := u.bundleRepository.GetByID(
		ctx,
		organizationID,
		entry.BundleID,
	)
	if err != nil {
		return nil, fmt.Errorf("get production bundle: %w", err)
	}
	if bundle == nil {
		return nil, errors.New("production bundle not found")
	}

	// Get worker.
	worker, err := u.workerRepository.GetByID(
		ctx,
		organizationID,
		entry.WorkerID,
	)
	if err != nil {
		return nil, fmt.Errorf("get production worker: %w", err)
	}
	if worker == nil {
		return nil, errors.New("production worker not found")
	}

	// Get style.
	style, err := u.styleRepository.GetByID(
		ctx,
		organizationID,
		bundle.StyleID,
	)
	if err != nil {
		return nil, fmt.Errorf("get production style: %w", err)
	}
	if style == nil {
		return nil, errors.New("production style not found")
	}

	// Get operation.
	operation, err := u.styleOperationRepository.GetByID(
		ctx,
		organizationID,
		bundle.StyleID,
		entry.StyleOperationID,
	)
	if err != nil {
		return nil, fmt.Errorf("get production operation: %w", err)
	}
	if operation == nil {
		return nil, errors.New("production operation not found")
	}

	// Calculate earnings using the historical rate stored
	// in the production entry, not the operation's current rate.
	totalAmount := int64(entry.Quantity) * entry.RatePerPiece

	return &output.ProductionEntry{
		ID:               entry.ID,
		OrganizationID:   entry.OrganizationID,
		BundleID:         entry.BundleID,
		WorkerID:         entry.WorkerID,
		StyleOperationID: entry.StyleOperationID,

		Quantity:     entry.Quantity,
		RatePerPiece: entry.RatePerPiece,
		TotalAmount:  totalAmount,

		ProductionDate: entry.ProductionDate,
		Remarks:        entry.Remarks,

		Worker: output.ProductionEntryWorker{
			ID:    worker.ID,
			Name:  worker.Name,
			Phone: worker.Phone,
		},

		Bundle: output.ProductionEntryBundle{
			ID:           bundle.ID,
			BundleNumber: bundle.BundleNumber,
			Quantity:     bundle.Quantity,
			Status:       string(bundle.Status),
		},

		Style: output.ProductionEntryStyle{
			ID:          style.ID,
			Name:        style.Name,
			StyleNumber: style.StyleNumber,
		},

		Operation: output.ProductionEntryOperation{
			ID:           operation.ID,
			Name:         operation.Name,
			RatePerPiece: entry.RatePerPiece,
		},

		CreatedAt: entry.CreatedAt,
		UpdatedAt: entry.UpdatedAt,
	}, nil
}
