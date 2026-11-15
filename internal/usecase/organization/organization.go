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

	if input.DashboardAlertEnabled != nil {
		org.DashboardAlertEnabled = *input.DashboardAlertEnabled
	}

	if input.DashboardAlertEmail != nil {
		org.DashboardAlertEmail = strings.TrimSpace(*input.DashboardAlertEmail)
	}

	if input.DashboardAlertSchedule != nil {
		org.DashboardAlertSchedule = *input.DashboardAlertSchedule
	}

	if org.DashboardAlertEnabled {
		if org.DashboardAlertEmail == "" {
			return errors.New("dashboard alert email is required")
		}

		if err := org.DashboardAlertSchedule.Validate(); err != nil {
			return err
		}
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
	organizationID = strings.TrimSpace(organizationID)
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

	if !role.IsValid() || role == entity.UserRoleOrgAdmin {
		return errors.New("invalid invitation role")
	}

	organization, err := u.repository.GetByID(ctx, organizationID)
	if err != nil {
		return fmt.Errorf("get organization: %w", err)
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

	rendered, err := u.buildInvitationEmail(
		organization,
		role,
		invitationURL,
	)
	if err != nil {
		return fmt.Errorf("render invitation email: %w", err)
	}

	return u.emailer.Send(ctx, email.Message{
		To:      []string{emailAddress},
		Subject: "You've been invited to StitchForce",
		HTML:    rendered,
	})
}

func (u *UseCase) ResendInvitation(
	ctx context.Context,
	organizationID string,
	invitationID string,
) error {
	organizationID = strings.TrimSpace(organizationID)
	invitationID = strings.TrimSpace(invitationID)

	if organizationID == "" || invitationID == "" {
		return errors.New("organization ID and invitation ID are required")
	}

	organization, err := u.repository.GetByID(ctx, organizationID)
	if err != nil {
		return fmt.Errorf("get organization: %w", err)
	}

	invitation, err := u.invitations.GetByID(
		ctx,
		organizationID,
		invitationID,
	)
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

	rendered, err := u.buildInvitationEmail(
		organization,
		invitation.Role,
		invitationURL,
	)
	if err != nil {
		return fmt.Errorf("render invitation email: %w", err)
	}

	return u.emailer.Send(ctx, email.Message{
		To:      []string{invitation.Email},
		Subject: "Your StitchForce invitation",
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

	if !invitation.Role.IsValid() || invitation.Role == entity.UserRoleOrgAdmin {
		return nil, errors.New("invalid role in invitation")
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

func (u *UseCase) SendDueDashboardAlerts(
	ctx context.Context,
	now time.Time,
) (int, error) {
	organizations, err := u.repository.ListDashboardAlertOrganizations(ctx)
	if err != nil {
		return 0, fmt.Errorf("list dashboard alert organizations: %w", err)
	}

	sent := 0
	var failures []error

	for _, org := range organizations {
		scheduledAt, due, err := dashboardAlertOccurrence(
			now,
			org.DashboardAlertSchedule,
		)
		if err != nil {
			failures = append(failures, fmt.Errorf(
				"organization %s: %w",
				org.ID,
				err,
			))
			continue
		}

		if !due {
			continue
		}

		if org.DashboardAlertLastSentAt != nil &&
			!org.DashboardAlertLastSentAt.Before(scheduledAt) {
			continue
		}

		overview, err := u.repository.GetDashboardOverview(ctx, org.ID)
		if err != nil {
			failures = append(failures, fmt.Errorf(
				"organization %s: get dashboard overview: %w",
				org.ID,
				err,
			))
			continue
		}

		body, err := u.buildDashboardAlertEmail(org, overview, now)
		if err != nil {
			failures = append(failures, fmt.Errorf(
				"organization %s: build dashboard alert email: %w",
				org.ID,
				err,
			))
			continue
		}

		message := email.Message{
			To:      []string{org.DashboardAlertEmail},
			Subject: org.Name + " - Dashboard Summary",
			HTML:    body,
		}

		if err := u.emailer.Send(ctx, message); err != nil {
			failures = append(failures, fmt.Errorf(
				"organization %s: send dashboard alert: %w",
				org.ID,
				err,
			))
			continue
		}

		if err := u.repository.UpdateDashboardAlertLastSent(
			ctx,
			org.ID,
			now,
		); err != nil {
			failures = append(failures, fmt.Errorf(
				"organization %s: update dashboard alert timestamp: %w",
				org.ID,
				err,
			))
			continue
		}

		sent++

		u.logger.InfoContext(ctx, "dashboard alert sent",
			"organization_id", org.ID,
			"email", org.DashboardAlertEmail,
		)
	}

	return sent, errors.Join(failures...)
}

// Helpers
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


		DashboardAlertEnabled: org.DashboardAlertEnabled,
		DashboardAlertEmail: org.DashboardAlertEmail,
		DashboardAlertSchedule:   org.DashboardAlertSchedule,
		DashboardAlertLastSentAt: org.DashboardAlertLastSentAt,

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

func dashboardAlertOccurrence(
	now time.Time,
	schedule entity.DashboardAlertSchedule,
) (time.Time, bool, error) {
	if err := schedule.Validate(); err != nil {
		return time.Time{}, false, err
	}

	location, err := time.LoadLocation(schedule.Timezone)
	if err != nil {
		return time.Time{}, false, err
	}

	localNow := now.In(location)

	parsedTime, err := time.Parse("15:04", schedule.Time)
	if err != nil {
		return time.Time{}, false, err
	}

	scheduledAt := time.Date(
		localNow.Year(),
		localNow.Month(),
		localNow.Day(),
		parsedTime.Hour(),
		parsedTime.Minute(),
		0,
		0,
		location,
	)

	switch schedule.Frequency {
	case entity.Daily:
		// Every day.

	case entity.Weekly:
		if strings.ToLower(localNow.Weekday().String()) != string(schedule.DayOfWeek) {
			return time.Time{}, false, nil
		}

	case entity.Monthly:
		if localNow.Day() != schedule.DayOfMonth {
			return time.Time{}, false, nil
		}
	}

	if localNow.Before(scheduledAt) {
		return time.Time{}, false, nil
	}

	return scheduledAt, true, nil
}

func (u *UseCase) buildInvitationEmail(
	organization *entity.Organization,
	role entity.UserRole,
	invitationURL string,
) (string, error) {
	if organization == nil {
		return "", errors.New("organization is required")
	}

	addressParts := []string{
		organization.AddressLine1,
		organization.AddressLine2,
		organization.City,
		organization.State,
		organization.PostalCode,
		organization.Country,
	}

	locationParts := make([]string, 0, len(addressParts))

	for _, part := range addressParts {
		part = strings.TrimSpace(part)
		if part != "" {
			locationParts = append(locationParts, part)
		}
	}

	organizationName := html.EscapeString(organization.Name)
	organizationAddress := html.EscapeString(
		strings.Join(locationParts, ", "),
	)
	escapedRole := html.EscapeString(string(role))
	escapedURL := html.EscapeString(invitationURL)

	body := fmt.Sprintf(`
		<div>
			<div style="
				margin-bottom: 28px;
				padding-bottom: 20px;
				border-bottom: 1px solid #e4e4e7;
			">
				<h2 style="
					margin: 0;
					font-size: 24px;
					font-weight: 700;
				">
					You're invited to StitchForce!
				</h2>
				<p style="
					margin: 8px 0 0;
					color: #71717a;
					font-size: 14px;
				">
					Garment Manufacturing Management System
				</p>
			</div>

			<p style="font-size: 15px; line-height: 1.6;">
				You've been invited to join an organization on StitchForce.
			</p>

			<div style="
				margin: 24px 0;
				padding: 20px;
				background: #f4f4f5;
				border: 1px solid #e4e4e7;
				border-radius: 8px;
			">
				<p style="
					margin: 0 0 6px;
					color: #71717a;
					font-size: 12px;
					font-weight: 600;
					text-transform: uppercase;
				">
					Organization
				</p>

				<h3 style="
					margin: 0;
					font-size: 19px;
					font-weight: 700;
				">
					%s
				</h3>

				%s

				<div style="
					margin-top: 18px;
					padding-top: 14px;
					border-top: 1px solid #e4e4e7;
				">
					<p style="
						margin: 0 0 5px;
						color: #71717a;
						font-size: 12px;
						font-weight: 600;
						text-transform: uppercase;
					">
						Assigned role
					</p>
					<p style="
						margin: 0;
						font-size: 15px;
						font-weight: 600;
					">
						%s
					</p>
				</div>
			</div>

			<p style="font-size: 14px; line-height: 1.6;">
				Accept the invitation below to join the organization and
				start collaborating with your team.
			</p>

			<div style="margin: 28px 0; text-align: center;">
				<a href="%s"
					style="
						display: inline-block;
						padding: 13px 28px;
						background: #18181b;
						color: #ffffff;
						text-decoration: none;
						border-radius: 7px;
						font-size: 14px;
						font-weight: 600;
					">
					Accept Invitation
				</a>
			</div>

			<p style="
				color: #71717a;
				font-size: 13px;
				line-height: 1.6;
			">
				This invitation expires in 7 days. If you weren't
				expecting this invitation, you can safely ignore this email.
			</p>
		</div>
	`,
		organizationName,
		func() string {
			if organizationAddress == "" {
				return ""
			}

			return fmt.Sprintf(`
				<p style="
					margin: 7px 0 0;
					color: #71717a;
					font-size: 13px;
					line-height: 1.5;
				">%s</p>
			`, organizationAddress)
		}(),
		escapedRole,
		escapedURL,
	)

	return email.Render(body)
}

func (u *UseCase) buildDashboardAlertEmail(
	organization *entity.Organization,
	overview *output.DashboardOverview,
	generatedAt time.Time,
) (string, error) {
	if organization == nil {
		return "", errors.New("organization is required")
	}

	if overview == nil {
		return "", errors.New("dashboard overview is required")
	}

	organizationName := html.EscapeString(organization.Name)
	generatedTime := html.EscapeString(
		generatedAt.Format("02 Jan 2006, 03:04 PM"),
	)

	totalEarned := formatDashboardMoney(overview.TotalEarned)
	totalPaid := formatDashboardMoney(overview.TotalPaid)
	totalPending := formatDashboardMoney(overview.TotalPending)

	todayProductionAmount := formatDashboardMoney(
		overview.Today.ProductionAmount,
	)
	todayPaymentAmount := formatDashboardMoney(
		overview.Today.PaymentAmount,
	)

	body := fmt.Sprintf(`
		<div>
			<div style="
				margin-bottom: 28px;
				padding-bottom: 20px;
				border-bottom: 1px solid #e4e4e7;
			">
				<h2 style="margin: 0; font-size: 24px; font-weight: 700;">
					Dashboard Summary
				</h2>
				<p style="
					margin: 8px 0 0;
					color: #71717a;
					font-size: 14px;
				">
					%s
				</p>
				<p style="
					margin: 5px 0 0;
					color: #a1a1aa;
					font-size: 12px;
				">
					Generated %s
				</p>
			</div>

			<h3 style="font-size: 17px;">Factory Overview</h3>

			<table style="
				width: 100%%;
				border-collapse: collapse;
				font-size: 14px;
			">
				<tr>
					<td style="padding: 10px; border-bottom: 1px solid #eee;">Total Workers</td>
					<td style="padding: 10px; text-align: right;">%d</td>
				</tr>
				<tr>
					<td style="padding: 10px; border-bottom: 1px solid #eee;">Active Workers</td>
					<td style="padding: 10px; text-align: right;">%d</td>
				</tr>
				<tr>
					<td style="padding: 10px; border-bottom: 1px solid #eee;">Inactive Workers</td>
					<td style="padding: 10px; text-align: right;">%d</td>
				</tr>
				<tr>
					<td style="padding: 10px; border-bottom: 1px solid #eee;">Styles</td>
					<td style="padding: 10px; text-align: right;">%d</td>
				</tr>
				<tr>
					<td style="padding: 10px; border-bottom: 1px solid #eee;">Bundles</td>
					<td style="padding: 10px; text-align: right;">%d</td>
				</tr>
				<tr>
					<td style="padding: 10px;">Production Entries</td>
					<td style="padding: 10px; text-align: right;">%d</td>
				</tr>
			</table>

			<h3 style="margin-top: 28px; font-size: 17px;">Bundle Status</h3>

			<table style="
				width: 100%%;
				border-collapse: collapse;
				font-size: 14px;
			">
				<tr><td style="padding: 9px;">Created</td><td style="text-align: right;">%d</td></tr>
				<tr><td style="padding: 9px;">Ready</td><td style="text-align: right;">%d</td></tr>
				<tr><td style="padding: 9px;">In Progress</td><td style="text-align: right;">%d</td></tr>
				<tr><td style="padding: 9px;">Completed</td><td style="text-align: right;">%d</td></tr>
			</table>

			<h3 style="margin-top: 28px; font-size: 17px;">Worker Earnings</h3>

			<table style="
				width: 100%%;
				border-collapse: collapse;
				font-size: 14px;
			">
				<tr>
					<td style="padding: 10px; border-bottom: 1px solid #eee;">Total Earned</td>
					<td style="padding: 10px; text-align: right;">%s</td>
				</tr>
				<tr>
					<td style="padding: 10px; border-bottom: 1px solid #eee;">Total Paid</td>
					<td style="padding: 10px; text-align: right;">%s</td>
				</tr>
				<tr>
					<td style="padding: 10px;">Pending Payment</td>
					<td style="padding: 10px; text-align: right; font-weight: 700;">%s</td>
				</tr>
			</table>

			<h3 style="margin-top: 28px; font-size: 17px;">Today's Activity</h3>

			<table style="
				width: 100%%;
				border-collapse: collapse;
				font-size: 14px;
			">
				<tr><td style="padding: 9px;">Production Quantity</td><td style="text-align: right;">%d</td></tr>
				<tr><td style="padding: 9px;">Production Entries</td><td style="text-align: right;">%d</td></tr>
				<tr><td style="padding: 9px;">Production Amount</td><td style="text-align: right;">%s</td></tr>
				<tr><td style="padding: 9px;">Payments Made</td><td style="text-align: right;">%d</td></tr>
				<tr><td style="padding: 9px;">Payment Amount</td><td style="text-align: right;">%s</td></tr>
			</table>

			<div style="
				margin-top: 30px;
				padding-top: 18px;
				border-top: 1px solid #e4e4e7;
				color: #a1a1aa;
				font-size: 12px;
				line-height: 1.6;
			">
				This is an automated dashboard report from StitchForce.
				<br>
				Garment Manufacturing Management System
			</div>
		</div>
	`,
		organizationName,
		generatedTime,
		overview.TotalWorkers,
		overview.ActiveWorkers,
		overview.InactiveWorkers,
		overview.TotalStyles,
		overview.TotalBundles,
		overview.TotalProductionEntries,
		overview.BundleStatus.Created,
		overview.BundleStatus.Ready,
		overview.BundleStatus.InProgress,
		overview.BundleStatus.Completed,
		totalEarned,
		totalPaid,
		totalPending,
		overview.Today.ProductionQuantity,
		overview.Today.ProductionEntries,
		todayProductionAmount,
		overview.Today.PaymentCount,
		todayPaymentAmount,
	)

	return email.Render(body)
}

func formatDashboardMoney(amount int64) string {
	return fmt.Sprintf("₹%.2f", float64(amount)/100)
}

var _ usecase.OrganizationUseCase = (*UseCase)(nil)
