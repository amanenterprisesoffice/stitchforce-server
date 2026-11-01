package usecase

import (
	"context"
	"errors"
	ucutils "stitchforce-server/internal/usecase/utils"

	"stitchforce-server/internal/entity"
	"stitchforce-server/internal/repo"
	"stitchforce-server/internal/usecase"
	"stitchforce-server/pkg/logger"
	"stitchforce-server/pkg/utils"
)

type UseCase struct {
	repository repo.SizeRepository
	logger     *logger.Logger
}

func New(
	repository repo.SizeRepository,
	logger *logger.Logger,
) usecase.SizeUseCase {
	return &UseCase{
		repository: repository,
		logger:     logger,
	}
}

// CREATE

func (u *UseCase) Create(
	ctx context.Context,
	organizationID string,
	size *entity.Size,
) error {
	var err error

	organizationID, err = ucutils.ValidateOrganizationID(organizationID)
	if err != nil {
		return err
	}

	if err := ucutils.ValidateSize(size); err != nil {
		return err
	}

	size.ID, err = utils.NewID()
	if err != nil {
		return err
	}

	size.OrganizationID = organizationID

	u.logger.InfoContext(
		ctx,
		"creating size",
		"id", size.ID,
		"organization_id", organizationID,
		"name", size.Name,
	)

	return u.repository.Create(
		ctx,
		organizationID,
		size,
	)
}

// GET BY ID

func (u *UseCase) GetByID(
	ctx context.Context,
	organizationID string,
	sizeID string,
) (*entity.Size, error) {
	var err error

	organizationID, err = ucutils.ValidateOrganizationID(organizationID)
	if err != nil {
		return nil, err
	}

	sizeID, err = ucutils.ValidateSizeID(sizeID)
	if err != nil {
		return nil, err
	}

	return u.repository.GetByID(
		ctx,
		organizationID,
		sizeID,
	)
}

// LIST

func (u *UseCase) List(
	ctx context.Context,
	organizationID string,
) ([]*entity.Size, error) {
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
	size *entity.Size,
) error {
	var err error

	organizationID, err = ucutils.ValidateOrganizationID(organizationID)
	if err != nil {
		return err
	}

	if size == nil {
		return errors.New("size is nil")
	}

	size.ID, err = ucutils.ValidateSizeID(size.ID)
	if err != nil {
		return err
	}

	if err := ucutils.ValidateSize(size); err != nil {
		return err
	}

	size.OrganizationID = organizationID

	u.logger.InfoContext(
		ctx,
		"updating size",
		"id", size.ID,
		"organization_id", organizationID,
		"name", size.Name,
	)

	return u.repository.Update(
		ctx,
		organizationID,
		size,
	)
}

// DELETE

func (u *UseCase) Delete(
	ctx context.Context,
	organizationID string,
	sizeID string,
) error {
	var err error

	organizationID, err = ucutils.ValidateOrganizationID(organizationID)
	if err != nil {
		return err
	}

	sizeID, err = ucutils.ValidateSizeID(sizeID)
	if err != nil {
		return err
	}

	u.logger.InfoContext(
		ctx,
		"deleting size",
		"organization_id", organizationID,
		"size_id", sizeID,
	)

	return u.repository.Delete(
		ctx,
		organizationID,
		sizeID,
	)
}
