package user

import (
	"context"
	"errors"
	"strings"

	"stitchforce-server/internal/entity"
	"stitchforce-server/internal/repo"
	"stitchforce-server/internal/usecase"
	"stitchforce-server/pkg/logger"
	"stitchforce-server/pkg/utils"

	"golang.org/x/crypto/bcrypt"
)

type UseCase struct {
	repository repo.UserRepository
	logger     *logger.Logger
}

func New(
	repository repo.UserRepository,
	logger *logger.Logger,
) usecase.UserUseCase {
	return &UseCase{
		repository: repository,
		logger:     logger,
	}
}

func (u *UseCase) Register(
	ctx context.Context,
	name, email, password string,
) (*entity.User, error) {
	name = strings.TrimSpace(name)
	email = strings.ToLower(strings.TrimSpace(email))

	if len(name) < 2 || len(name) > 100 {
		return nil, errors.New("name must be between 2 and 100 characters")
	}
	if email == "" || password == "" {
		return nil, errors.New("email and password are required")
	}
	if len(password) < 8 || len(password) > 72 {
		return nil, errors.New("password must be between 8 and 72 bytes")
	}

	existing, err := u.repository.GetByEmail(ctx, email)
	if err == nil && existing != nil {
		return nil, errors.New("email already registered")
	}
	if err != nil && !errors.Is(err, repo.ErrNotFound) {
		return nil, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(password), bcrypt.DefaultCost,
	)
	if err != nil {
		return nil, err
	}

	orgID, err := utils.NewID()
	if err != nil {
		return nil, err
	}

	userID, err := utils.NewID()
	if err != nil {
		return nil, err
	}

	org := &entity.Organization{
		ID:            orgID,
		Name:          name + "'s Organization",
		LicenseStatus: "active",
	}

	user := &entity.User{
		ID:             userID,
		OrganizationID: orgID,
		Name:           name,
		Email:          email,
		Role:           "org_admin",
		PasswordHash:   string(passwordHash),
	}

	if err := u.repository.RegisterOrganization(ctx, org, user); err != nil {
		return nil, err
	}

	u.logger.InfoContext(ctx, "registered organization admin",
		"email", email,
		"organization_id", orgID,
	)

	user.PasswordHash = ""
	return user, nil
}

func (u *UseCase) Login(
	ctx context.Context,
	email, password string,
) (*entity.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	if email == "" || password == "" {
		return nil, errors.New("email and password are required")
	}

	user, err := u.repository.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, errors.New("invalid email or password")
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(password),
	); err != nil {
		return nil, errors.New("invalid email or password")
	}

	user.PasswordHash = ""
	return user, nil
}

func (u *UseCase) GetByID(
	ctx context.Context,
	id string,
) (*entity.User, error) {
	user, err := u.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	user.PasswordHash = ""
	return user, nil
}

func (u *UseCase) List(
	ctx context.Context,
) ([]*entity.User, error) {
	users, err := u.repository.List(ctx)
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		user.PasswordHash = ""
	}
	return users, nil
}

func (u *UseCase) Update(
	ctx context.Context,
	user *entity.User,
) error {
	if user == nil {
		return errors.New("user is required")
	}
	user.Name = strings.TrimSpace(user.Name)
	user.Email = strings.ToLower(strings.TrimSpace(user.Email))

	if len(user.Name) < 2 || len(user.Name) > 100 {
		return errors.New("name must be between 2 and 100 characters")
	}
	if user.Email == "" {
		return errors.New("email is required")
	}
	if user.Role != "org_admin" && user.Role != "staff" {
		return errors.New("invalid user role")
	}

	return u.repository.Update(ctx, user)
}

func (u *UseCase) Delete(
	ctx context.Context,
	id string,
) error {
	return u.repository.Delete(ctx, id)
}
