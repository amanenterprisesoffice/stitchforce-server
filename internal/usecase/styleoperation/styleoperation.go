package styleoperation

import (
	"context"
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

const (
	MaxPaginationLimit = 600
)

type UseCase struct {
	repository repo.StyleOperationRepository
	logger     *logger.Logger
}

func New(
	repository repo.StyleOperationRepository,
	logger *logger.Logger,
) usecase.StyleOperationUseCase {
	return &UseCase{
		repository: repository,
		logger:     logger,
	}
}

func (u *UseCase) Create(
	ctx context.Context,
	organizationID string,
	styleID string,
	operation *entity.StyleOperation,
) error {
	var err error

	organizationID, err = ucutils.ValidateOrganizationID(organizationID)
	if err != nil {
		return err
	}

	styleID, err = ucutils.ValidateStyleID(styleID)
	if err != nil {
		return err
	}

	if err := ucutils.ValidateStyleOperation(operation); err != nil {
		return err
	}

	operation.ID, err = utils.NewID()
	if err != nil {
		return err
	}

	operation.OrganizationID = organizationID
	operation.StyleID = styleID

	u.logger.InfoContext(
		ctx,
		"creating style operation",
		"id", operation.ID,
		"organization_id", organizationID,
		"style_id", styleID,
		"name", operation.Name,
	)

	return u.repository.Create(
		ctx,
		organizationID,
		styleID,
		operation,
	)
}

func (u *UseCase) GetByID(
	ctx context.Context,
	organizationID string,
	styleID string,
	operationID string,
) (*entity.StyleOperation, error) {
	var err error

	organizationID, err = ucutils.ValidateOrganizationID(organizationID)
	if err != nil {
		return nil, err
	}

	styleID, err = ucutils.ValidateStyleID(styleID)
	if err != nil {
		return nil, err
	}

	operationID = strings.TrimSpace(operationID)
	if operationID == "" {
		return nil, errors.New("style operation id is required")
	}

	return u.repository.GetByID(
		ctx,
		organizationID,
		styleID,
		operationID,
	)
}

func (u *UseCase) List(
	ctx context.Context,
	organizationID string,
	styleID string,
	filter input.StyleOperationFilter,
) ([]*entity.StyleOperation, error) {
	var err error

	organizationID, err = ucutils.ValidateOrganizationID(organizationID)
	if err != nil {
		return nil, err
	}

	styleID, err = ucutils.ValidateStyleID(styleID)
	if err != nil {
		return nil, err
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

	return u.repository.List(ctx, organizationID, styleID, filter)
}

func (u *UseCase) Update(
	ctx context.Context,
	organizationID string,
	styleID string,
	operation *entity.StyleOperation,
) error {
	var err error

	organizationID, err = ucutils.ValidateOrganizationID(organizationID)
	if err != nil {
		return err
	}

	styleID, err = ucutils.ValidateStyleID(styleID)
	if err != nil {
		return err
	}

	if operation == nil {
		return errors.New("style operation is required")
	}

	operation.ID = strings.TrimSpace(operation.ID)

	if operation.ID == "" {
		return errors.New("style operation id is required")
	}

	if err := ucutils.ValidateStyleOperation(operation); err != nil {
		return err
	}

	operation.OrganizationID = organizationID
	operation.StyleID = styleID

	u.logger.InfoContext(
		ctx,
		"updating style operation",
		"organization_id", organizationID,
		"style_id", styleID,
		"id", operation.ID,
		"name", operation.Name,
	)

	return u.repository.Update(
		ctx,
		organizationID,
		styleID,
		operation,
	)
}

func (u *UseCase) Delete(
	ctx context.Context,
	organizationID string,
	styleID string,
	operationID string,
) error {
	var err error

	organizationID, err = ucutils.ValidateOrganizationID(organizationID)
	if err != nil {
		return err
	}

	styleID, err = ucutils.ValidateStyleID(styleID)
	if err != nil {
		return err
	}

	operationID = strings.TrimSpace(operationID)
	if operationID == "" {
		return errors.New("style operation id is required")
	}

	u.logger.InfoContext(
		ctx,
		"deleting style operation",
		"organization_id", organizationID,
		"style_id", styleID,
		"id", operationID,
	)

	return u.repository.Delete(
		ctx,
		organizationID,
		styleID,
		operationID,
	)
}
