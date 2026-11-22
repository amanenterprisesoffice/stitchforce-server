package postgres

import (
	"context"
	"database/sql"
	"encoding/json"

	"stitchforce-server/internal/entity"
	"stitchforce-server/internal/repo"
)

type NotificationRepository struct {
	db *sql.DB
}

func New(db *sql.DB) repo.NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(
	ctx context.Context,
	notification *entity.Notification,
) error {
	data, err := json.Marshal(notification.Data)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO notifications (
			organization_id,
			user_id,
			type,
			title,
			message,
			data,
			action_url
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`

	return r.db.QueryRowContext(
		ctx,
		query,
		notification.OrganizationID,
		notification.UserID,
		notification.Type,
		notification.Title,
		notification.Message,
		data,
		notification.ActionURL,
	).Scan(
		&notification.ID,
		&notification.CreatedAt,
	)
}

// add CreateBulk
func (r *NotificationRepository) CreateBulk(
	ctx context.Context,
	notifications []*entity.Notification,
) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO notifications (
			organization_id,
			user_id,
			type,
			title,
			message,
			data,
			action_url
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, notification := range notifications {
		data, err := json.Marshal(notification.Data)
		if err != nil {
			tx.Rollback()
			return err
		}

		err = stmt.QueryRowContext(
			ctx,
			notification.OrganizationID,
			notification.UserID,
			notification.Type,
			notification.Title,
			notification.Message,
			data,
			notification.ActionURL,
		).Scan(
			&notification.ID,
			&notification.CreatedAt,
		)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (r *NotificationRepository) List(
	ctx context.Context,
	userID string,
	limit int,
	offset int,
) ([]*entity.Notification, error) {
	query := `
		SELECT
			id,
			organization_id,
			user_id,
			type,
			title,
			message,
			data,
			action_url,
			read_at,
			created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notifications := make([]*entity.Notification, 0)

	for rows.Next() {
		notification := &entity.Notification{}
		var data []byte
		var readAt sql.NullTime

		err := rows.Scan(
			&notification.ID,
			&notification.OrganizationID,
			&notification.UserID,
			&notification.Type,
			&notification.Title,
			&notification.Message,
			&data,
			&notification.ActionURL,
			&readAt,
			&notification.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if len(data) > 0 {
			if err := json.Unmarshal(data, &notification.Data); err != nil {
				return nil, err
			}
		}

		if notification.Data == nil {
			notification.Data = map[string]any{}
		}

		if readAt.Valid {
			notification.ReadAt = &readAt.Time
		}

		notifications = append(notifications, notification)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notifications, nil
}

func (r *NotificationRepository) GetUnreadCount(
	ctx context.Context,
	userID string,
) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM notifications
		WHERE user_id = $1
		  AND read_at IS NULL
	`

	var count int

	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *NotificationRepository) MarkAsRead(
	ctx context.Context,
	userID string,
	notificationID string,
) error {
	query := `
		UPDATE notifications
		SET read_at = NOW()
		WHERE id = $1
		  AND user_id = $2
		  AND read_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, notificationID, userID)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		// It may already be read, or it may not belong to this user.
		var exists bool

		err := r.db.QueryRowContext(
			ctx,
			`SELECT EXISTS (
				SELECT 1 FROM notifications
				WHERE id = $1 AND user_id = $2
			)`,
			notificationID,
			userID,
		).Scan(&exists)
		if err != nil {
			return err
		}

		if !exists {
			return sql.ErrNoRows
		}
	}

	return nil
}

func (r *NotificationRepository) MarkAllAsRead(
	ctx context.Context,
	userID string,
) error {
	query := `
		UPDATE notifications
		SET read_at = NOW()
		WHERE user_id = $1
		  AND read_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

func (r *NotificationRepository) Delete(
	ctx context.Context,
	userID string,
	notificationID string,
) error {
	query := `
		DELETE FROM notifications
		WHERE id = $1
		  AND user_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, notificationID, userID)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
