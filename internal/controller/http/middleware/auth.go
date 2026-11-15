package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"stitchforce-server/config"
	"stitchforce-server/internal/entity"
	"stitchforce-server/internal/repo"
	"stitchforce-server/pkg/jwt"
	"stitchforce-server/pkg/logger"

	"github.com/gin-gonic/gin"
)

const UserIDKey = "user_id"
const TokenKey = "access_token"
const UserRoleKey = "role"
const OrgIDKey = "organization_id"
const SessionIDKey = "session_id"

func RequestLogger(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		c.Next()
		log.Info("http request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration", time.Since(started).String(),
		)
	}
}

func Recovery(log *logger.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		log.Error("http panic recovered", "panic", recovered)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
	})
}

type SessionValidator interface {
	GetActiveSession(ctx context.Context, sessionID string) (*entity.UserSession, error)
}

func Auth(
	manager *jwt.Manager,
	users SessionValidator,
	cfg config.AppConfig,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string

		token, err := c.Cookie(TokenKey)

		if err != nil || token == "" {
			authHeader := c.GetHeader("Authorization")
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}

		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Authentication required",
			})
			return
		}

		claims, err := manager.Parse(token)
		if err != nil {
			RemoveCookie(c, cfg.Env)

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid or expired token",
			})
			return
		}

		if claims.SessionID == "" {
			RemoveCookie(c, cfg.Env)

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid session",
			})
			return
		}

		_, err = users.GetActiveSession(c.Request.Context(), claims.SessionID)
		if err != nil {
			if errors.Is(err, repo.ErrNotFound) {

				RemoveCookie(c, cfg.Env)

				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"message": "Session expired or revoked",
				})
				return
			}

			RemoveCookie(c, cfg.Env)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to validate session",
			})
			return
		}

		c.Set(UserIDKey, claims.UserID)
		c.Set(UserRoleKey, claims.Role)
		c.Set(OrgIDKey, claims.OrganizationID)
		c.Set(SessionIDKey, claims.SessionID)

		c.Next()
	}
}

func UserID(c *gin.Context) string {
	value, _ := c.Get(UserIDKey)
	userID, _ := value.(string)
	return userID
}

func RemoveCookie(c *gin.Context, appEnv string) {
	secure := appEnv == "production"

	c.SetCookie(
		TokenKey,
		"",
		-1,
		"/",
		"",
		secure,
		true,
	)
}
