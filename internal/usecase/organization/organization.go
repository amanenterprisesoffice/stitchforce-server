package organization

import (
	"context"
	"database/sql"
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
	repository  repo.OrganizationRepository
	invitations repo.OrganizationInvitationRepository
	users       repo.UserRepository
	emailer     email.Sender
	storage     storage.Storage
	frontendCfg config.FrontendConfig
	logger      *logger.Logger
}

func New(repository repo.OrganizationRepository,
	invitations repo.OrganizationInvitationRepository,
	users repo.UserRepository,
	storage storage.Storage,
	emailer email.Sender,
	frontendCfg config.FrontendConfig,
	log *logger.Logger) usecase.OrganizationUseCase {
	return &UseCase{repository: repository,
		invitations: invitations,
		users:       users,
		emailer:     emailer,
		storage:     storage,
		frontendCfg: frontendCfg,
		logger:      log}
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

func (u *UseCase) GetByUserAndOrganization(
	ctx context.Context,
	userID string,
	organizationID string,
) (*output.Organization, error) {
	if !ucutils.ValidateID(userID) {
		return nil, errors.New("invalid user ID")
	}

	if !ucutils.ValidateID(organizationID) {
		return nil, errors.New("invalid organization ID")
	}

	org, err := u.repository.GetByUserAndOrganization(ctx, userID, organizationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		u.logger.ErrorContext(ctx, "failed to verify organization access",
			"user_id", userID,
			"organization_id", organizationID,
			"error", err,
		)
		return nil, err
	}

	return u.toOutput(ctx, org)
}

func (u *UseCase) GetByID(
	ctx context.Context,
	id string,
) (*output.Organization, error) {
	if !ucutils.ValidateID(id) {
		return nil, errors.New("invalid ID")
	}

	org, err := u.repository.GetByID(ctx, id)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get organization", "id", id, "error", err)
		return nil, err
	}

	return u.toOutput(ctx, org)
}

func (u *UseCase) List(ctx context.Context) ([]*output.Organization, error) {
	orgs, err := u.repository.List(ctx)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to list organizations", "error", err)
		return nil, err
	}

	result := make([]*output.Organization, 0, len(orgs))

	for _, org := range orgs {
		item, err := u.toOutput(ctx, org)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}

	return result, nil
}

func (u *UseCase) Update(
	ctx context.Context,
	id string,
	input *input.UpdateOrganizationInput,
) error {
	if !ucutils.ValidateID(id) {
		return errors.New("invalid organization ID")
	}
	if input == nil {
		return errors.New("update input is required")
	}

	org, err := u.repository.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if input.Name != nil {
		org.Name = strings.TrimSpace(*input.Name)
	}
	if input.AddressLine1 != nil {
		org.AddressLine1 = strings.TrimSpace(*input.AddressLine1)
	}
	if input.AddressLine2 != nil {
		org.AddressLine2 = strings.TrimSpace(*input.AddressLine2)
	}
	if input.City != nil {
		org.City = strings.TrimSpace(*input.City)
	}
	if input.State != nil {
		org.State = strings.TrimSpace(*input.State)
	}
	if input.PostalCode != nil {
		org.PostalCode = strings.TrimSpace(*input.PostalCode)
	}
	if input.Country != nil {
		org.Country = strings.TrimSpace(*input.Country)
	}

	if err := ucutils.ValidateOrganization(org); err != nil {
		return err
	}

	oldPhotoKey := org.PhotoKey
	newPhotoKey := ""

	if input.Photo != nil {
		photo := input.Photo

		if photo.Size <= 0 || photo.Size > 5*1024*1024 {
			return errors.New("photo must be between 1 byte and 5MB")
		}

		switch photo.ContentType {
		case "image/jpeg", "image/png", "image/webp":
		default:
			return errors.New("photo must be JPEG, PNG, or WebP")
		}

		photoID, err := utils.NewID()
		if err != nil {
			return err
		}

		newPhotoKey = fmt.Sprintf(
			"organizations/%s/profile/%s",
			org.ID,
			photoID,
		)

		if err := u.storage.Upload(
			ctx,
			newPhotoKey,
			photo.Reader,
			photo.Size,
			photo.ContentType,
		); err != nil {
			return err
		}

		org.PhotoKey = newPhotoKey
	}

	if err := u.repository.Update(ctx, org); err != nil {
		if newPhotoKey != "" {
			_ = u.storage.Delete(ctx, newPhotoKey)
		}

		u.logger.ErrorContext(ctx, "failed to update organization",
			"id", id,
			"error", err,
		)
		return err
	}

	if newPhotoKey != "" && oldPhotoKey != "" {
		if err := u.storage.Delete(ctx, oldPhotoKey); err != nil {
			u.logger.ErrorContext(ctx, "failed to delete old organization photo",
				"photo_key", oldPhotoKey,
				"error", err,
			)
		}
	}

	return nil
}

