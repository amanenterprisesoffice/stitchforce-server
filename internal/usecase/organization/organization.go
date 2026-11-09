package organization

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"stitchforce-server/internal/entity"
	"stitchforce-server/internal/repo"
	"stitchforce-server/internal/usecase"
	"stitchforce-server/internal/usecase/input"
	"stitchforce-server/internal/usecase/output"
	"stitchforce-server/internal/usecase/utils"
	ucutils "stitchforce-server/internal/usecase/utils"
	"stitchforce-server/pkg/logger"
)

type UseCase struct {
	repository repo.OrganizationRepository
	logger     *logger.Logger
}

func New(repository repo.OrganizationRepository, log *logger.Logger) usecase.OrganizationUseCase {
	return &UseCase{repository: repository, logger: log}
}

func validateID(id string) (string, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", errors.New("organization ID is required")
	}
	return id, nil
}

func (u *UseCase) Create(ctx context.Context, org *entity.Organization) error {
	if err := ucutils.ValidateOrganization(org); err != nil {
		return err
	}

	u.logger.InfoContext(ctx, "creating organization", "name", org.Name)

	if err := u.repository.Create(ctx, org); err != nil {
		u.logger.ErrorContext(ctx, "failed to create organization", "name", org.Name, "error", err)
		return err
	}

	u.logger.InfoContext(ctx, "organization created", "id", org.ID)
	return nil
}

func (u *UseCase) GetByID(ctx context.Context, id string) (*entity.Organization, error) {
	id, err := validateID(id)
	if err != nil {
		return nil, err
	}

	org, err := u.repository.GetByID(ctx, id)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get organization", "id", id, "error", err)
		return nil, err
	}

	return org, nil
}

func (u *UseCase) List(ctx context.Context) ([]*entity.Organization, error) {
	orgs, err := u.repository.List(ctx)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to list organizations", "error", err)
		return nil, err
	}

	return orgs, nil
}

func (u *UseCase) Update(ctx context.Context, org *entity.Organization) error {
	if org == nil {
		return errors.New("organization is required")
	}

	id, err := validateID(org.ID)
	if err != nil {
		return err
	}

	org.ID = id

	if err := ucutils.ValidateOrganization(org); err != nil {
		return err
	}

	u.logger.InfoContext(ctx, "updating organization", "id", id, "name", org.Name)

	if err := u.repository.Update(ctx, org); err != nil {
		u.logger.ErrorContext(ctx, "failed to update organization", "id", id, "error", err)
		return err
	}

	u.logger.InfoContext(ctx, "organization updated", "id", id)
	return nil
}

func (u *UseCase) Delete(ctx context.Context, id string) error {
	id, err := validateID(id)
	if err != nil {
		return err
	}

	u.logger.InfoContext(ctx, "deleting organization", "id", id)

	if err := u.repository.Delete(ctx, id); err != nil {
		u.logger.ErrorContext(ctx, "failed to delete organization", "id", id, "error", err)
		return err
	}

	u.logger.InfoContext(ctx, "organization deleted", "id", id)
	return nil
}

func (u *UseCase) GetDashboardOverview(
	ctx context.Context,
	organizationID string,
) (*output.DashboardOverview, error) {
	organizationID, err := utils.ValidateOrganizationID(organizationID)
	if err != nil {
		return nil, err
	}

	dashboard, err := u.repository.GetDashboardOverview(
		ctx,
		organizationID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"get organization dashboard: %w",
			err,
		)
	}

	return dashboard, nil
}
func (u *UseCase) GetDashboardTrend(
	ctx context.Context,
	organizationID string,
	filter input.DashboardFilter,
) (*output.OrganizationDashboard, error) {
	organizationID, err := utils.ValidateOrganizationID(organizationID)
	if err != nil {
		return nil, err
	}

	dashboard, err := u.repository.GetDashboardTrend(
		ctx,
		organizationID,
		filter,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"get organization dashboard: %w",
			err,
		)
	}

	return dashboard, nil
}

var _ usecase.OrganizationUseCase = (*UseCase)(nil)
