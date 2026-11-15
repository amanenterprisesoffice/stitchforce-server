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
	"stitchforce-server/internal/authorization"
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
	sessions    repo.UserSessionRepository
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
	sessions repo.UserSessionRepository,
	logger *logger.Logger,
	storage storage.Storage,
	emailer email.Sender,
	frontendCfg config.FrontendConfig,
) usecase.UserUseCase {
	return &UseCase{
		repository:  repository,
		sessions:    sessions,
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
	deviceName string,
	ipAddress string,
	userAgent string,
) (*output.User, *entity.UserSession, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	if email == "" || password == "" {
		return nil, nil, errors.New("email and password are required")
	}

	user, err := u.repository.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, nil, errors.New("no user found with the provided email")
		}
		return nil, nil, err
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(password),
	); err != nil {
		return nil, nil, errors.New("invalid email or password")
	}

	if user.EmailVerifiedAt == nil {
		if err := u.SendEmailVerification(ctx, user); err != nil {
			u.logger.Error("failed to resend email verification", "error", err)
			return nil, nil, errors.New("email is not verified and verification email could not be sent")
		}

		return nil, nil, errors.New("email is not verified. A new verification link has been sent to your email")
	}

	newSessionID, err := utils.NewID()
	if err != nil {
		return nil, nil, errors.New("failed to generate session ID")
	}

	session := &entity.UserSession{
		ID:             newSessionID,
		UserID:         user.ID,
		OrganizationID: user.OrganizationID,
		DeviceName:     deviceName,
		IPAddress:      ipAddress,
		UserAgent:      userAgent,
		ExpiresAt:      time.Now().Add(30 * 24 * time.Hour),
	}

	if err := u.sessions.CreateWithLimit(
		ctx,
		session,
		user.MaxSessions,
	); err != nil {
		u.logger.Error("failed to create user session", "error", err)
		return nil, nil, errors.New("failed to create login session")
	}

	userOutput, err := u.toOutput(ctx, user)
	if err != nil {
		_ = u.sessions.Revoke(ctx, session.ID)
		return nil, nil, err
	}

	return userOutput, session, nil
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

	if input.Role != nil {
		if !input.Role.IsValid() {
			return errors.New("invalid user role")
		}

		if existing.Role == entity.UserRoleOrgAdmin {
			return errors.New("organization admin role cannot be changed")
		}

		if *input.Role == entity.UserRoleOrgAdmin {
			return errors.New("cannot assign organization admin role through user update")
		}

		existing.Role = *input.Role
	}

	// Validate the final user state.
	if len(existing.Name) < 2 || len(existing.Name) > 100 {
		return errors.New("name must be between 2 and 100 characters")
	}

	if existing.Email == "" {
		return errors.New("email is required")
	}

	if !existing.Role.IsValid() {
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

func (u *UseCase) DeleteMe(
	ctx context.Context,
	userID string,
	password string,
) error {
	user, err := u.repository.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.Role == entity.UserRoleOrgAdmin {
		return errors.New(
			"organization admins cannot delete their account; delete the organization first",
		)
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(password),
	); err != nil {
		return errors.New("invalid password")
	}

	return u.deleteUser(ctx, user)
}

func (u *UseCase) Delete(
	ctx context.Context,
	userID string,
) error {
	user, err := u.repository.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.Role == entity.UserRoleOrgAdmin {
		return errors.New("organization admin accounts cannot be deleted")
	}

	return u.deleteUser(ctx, user)
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

const passwordResetTokenExpiry = 30 * time.Minute

func (u *UseCase) ForgotPassword(
	ctx context.Context,
	emailAddress string,
) error {
	emailAddress = strings.ToLower(strings.TrimSpace(emailAddress))

	if emailAddress == "" {
		return errors.New("email is required")
	}

	user, err := u.repository.GetByEmail(ctx, emailAddress)
	if err != nil {
		// Return the same response for unknown emails.
		// Log unexpected repository errors in your project logger.
		return nil
	}

	if user.DeletedAt != nil {
		return nil
	}

	rawToken, err := utils.GenerateToken()
	if err != nil {
		return err
	}

	tokenHash := utils.HashToken(rawToken)
	expiresAt := time.Now().Add(passwordResetTokenExpiry)

	if err := u.repository.CreatePasswordResetToken(
		ctx,
		user.ID,
		tokenHash,
		expiresAt,
	); err != nil {
		return err
	}

	resetURL := strings.TrimRight(u.frontendCfg.URL, "/") +
		"/reset-password?token=" +
		url.QueryEscape(rawToken)

	body := buildPasswordResetEmail(user.Name, resetURL)

	htmlBody, err := email.Render(body)
	if err != nil {
		return err
	}

	err = u.emailer.Send(ctx, email.Message{
		To:      []string{user.Email},
		Subject: "Reset your password",
		HTML:    htmlBody,
	})

	return err
}

func (u *UseCase) ResetPassword(
	ctx context.Context,
	rawToken string,
	newPassword string,
) error {
	rawToken = strings.TrimSpace(rawToken)

	if rawToken == "" {
		return errors.New("reset token is required")
	}

	if len(newPassword) < 8 || len(newPassword) > 72 {
		return errors.New("password must be between 8 and 72 characters")
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(newPassword),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return err
	}

	user, err := u.repository.ResetPassword(
		ctx,
		utils.HashToken(rawToken),
		string(passwordHash),
	)
	if err != nil {
		return err
	}

	htmlBody, err := buildPasswordChangedEmail(user.Name)
	if err != nil {
		return err
	}

	return u.emailer.Send(ctx, email.Message{
		To:      []string{user.Email},
		Subject: "Your StitchForce password has been changed",
		HTML:    htmlBody,
	})
}

func (u *UseCase) RevokeSession(
	ctx context.Context,
	sessionID string,
) error {
	if sessionID == "" {
		return errors.New("session ID is required")
	}

	return u.sessions.Revoke(ctx, sessionID)
}

func (u *UseCase) GetActiveSession(
	ctx context.Context,
	sessionID string,
) (*entity.UserSession, error) {
	if sessionID == "" {
		return nil, errors.New("session ID is required")
	}

	return u.sessions.GetActive(ctx, sessionID)
}

// Helpers
func (u *UseCase) deleteUser(
	ctx context.Context,
	user *entity.User,
) error {
	if err := u.repository.Delete(ctx, user.ID); err != nil {
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
		Permissions:     make([]string, 0),
		EmailVerifiedAt: user.EmailVerifiedAt,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
	}

	for _, permission := range authorization.GetPermissions(user.Role) {
		result.Permissions = append(result.Permissions, string(permission))
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

func buildPasswordResetEmail(name, resetURL string) string {
	safeName := html.EscapeString(name)
	safeURL := html.EscapeString(resetURL)

	body := fmt.Sprintf(`
		<div>
			<h2 style="margin:0 0 20px;">Reset your password</h2>

			<p>Hello %s,</p>

			<p>
				We received a request to reset the password for your StitchForce account.
				Click the button below to choose a new password.
			</p>

			<div style="margin:32px 0;">
				<a href="%s"
				   style="
						display:inline-block;
						padding:13px 24px;
						background:#18181b;
						color:#ffffff;
						text-decoration:none;
						border-radius:7px;
						font-weight:bold;
				   ">
					Reset password
				</a>
			</div>

			<p style="font-size:14px;color:#52525b;">
				This link expires in 30 minutes. If you did not request a password
				reset, you can safely ignore this email.
			</p>

			<p style="font-size:13px;color:#71717a;word-break:break-all;">
				If the button does not work, copy this link into your browser:<br>
				%s
			</p>

			<hr style="border:none;border-top:1px solid #e4e4e7;margin:28px 0;">

			<p style="font-size:12px;color:#a1a1aa;">
				This is an automated email from StitchForce. Please do not reply.
			</p>
		</div>
	`, safeName, safeURL, safeURL)

	return body
}

func buildPasswordChangedEmail(name string) (string, error) {
	safeName := html.EscapeString(name)

	body := fmt.Sprintf(`
		<div>
			<h2 style="margin:0 0 20px;">Password successfully changed</h2>

			<p>Hello %s,</p>

			<p>
				Your StitchForce account password has been successfully changed.
			</p>

			<p style="font-size:14px;color:#52525b;">
				If you made this change, no further action is required.
			</p>

			<p style="font-size:14px;color:#52525b;">
				If you did not change your password, please contact our support
				team immediately to secure your account.
			</p>

			<hr style="border:none;border-top:1px solid #e4e4e7;margin:28px 0;">

			<p style="font-size:12px;color:#a1a1aa;">
				This is an automated email from StitchForce. Please do not reply.
			</p>
		</div>
	`,
		safeName,
	)

	return email.Render(body)
}
