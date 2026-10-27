
package organization

import (
    "context"
    "errors"
    "strings"

    "stitchforce-server/internal/entity"
    "stitchforce-server/internal/repo"
)

type UseCase struct {
    repository repo.OrganizationRepository
}

func New(repository repo.OrganizationRepository) *UseCase {
    return &UseCase{repository: repository}
}

func (u *UseCase) Create(
    ctx context.Context,
    name string,
) (*entity.Organization, error) {
    name = strings.TrimSpace(name)

    if len([]rune(name)) < 2 || len([]rune(name)) > 150 {
        return nil, errors.New(
            "organization name must be between 2 and 150 characters",
        )
    }

    org := &entity.Organization{
        Name: name,
    }

    if err := u.repository.Create(ctx, org); err != nil {
        return nil, err
    }

    return org, nil
}

func (u *UseCase) GetByID(
    ctx context.Context,
    id string,
) (*entity.Organization, error) {
    if strings.TrimSpace(id) == "" {
        return nil, errors.New("organization ID is required")
    }

    return u.repository.GetByID(ctx, id)
}

func (u *UseCase) List(
    ctx context.Context,
) ([]*entity.Organization, error) {
    return u.repository.List(ctx)
}

func (u *UseCase) Update(
    ctx context.Context,
    id string,
    name string,
) (*entity.Organization, error) {
    id = strings.TrimSpace(id)
    name = strings.TrimSpace(name)

    if id == "" {
        return nil, errors.New("organization ID is required")
    }

    if len([]rune(name)) < 2 || len([]rune(name)) > 150 {
        return nil, errors.New(
            "organization name must be between 2 and 150 characters",
        )
    }

    org, err := u.repository.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }

    org.Name = name

    if err := u.repository.Update(ctx, org); err != nil {
        return nil, err
    }

    return u.repository.GetByID(ctx, id)
}

func (u *UseCase) Delete(
    ctx context.Context,
    id string,
) error {
    if strings.TrimSpace(id) == "" {
        return errors.New("organization ID is required")
    }

    return u.repository.Delete(ctx, id)
}
