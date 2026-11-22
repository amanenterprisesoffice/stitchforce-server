package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"stitchforce-server/internal/controller/http/middleware"
	"stitchforce-server/internal/usecase"
)

type NotificationHandler struct {
	notifications usecase.NotificationUseCase
	validate      *validator.Validate
}

func NewNotificationHandler(
	notifications usecase.NotificationUseCase,
	validate *validator.Validate,

) *NotificationHandler {
	return &NotificationHandler{
		notifications: notifications,
		validate:      validate,
	}
}

// LIST
func (h *NotificationHandler) List(c *gin.Context) {
	userID := middleware.UserID(c)

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "you must be logged in to view notifications",
		})
		return
	}

	const (
		QueryLimit  = "limit"
		QueryOffset = "offset"
	)

	limit, err := strconv.Atoi(c.DefaultQuery(QueryLimit, "20"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid limit",
		})
		return
	}

	offset, err := strconv.Atoi(c.DefaultQuery(QueryOffset, "0"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid offset",
		})
		return
	}

	notifications, err := h.notifications.List(
		c.Request.Context(),
		userID,
		limit,
		offset,
	)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"limit":         limit,
		"offset":        offset,
	})
}

// UNREAD COUNT
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID := middleware.UserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}
	// log user ID for debugging purposes

	count, err := h.notifications.GetUnreadCount(
		c.Request.Context(),
		userID,
	)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"unread_count": count,
	})
}

// MARK AS READ
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID := middleware.UserID(c)
	notificationID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	if notificationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "notification id is required",
		})
		return
	}

	if err := h.notifications.MarkAsRead(
		c.Request.Context(),
		userID,
		notificationID,
	); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "notification marked as read",
	})
}

// MARK ALL AS READ
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID := middleware.UserID(c)

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	if err := h.notifications.MarkAllAsRead(
		c.Request.Context(),
		userID,
	); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "all notifications marked as read",
	})
}

// DELETE
func (h *NotificationHandler) Delete(c *gin.Context) {
	userID := middleware.UserID(c)
	notificationID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	if notificationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "notification id is required",
		})
		return
	}

	if err := h.notifications.Delete(
		c.Request.Context(),
		userID,
		notificationID,
	); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "notification deleted",
	})
}
