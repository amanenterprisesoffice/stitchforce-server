package bundle

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"stitchforce-server/internal/entity"
	"stitchforce-server/internal/repo"
	"stitchforce-server/internal/usecase"
	"stitchforce-server/internal/usecase/input"
	ucutils "stitchforce-server/internal/usecase/utils"
	"stitchforce-server/pkg/logger"
	"stitchforce-server/pkg/utils"
)

type UseCase struct {
	repository repo.BundleRepository
	logger     *logger.Logger
}

func New(
	repository repo.BundleRepository,
	logger *logger.Logger,
) usecase.BundleUseCase {
	return &UseCase{
		repository: repository,
		logger:     logger,
	}
}

// CREATE

func (u *UseCase) Create(
	ctx context.Context,
	organizationID string,
	bundle *entity.Bundle,
) error {
	var err error

	organizationID, err = ucutils.ValidateOrganizationID(organizationID)
	if err != nil {
		return err
	}

	if err := ucutils.ValidateBundle(bundle); err != nil {
		return err
	}

	bundle.ID, err = utils.NewID()
	if err != nil {
		return err
	}

	bundle.OrganizationID = organizationID

	u.logger.InfoContext(
		ctx,
		"creating bundle",
		"id", bundle.ID,
		"action", "creating bundle",
		"organization_id", organizationID,
		"bundle_number", bundle.BundleNumber,
		"style_id", bundle.StyleID,
		"size_id", bundle.SizeID,
		"quantity", bundle.Quantity,
	)

	return u.repository.Create(
		ctx,
		organizationID,
		bundle,
	)
}

// GET BY ID

func (u *UseCase) GetByID(
	ctx context.Context,
	organizationID string,
	bundleID string,
) (*entity.Bundle, error) {
	var err error

	organizationID, err = ucutils.ValidateOrganizationID(organizationID)
	if err != nil {
		return nil, err
	}

	bundleID, err = ucutils.ValidateBundleID(bundleID)
	if err != nil {
		return nil, err
	}

	return u.repository.GetByID(
		ctx,
		organizationID,
		bundleID,
	)
}

// LIST

func (u *UseCase) List(
	ctx context.Context,
	organizationID string,
) ([]*entity.Bundle, error) {
	organizationID, err := ucutils.ValidateOrganizationID(organizationID)
	if err != nil {
		return nil, err
	}

	return u.repository.List(
		ctx,
		organizationID,
	)
}

// UPDATE
func (u *UseCase) Update(
	ctx context.Context,
	organizationID string,
	bundleID string,
	update *input.UpdateBundleInput,
) error {
	var err error

	organizationID, err = ucutils.ValidateOrganizationID(organizationID)
	if err != nil {
		return err
	}

	bundleID, err = ucutils.ValidateBundleID(bundleID)
	if err != nil {
		return err
	}

	if update == nil {
		return errors.New("bundle update is required")
	}

	// Fetch existing bundle.
	bundle, err := u.repository.GetByID(ctx, organizationID, bundleID)
	if err != nil {
		return err
	}

	if update.BundleNumber != nil {
		value := strings.TrimSpace(*update.BundleNumber)
		if value == "" || len(value) > 100 {
			// log invalid bundle number
			u.logger.ErrorContext(
				ctx,
				"invalid bundle number",
				"bundle_id", bundleID,
				"bundle_number", value,
				"update_bundle_number", update.BundleNumber,
			)
			return errors.New("invalid bundle number in usecase")
		}
		bundle.BundleNumber = value
	}

	if update.SizeID != nil {
		value := strings.TrimSpace(*update.SizeID)
		if value == "" {
			return errors.New("invalid size id")
		}
		bundle.SizeID = value
	}

	if update.Quantity != nil {
		if *update.Quantity <= 0 {
			return errors.New("quantity must be greater than zero")
		}
		bundle.Quantity = *update.Quantity
	}

	if update.Attributes != nil {
		if !json.Valid(*update.Attributes) {
			return errors.New("invalid attributes")
		}
		bundle.Attributes = *update.Attributes
	}

	if update.Remarks != nil {
		value := strings.TrimSpace(*update.Remarks)
		if len(value) > 1000 {
			return errors.New("remarks too long")
		}
		bundle.Remarks = value
	}

	if update.Status != nil {
		statusOrder := map[entity.BundleStatus]int{
			entity.BundleStatusCreated:    1,
			entity.BundleStatusReady:      2,
			entity.BundleStatusInProgress: 3,
			entity.BundleStatusCompleted:  4,
		}

		currentOrder, currentValid := statusOrder[bundle.Status]
		newOrder, newValid := statusOrder[*update.Status]

		if !currentValid || !newValid {
			return errors.New("invalid status update")
		}

		if newOrder < currentOrder {
			return errors.New("status cannot be moved backward")
		}

		bundle.Status = *update.Status
	}

	bundle.OrganizationID = organizationID

	u.logger.InfoContext(
		ctx,
		"updating bundle",
		"id", bundle.ID,
		"action", "updating bundle",
		"organization_id", organizationID,
		"bundle_number", bundle.BundleNumber,
		"size_id", bundle.SizeID,
		"quantity", bundle.Quantity,
		"status", bundle.Status,
		"remarks", bundle.Remarks,
		"attributes", bundle.Attributes,
	)

	return u.repository.Update(ctx, organizationID, bundle.ID, update)
}

// DELETE

func (u *UseCase) Delete(
	ctx context.Context,
	organizationID string,
	bundleID string,
) error {
	var err error

	organizationID, err = ucutils.ValidateOrganizationID(organizationID)
	if err != nil {
		return err
	}

	bundleID, err = ucutils.ValidateBundleID(bundleID)
	if err != nil {
		return err
	}

	u.logger.InfoContext(
		ctx,
		"deleting bundle",
		"action", "deleting bundle",
		"organization_id", organizationID,
		"bundle_id", bundleID,
	)

	return u.repository.Delete(
		ctx,
		organizationID,
		bundleID,
	)
}
