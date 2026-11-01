package styleoperation

import (
	"context"
	"errors"
	"strings"

	"stitchforce-server/internal/entity"
	"stitchforce-server/internal/repo"
	"stitchforce-server/internal/usecase"
	"stitchforce-server/pkg/logger"
	"stitchforce-server/pkg/utils"
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

func validateOrganizationID(organizationID string) (string, error) {
	organizationID = strings.TrimSpace(organizationID)

	if organizationID == "" {
		return "", errors.New("organization id is required")
	}

	return organizationID, nil
}

func validateStyleID(styleID string) (string, error) {
	styleID = strings.TrimSpace(styleID)

	if styleID == "" {
		return "", errors.New("style id is required")
	}

	return styleID, nil
}

func validateStyleOperation(operation *entity.StyleOperation) error {
	if operation == nil {
		return errors.New("style operation is required")
	}

	operation.Name = strings.TrimSpace(operation.Name)

	if operation.Name == "" {
		return errors.New("operation name is required")
	}

	if operation.RatePerPiece < 0 {
		return errors.New("rate per piece cannot be negative")
	}

	return nil
}

func (u *UseCase) Create(
	ctx context.Context,
	organizationID string,
	styleID string,
	operation *entity.StyleOperation,
) error {
	var err error

	organizationID, err = validateOrganizationID(organizationID)
	if err != nil {
		return err
	}

	styleID, err = validateStyleID(styleID)
	if err != nil {
		return err
	}

	if err := validateStyleOperation(operation); err != nil {
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
		"id", operation.ID,
		"action", "creating style operation",
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

	organizationID, err = validateOrganizationID(organizationID)
	if err != nil {
		return nil, err
	}

	styleID, err = validateStyleID(styleID)
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
) ([]*entity.StyleOperation, error) {
	var err error

	organizationID, err = validateOrganizationID(organizationID)
	if err != nil {
		return nil, err
	}

	styleID, err = validateStyleID(styleID)
	if err != nil {
		return nil, err
	}

	return u.repository.List(ctx, organizationID, styleID)
}

func (u *UseCase) Update(
	ctx context.Context,
	organizationID string,
	styleID string,
	operation *entity.StyleOperation,
) error {
	var err error

	organizationID, err = validateOrganizationID(organizationID)
	if err != nil {
		return err
	}

	styleID, err = validateStyleID(styleID)
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

	if err := validateStyleOperation(operation); err != nil {
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

	organizationID, err = validateOrganizationID(organizationID)
	if err != nil {
		return err
	}

	styleID, err = validateStyleID(styleID)
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