func (u *UseCase) Delete(ctx context.Context, id string) error {
	ok := ucutils.ValidateID(id)
	if !ok {
		return errors.New("invalid ID")
	}

	u.logger.InfoContext(ctx, "deleting organization", "id", id)

	if err := u.repository.Delete(ctx, id); err != nil {
		u.logger.ErrorContext(ctx, "failed to delete organization", "id", id, "error", err)
		return err
	}

	u.logger.InfoContext(ctx, "organization deleted", "id", id)
	return nil
}

func (u *UseCase) Invite(
	ctx context.Context,
	organizationID string,
	emailAddress string,
	role entity.UserRole,
) error {
	emailAddress = strings.ToLower(strings.TrimSpace(emailAddress))

	if organizationID == "" {
		return errors.New("organization ID is required")
	}
	if emailAddress == "" {
		return errors.New("email is required")
	}
	if !utils.IsValidEmail(emailAddress) {
		return errors.New("invalid email address")
	}
	if !role.IsValid() {
		return errors.New("invalid invitation role")
	}

	existing, err := u.users.GetByEmail(ctx, emailAddress)
	if err == nil && existing != nil {
		return errors.New("user already exists and cannot be invited to another organization")
	}
	if err != nil && !errors.Is(err, repo.ErrNotFound) {
		return err
	}

	rawToken, err := utils.GenerateToken()
	if err != nil {
		return err
	}

	id, err := utils.NewID()
	if err != nil {
		return err
	}

	invitation := &entity.OrganizationInvitation{
		ID:             id,
		OrganizationID: organizationID,
		Email:          emailAddress,
		Role:           role,
		TokenHash:      utils.HashToken(rawToken),
		ExpiresAt:      time.Now().Add(7 * 24 * time.Hour),
	}

	if err := u.invitations.Upsert(ctx, invitation); err != nil {
		return err
	}

	invitationURL := fmt.Sprintf(
		"%s/accept-invitation?token=%s",
		strings.TrimRight(u.frontendCfg.URL, "/"),
		url.QueryEscape(rawToken),
	)
	body := fmt.Sprintf(`
    <h2>You're invited to StitchForce!</h2>
    <p>You have been invited to join an organization.</p>
    <p>Assigned role: %s</p>
    <a href="%s"
       style="display:inline-block;background:#18181b;color:#fff;
              padding:12px 24px;border-radius:6px;text-decoration:none;">
        Accept Invitation
    </a>
    <p>This invitation expires in 7 days.</p>
`,
		html.EscapeString(string(role)),
		html.EscapeString(invitationURL),
	)

	rendered, err := email.Render(body)
	if err != nil {
		return err
	}

	return u.emailer.Send(ctx, email.Message{
		To:      []string{emailAddress},
		Subject: "You've been invited to StitchForce",
		HTML:    rendered,
	})
}

func (u *UseCase) AcceptInvitation(
	ctx context.Context,
	rawToken string,
	name string,
	password string,
) (*entity.User, error) {
	name = strings.TrimSpace(name)

	if len(name) < 2 || len(name) > 100 {
		return nil, errors.New("name must be between 2 and 100 characters")
	}

	if len(password) < 8 || len(password) > 72 {
		return nil, errors.New("password must be between 8 and 72 bytes")
	}

	if strings.TrimSpace(rawToken) == "" {
		return nil, errors.New("invitation token is required")
	}

	invitation, err := u.invitations.GetByTokenHash(
		ctx,
		utils.HashToken(rawToken),
	)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, errors.New("invalid invitation")
		}
		return nil, err
	}

	if invitation.AcceptedAt != nil {
		return nil, errors.New("invitation already accepted")
	}

	if !invitation.ExpiresAt.After(time.Now()) {
		return nil, errors.New("invitation expired")
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return nil, err
	}

	userID, err := utils.NewID()
	if err != nil {
		return nil, err
	}

	user := &entity.User{
		ID:           userID,
		Name:         name,
		PasswordHash: string(passwordHash),
	}

	if err := u.invitations.Accept(ctx, invitation.ID, user); err != nil {
		return nil, err
	}

	user.OrganizationID = invitation.OrganizationID
	user.Email = invitation.Email
	user.Role = invitation.Role

	verifiedAt := time.Now()
	user.EmailVerifiedAt = &verifiedAt
	user.PasswordHash = ""

	return user, nil
}

