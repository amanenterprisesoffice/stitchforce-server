package middleware

import (
	"net/http"
	"stitchforce-server/config"
	"strings"

	"github.com/gin-gonic/gin"
)

func VerifyCronSecret(vercelCfg config.VercelConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		secret := vercelCfg.CronSecret
		if secret == "" {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "cron secret is not configured",
			})
			return
		}

		authHeader := c.GetHeader("Authorization")
		token := strings.TrimPrefix(authHeader, "Bearer ")

		if token == authHeader || token != secret {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			return
		}

		c.Next()
	}
}
