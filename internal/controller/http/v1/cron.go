package v1

import (
	"net/http"
	"time"

	"stitchforce-server/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type CronHandler struct {
	organizations usecase.OrganizationUseCase
	reports       usecase.ReportUseCase
	validate      *validator.Validate
}

func NewCronHandler(
	organizations usecase.OrganizationUseCase,
	reports usecase.ReportUseCase,
	validate *validator.Validate,
) *CronHandler {
	return &CronHandler{
		organizations: organizations,
		reports:       reports,
		validate:      validate,
	}
}

func (h *CronHandler) SendDashboardAlerts(c *gin.Context) {
	count, err := h.organizations.SendDueDashboardAlerts(
		c.Request.Context(),
		time.Now(),
	)
	if err != nil {
		writeError(c, err)
			return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "dashboard alerts processed",
		"emails_sent": count,
	})
}

func (h *CronHandler) GenerateMonthlyReports(c *gin.Context) {
	count, err := h.reports.GenerateMonthlyReports(
		c.Request.Context(),
		time.Now(),
	)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":           "monthly reports processed",
		"reports_generated": count,
	})
}

func (h *CronHandler) DeleteExpiredReports(c *gin.Context) {
	count, err := h.reports.DeleteExpiredReports(
		c.Request.Context(),
		time.Now(),
	)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "expired reports cleanup processed",
		"reports_deleted": count,
	})
}

func (h *CronHandler) SendReportExpirationReminders(c *gin.Context) {
	count, err := h.reports.SendExpirationReminders(
		c.Request.Context(),
		time.Now(),
	)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":          "report expiration reminders processed",
		"reminders_sent": count,
	})
}