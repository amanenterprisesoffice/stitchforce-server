package v1

import (
	"net/http"
	"stitchforce-server/internal/usecase"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type CronHandler struct {
	organizations usecase.OrganizationUseCase
	validate      *validator.Validate
}

func NewCronHandler(
	organizations usecase.OrganizationUseCase,
	validate *validator.Validate,

) *CronHandler {
	return &CronHandler{
		organizations: organizations,
		validate:      validate,
	}
}

func (h *CronHandler) SendDashboardAlerts(c *gin.Context) {
	count, err := h.organizations.SendDueDashboardAlerts(c.Request.Context(), time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "dashboard alerts processed",
		"emails_sent": count,
	})
}
