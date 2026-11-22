package notification

import (
	"context"
	"errors"
	"strings"

	"stitchforce-server/internal/entity"
	"stitchforce-server/internal/repo"
	"stitchforce-server/internal/usecase"
	ucutils "stitchforce-server/internal/usecase/utils"
	"stitchforce-server/pkg/logger"
)

type UseCase struct {
	repository repo.NotificationRepository
	logger     *logger.Logger
}

func New(
	repository repo.NotificationRepository,
	log *logger.Logger,
) usecase.NotificationUseCase {
	return &UseCase{
		repository: repository,
		logger:     log,
	}
}

func (u *UseCase) Create(
	ctx context.Context,
	notification *entity.Notification,
) error {
	if notification == nil {
		return errors.New("notification is required")
	}

	if !ucutils.ValidateID(notification.OrganizationID) {
		return errors.New("invalid organization ID")
	}

	if !ucutils.ValidateID(notification.UserID) {
		return errors.New("invalid user ID")
	}

	if strings.TrimSpace(string(notification.Type)) == "" {
		return errors.New("notification type is required")
	}

	if strings.TrimSpace(notification.Title) == "" {
		return errors.New("notification title is required")
	}

	if strings.TrimSpace(notification.Message) == "" {
		return errors.New("notification message is required")
	}

	notification.Title = strings.TrimSpace(notification.Title)
	notification.Message = strings.TrimSpace(notification.Message)

	if notification.Data == nil {
		notification.Data = map[string]any{}
	}

	if err := u.repository.Create(ctx, notification); err != nil {
		u.logger.ErrorContext(ctx, "failed to create notification",
			"user_id", notification.UserID,
			"organization_id", notification.OrganizationID,
			"error", err,
		)
		return err
	}

	return nil
}

// add create bulk
func (u *UseCase) CreateBulk(
	ctx context.Context,
	notifications []*entity.Notification,
) error {
	if len(notifications) == 0 {
		return errors.New("notifications are required")
	}

	for _, notification := range notifications {
		if notification == nil {
			return errors.New("notification is required")
		}

		if !ucutils.ValidateID(notification.OrganizationID) {
			return errors.New("invalid organization ID")
		}

		if !ucutils.ValidateID(notification.UserID) {
			return errors.New("invalid user ID")
		}

		if strings.TrimSpace(string(notification.Type)) == "" {
			return errors.New("notification type is required")
		}

		if strings.TrimSpace(notification.Title) == "" {
			return errors.New("notification title is required")
		}

		if strings.TrimSpace(notification.Message) == "" {
			return errors.New("notification message is required")
		}

		notification.Title = strings.TrimSpace(notification.Title)
		notification.Message = strings.TrimSpace(notification.Message)

		if notification.Data == nil {
			notification.Data = map[string]any{}
		}
	}

	if err := u.repository.CreateBulk(ctx, notifications); err != nil {
		u.logger.ErrorContext(ctx, "failed to create bulk notifications",
			"error", err,
		)
		return err
	}

	return nil
}

func (u *UseCase) List(
	ctx context.Context,
	userID string,
	limit int,
	offset int,
) ([]*entity.Notification, error) {
	if !ucutils.ValidateID(userID) {
		return nil, errors.New("invalid user ID")
	}

	if limit <= 0 {
		limit = 20
	}

	if limit > 100 {
		limit = 100
	}

	if offset < 0 {
		offset = 0
	}

	notifications, err := u.repository.List(ctx, userID, limit, offset)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to list notifications",
			"user_id", userID,
			"error", err,
		)
		return nil, err
	}

	return notifications, nil
}

func (u *UseCase) GetUnreadCount(
	ctx context.Context,
	userID string,
) (int, error) {
	if !ucutils.ValidateID(userID) {
		return 0, errors.New("invalid user ID")
	}

	u.logger.InfoContext(ctx, "Getting unread count for User ID:", "user_id", userID) // Log the user ID for debugging purposes
	count, err := u.repository.GetUnreadCount(ctx, userID)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get unread notification count",
			"user_id", userID,
			"error", err,
		)
		return 0, err
	}

	return count, nil
}

func (u *UseCase) MarkAsRead(
	ctx context.Context,
	userID string,
	notificationID string,
) error {
	if !ucutils.ValidateID(userID) {
		return errors.New("invalid user ID")
	}

	if !ucutils.ValidateID(notificationID) {
		return errors.New("invalid notification ID")
	}

	if err := u.repository.MarkAsRead(ctx, userID, notificationID); err != nil {
		u.logger.ErrorContext(ctx, "failed to mark notification as read",
			"user_id", userID,
			"notification_id", notificationID,
			"error", err,
		)
		return err
	}

	return nil
}

func (u *UseCase) MarkAllAsRead(
	ctx context.Context,
	userID string,
) error {
	if !ucutils.ValidateID(userID) {
		return errors.New("invalid user ID")
	}

	if err := u.repository.MarkAllAsRead(ctx, userID); err != nil {
		u.logger.ErrorContext(ctx, "failed to mark all notifications as read",
			"user_id", userID,
			"error", err,
		)
		return err
	}

	return nil
}

func (u *UseCase) Delete(
	ctx context.Context,
	userID string,
	notificationID string,
) error {
	if !ucutils.ValidateID(userID) {
		return errors.New("invalid user ID")
	}

	if !ucutils.ValidateID(notificationID) {
		return errors.New("invalid notification ID")
	}

	if err := u.repository.Delete(ctx, userID, notificationID); err != nil {
		u.logger.ErrorContext(ctx, "failed to delete notification",
			"user_id", userID,
			"notification_id", notificationID,
			"error", err,
		)
		return err
	}

	return nil
}