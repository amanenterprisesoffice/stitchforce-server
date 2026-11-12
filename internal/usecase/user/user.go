package user

import (
	"context"
	"errors"
	"fmt"
	"html"
	"net/url"
	"strings"
	"time"

	"stitchforce-server/config"
	"stitchforce-server/internal/entity"
	"stitchforce-server/internal/repo"
	"stitchforce-server/internal/storage"
	"stitchforce-server/internal/usecase"
	"stitchforce-server/internal/usecase/input"
	"stitchforce-server/internal/usecase/output"
	ucutils "stitchforce-server/internal/usecase/utils"
	"stitchforce-server/pkg/email"
	"stitchforce-server/pkg/logger"
	"stitchforce-server/pkg/utils"

	"golang.org/x/crypto/bcrypt"
)

type UseCase struct {
	repository  repo.UserRepository
	logger      *logger.Logger
	storage     storage.Storage
	emailer     email.Sender
	frontendCfg config.FrontendConfig
}

const (
	MaxPaginationLimit = 500
	MaxPhotoSize       = 5 * 1024 * 1024
)

func New(
	repository repo.UserRepository,
	logger *logger.Logger,
	storage storage.Storage,
	emailer email.Sender,
	frontendCfg config.FrontendConfig,
) usecase.UserUseCase {
	return &UseCase{
		repository:  repository,
		logger:      logger,
		storage:     storage,
		emailer:     emailer,
		frontendCfg: frontendCfg,
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

	trailDuration := time.Now().AddDate(0, 0, 7)
	org := &entity.Organization{
		ID:               orgID,
		Name:             name + "'s Organization",
		LicenseStatus:    "active",
		LicenseExpiresAt: &trailDuration,
	}

	user := &entity.User{
		ID:             userID,
		OrganizationID: orgID,
		Name:           name,
		Email:          email,
		Role:           "org_admin",
		PasswordHash:   string(passwordHash),
		PhotoKey:       "",
	}

	if err := u.repository.RegisterOrganization(ctx, org, user); err != nil {
		return nil, err
	}

	u.logger.InfoContext(ctx, "registered organization admin",
		"email", email,
		"organization_id", orgID,
	)

	if err := u.SendEmailVerification(ctx, user); err != nil {
		// Account has already been created.
		// Log the error; user can request another verification email.
		u.logger.Error("failed to send verification email", "error", err)
	}

	user.PasswordHash = ""
	return user, nil
}

func (u *UseCase) Login(
	ctx context.Context,
	email, password string,
) (*output.User, error) {
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

	if user.EmailVerifiedAt == nil {
		if err := u.SendEmailVerification(ctx, user); err != nil {
			u.logger.Error("failed to resend email verification", "error", err)
			return nil, errors.New("email is not verified and verification email could not be sent")
		}

		return nil, errors.New("email is not verified. A new verification link has been sent to your email")
	}

	return u.toOutput(ctx, user)
}

func (u *UseCase) GetByID(
	ctx context.Context,
	id string,
) (*output.User, error) {
	user, err := u.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return u.toOutput(ctx, user)
}

func (u *UseCase) List(
	ctx context.Context,
	organizationID string,
	filter input.UserFilter,
) ([]*output.User, error) {
	if !ucutils.ValidateID(organizationID) {
		return nil, errors.New("invalid organization ID")
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

	users, err := u.repository.List(ctx, organizationID, filter)
	if err != nil {
		return nil, err
	}

	results := make([]*output.User, 0, len(users))

	for _, user := range users {
		result, err := u.toOutput(ctx, user)
		if err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	return results, nil
}

func (u *UseCase) Update(
	ctx context.Context,
	userID string,
	input *input.UpdateUserInput,
) error {
	userID = strings.TrimSpace(userID)

	if userID == "" {
		return errors.New("user ID is required")
	}

	if input == nil {
		return errors.New("user input is required")
	}

	existing, err := u.repository.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// Apply only provided fields.
	if input.Name != nil {
		existing.Name = strings.TrimSpace(*input.Name)
	}

	if input.Email != nil {
		existing.Email = strings.ToLower(strings.TrimSpace(*input.Email))
	}


	// Validate the final user state.
	if len(existing.Name) < 2 || len(existing.Name) > 100 {
		return errors.New("name must be between 2 and 100 characters")
	}

	if existing.Email == "" {
		return errors.New("email is required")
	}

	if existing.Role != "org_admin" && existing.Role != "staff" {
		return errors.New("invalid user role")
	}

	oldPhotoKey := existing.PhotoKey
	var newPhotoKey string

	// Upload replacement photo if provided.
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
			return fmt.Errorf("generate photo ID: %w", err)
		}

		newPhotoKey = fmt.Sprintf(
			"organizations/%s/users/%s/profile/%s",
			existing.OrganizationID,
			existing.ID,
			photoID,
		)

		if err := u.storage.Upload(
			ctx,
			newPhotoKey,
			input.Photo.Reader,
			input.Photo.Size,
			input.Photo.ContentType,
		); err != nil {
			return fmt.Errorf("upload user photo: %w", err)
		}

		existing.PhotoKey = newPhotoKey
	}

	u.logger.InfoContext(
		ctx,
		"updating user",
		"user_id", userID,
		"organization_id", existing.OrganizationID,
	)

	if err := u.repository.Update(ctx, existing); err != nil {
		// Cleanup newly uploaded photo if database update fails.
		if newPhotoKey != "" {
			if deleteErr := u.storage.Delete(ctx, newPhotoKey); deleteErr != nil {
				u.logger.ErrorContext(
					ctx,
					"failed to cleanup uploaded user photo",
					"photo_key", newPhotoKey,
					"error", deleteErr,
				)
			}
		}

		return fmt.Errorf("update user: %w", err)
	}

	// Delete previous photo only after successful database update.
	if newPhotoKey != "" && oldPhotoKey != "" {
		if err := u.storage.Delete(ctx, oldPhotoKey); err != nil {
			u.logger.ErrorContext(
				ctx,
				"failed to delete previous user photo",
				"photo_key", oldPhotoKey,
				"error", err,
			)
		}
	}

	return nil
}

func (u *UseCase) Delete(
	ctx context.Context,
	id string,
) error {
	user, err := u.repository.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := u.repository.Delete(ctx, id); err != nil {
		return err
	}

	if user.PhotoKey != "" {
		if err := u.storage.Delete(ctx, user.PhotoKey); err != nil {
			u.logger.ErrorContext(
				ctx,
				"failed to delete user photo",
				"photo_key", user.PhotoKey,
				"error", err,
			)
		}
	}

	return nil
}

func (u *UseCase) SendEmailVerification(
	ctx context.Context,
	user *entity.User,
) error {
	rawToken, err := utils.GenerateToken()
	if err != nil {
		return err
	}

	tokenHash := utils.HashToken(rawToken)
	expiresAt := time.Now().Add(24 * time.Hour)

	if err := u.repository.CreateEmailVerificationToken(
		ctx,
		user.ID,
		tokenHash,
		expiresAt,
	); err != nil {
		return err
	}

	verificationURL := fmt.Sprintf(
		"%s/verify-email?token=%s",
		u.frontendCfg.URL,
		url.QueryEscape(rawToken),
	)

	err = u.emailer.Send(ctx, email.Message{
		To:      []string{user.Email},
		Subject: "Verify your email address",
		HTML: fmt.Sprintf(`
			<h2>Verify your email</h2>
			<p>Hello %s,</p>
			<p>Please click the link below to verify your email address.</p>
			<a href="%s">Verify Email</a>
			<p>This link expires in 24 hours.</p>
		`, html.EscapeString(user.Name), html.EscapeString(verificationURL)),
	})

	if err != nil {
		u.logger.Error("failed to send email verification", "error", err)
		return err
	}

	return nil
}

func (u *UseCase) VerifyEmail(
	ctx context.Context,
	rawToken string,
) error {
	if rawToken == "" {
		return errors.New("verification token is required")
	}

	tokenHash := utils.HashToken(rawToken)

	if err := u.repository.VerifyEmail(ctx, tokenHash); err != nil {
		u.logger.Error("failed to verify email", "error", err)
		return err
	}

	return nil
}

// Helpers
func (u *UseCase) toOutput(
	ctx context.Context,
	user *entity.User,
) (*output.User, error) {
	if user == nil {
		return nil, errors.New("user not found")
	}

	result := &output.User{
		ID:              user.ID,
		OrganizationID:  user.OrganizationID,
		Name:            user.Name,
		Email:           user.Email,
		Role:            user.Role,
		EmailVerifiedAt: user.EmailVerifiedAt,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
	}

	if strings.TrimSpace(user.PhotoKey) != "" {
		photoURL, err := u.storage.SignedURL(
			ctx,
			user.PhotoKey,
			15*time.Minute,
		)
		if err != nil {
			return nil, fmt.Errorf("generate user photo URL: %w", err)
		}

		result.PhotoURL = &photoURL
	}

	return result, nil
}
