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
				"error": gin.H{
					"message": "You are not authorized to perform this action",
				},
			})
			return
		}

		if !authorization.HasPermission(roleValue, permission) {
			message, exists := authorization.PermissionMessages[permission]

			if !exists {
				message = "You don't have permission to perform this action"
			}

			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"message": message,
				},
			})
			return
		}

		c.Next()
	}
}