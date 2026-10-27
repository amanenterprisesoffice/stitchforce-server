package middleware

import (
	"net/http"
	"strings"
	"time"

	"stitchforce-server/pkg/jwt"
	"stitchforce-server/pkg/logger"

	"github.com/gin-gonic/gin"
)

const UserIDKey = "user_id"

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
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	})
}

func Auth(manager *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := manager.Parse(strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer "))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Set(UserIDKey, claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func UserID(c *gin.Context) string {
	value, _ := c.Get(UserIDKey)
	userID, _ := value.(string)
	return userID
}
