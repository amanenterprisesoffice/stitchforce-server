package style

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

type UseCase struct {
	repository repo.StyleRepository
	logger     *logger.Logger
}

func New(
	repository repo.StyleRepository,
	logger *logger.Logger,
) usecase.StyleUseCase {
	return &UseCase{
		repository: repository,
		logger:     logger,
	}
}




func (u *UseCase) Create(
	ctx context.Context,
	organizationID string,
	style *entity.Style,
) error {
	var err error

	organizationID, err = ucutils.ValidateOrganizationID(organizationID)
	if err != nil {
		return err
	}

	if err := ucutils.ValidateStyle(style); err != nil {
		return err
	}

	style.ID, err = utils.NewID()
	if err != nil {
		return err
	}
	style.OrganizationID = organizationID

	u.logger.InfoContext(
		ctx,
		"creating style",
		"organization_id", organizationID,
		"style_number", style.StyleNumber,
		"name", style.Name,
	)

	return u.repository.Create(ctx, organizationID, style)
}

func (u *UseCase) GetByID(
	ctx context.Context,
	organizationID string,
	styleID string,
) (*entity.Style, error) {
	var err error

	organizationID, err = ucutils.ValidateOrganizationID(organizationID)
	if err != nil {
		return nil, err
	}

	styleID = strings.TrimSpace(styleID)
	if styleID == "" {
		return nil, errors.New("style id is required")
	}

	return u.repository.GetByID(ctx, organizationID, styleID)
}

func (u *UseCase) List(
	ctx context.Context,
	organizationID string,
	filter input.StyleFilter,
) ([]*entity.Style, error) {
	organizationID, err := ucutils.ValidateOrganizationID(organizationID)
	if err != nil {
		return nil, err
	}

	return u.repository.List(ctx, organizationID, filter)
}

func (u *UseCase) Update(
	ctx context.Context,
	organizationID string,
	style *entity.Style,
) error {
	var err error

	organizationID, err = ucutils.ValidateOrganizationID(organizationID)
	if err != nil {
		return err
	}

	if style == nil {
		return errors.New("style is required")
	}

	style.ID = strings.TrimSpace(style.ID)

	if style.ID == "" {
		return errors.New("style id is required")
	}

	if err := ucutils.ValidateStyle(style); err != nil {
		return err
	}

	style.OrganizationID = organizationID

	u.logger.InfoContext(
		ctx,
		"updating style",
		"organization_id", organizationID,
		"id", style.ID,
		"style_number", style.StyleNumber,
	)

	return u.repository.Update(ctx, organizationID, style)
}

func (u *UseCase) Delete(
	ctx context.Context,
	organizationID string,
	styleID string,
) error {
	var err error

	organizationID, err = ucutils.ValidateOrganizationID(organizationID)
	if err != nil {
		return err
	}

	styleID = strings.TrimSpace(styleID)
	if styleID == "" {
		return errors.New("style id is required")
	}

	u.logger.InfoContext(
		ctx,
		"deleting style",
		"organization_id", organizationID,
		"id", styleID,
	)

	return u.repository.Delete(ctx, organizationID, styleID)
}
