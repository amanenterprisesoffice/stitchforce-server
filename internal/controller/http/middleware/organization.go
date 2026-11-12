package middleware

import (
	"context"
	"net/http"
	"time"

	"stitchforce-server/internal/usecase/output"

	"github.com/gin-gonic/gin"
)

type OrganizationChecker interface {
	GetByUserAndOrganization(
		ctx context.Context,
		userID string,
		organizationID string,
	) (*output.Organization, error)
}

func OrganizationAccess(
	organizationUseCase OrganizationChecker,
	requireActiveLicense bool,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := UserID(c)
		organizationID := c.Param("organizationID")

		if userID == "" || organizationID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			return
		}

		// Single database query: membership + organization details.
		org, err := organizationUseCase.GetByUserAndOrganization(
			c.Request.Context(),
			userID,
			organizationID,
		)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "you do not have access to this organization",
			})
			return
		}

		if requireActiveLicense {
			if org.LicenseStatus != "active" ||
				org.LicenseExpiresAt == nil ||
				!org.LicenseExpiresAt.After(time.Now()) {

				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error":              "organization license is expired or inactive",
					"license_status":     org.LicenseStatus,
					"license_expires_at": org.LicenseExpiresAt,
					// give some indication to client so it can redner some ui
					"invalid_license": true,
				})
				return
			}
		}

		c.Set("organization", org)
		c.Next()
	}
}
