package middleware

import (
	"net/http"
	"stitchforce-server/internal/authorization"

	"github.com/gin-gonic/gin"
)

func RequirePermission(permission authorization.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleValue, exists := c.Get(UserRoleKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "You are not authorized to perform this action",
			})
			return
		}

		if !authorization.HasPermission(roleValue, permission) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "you don't have permission to perform this action",
			})
			return
		}

		c.Next()
	}
}