func (u *UseCase) GetDashboardOverview(
	ctx context.Context,
	organizationID string,
) (*output.DashboardOverview, error) {
	organizationID, err := ucutils.ValidateOrganizationID(organizationID)
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
	organizationID, err := ucutils.ValidateOrganizationID(organizationID)
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

func (u *UseCase) ListInvitations(
	ctx context.Context,
	organizationID string,
) ([]*entity.OrganizationInvitation, error) {
	if organizationID == "" {
		return nil, errors.New("organization ID is required")
	}

	return u.invitations.ListByOrganization(ctx, organizationID)
}

func (u *UseCase) ResendInvitation(
	ctx context.Context,
	organizationID string,
	invitationID string,
) error {
	if organizationID == "" || invitationID == "" {
		return errors.New("organization ID and invitation ID are required")
	}

	invitation, err := u.invitations.GetByID(ctx, organizationID, invitationID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return errors.New("invitation not found")
		}
		return err
	}

	if invitation.AcceptedAt != nil {
		return errors.New("invitation already accepted")
	}

	rawToken, err := utils.GenerateToken()
	if err != nil {
		return err
	}

	invitation.TokenHash = utils.HashToken(rawToken)
	invitation.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)

	if err := u.invitations.Upsert(ctx, invitation); err != nil {
		return err
	}

	invitationURL := fmt.Sprintf(
		"%s/accept-invitation?token=%s",
		strings.TrimRight(u.frontendCfg.URL, "/"),
		url.QueryEscape(rawToken),
	)

	body := fmt.Sprintf(`
        <h2>You're invited to StitchForce!</h2>
        <p>You have been invited to join an organization.</p>
        <p>Assigned role: %s</p>

        <p style="margin: 28px 0;">
            <a href="%s"
               style="display:inline-block;
                      padding:12px 24px;
                      background:#18181b;
                      color:#ffffff;
                      text-decoration:none;
                      border-radius:6px;
                      font-weight:bold;">
                Accept Invitation
            </a>
        </p>

        <p>This invitation expires in 7 days.</p>
        <p style="color:#71717a;font-size:13px;">
            If you weren't expecting this invitation, you can ignore this email.
        </p>
    `,
		html.EscapeString(string(invitation.Role)),
		html.EscapeString(invitationURL),
	)

	rendered, err := email.Render(body)
	if err != nil {
		return fmt.Errorf("render invitation email: %w", err)
	}

	return u.emailer.Send(ctx, email.Message{
		To:      []string{invitation.Email},
		Subject: "Your StitchForce invitation",
		HTML:    rendered,
	})
}

func (u *UseCase) CancelInvitation(
	ctx context.Context,
	organizationID string,
	invitationID string,
) error {
	if organizationID == "" || invitationID == "" {
		return errors.New("organization ID and invitation ID are required")
	}

	invitation, err := u.invitations.GetByID(ctx, organizationID, invitationID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return errors.New("invitation not found")
		}
		return err
	}

	if invitation.AcceptedAt != nil {
		return errors.New("cannot cancel an accepted invitation")
	}

	return u.invitations.Delete(ctx, organizationID, invitationID)
}

var _ usecase.OrganizationUseCase = (*UseCase)(nil)

func (u *UseCase) toOutput(
	ctx context.Context,
	org *entity.Organization,
) (*output.Organization, error) {
	result := &output.Organization{
		ID:               org.ID,
		Name:             org.Name,
		AddressLine1:     org.AddressLine1,
		AddressLine2:     org.AddressLine2,
		City:             org.City,
		State:            org.State,
		PostalCode:       org.PostalCode,
		Country:          org.Country,
		LicenseStatus:    org.LicenseStatus,
		LicenseExpiresAt: org.LicenseExpiresAt,
		CreatedAt:        org.CreatedAt,
		UpdatedAt:        org.UpdatedAt,
	}

	if org.PhotoKey != "" {
		photoURL, err := u.storage.SignedURL(ctx, org.PhotoKey, 15*time.Minute)
		if err != nil {
			return nil, err
		}
		result.PhotoURL = &photoURL
	}

	return result, nil
}
