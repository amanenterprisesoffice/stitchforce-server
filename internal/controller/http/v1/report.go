package v1

import (
	"net/http"
	"stitchforce-server/internal/usecase"
	"stitchforce-server/internal/usecase/input"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ReportHandler struct {
	reportUseCase usecase.ReportUseCase
	validate      *validator.Validate
}

func NewReportHandler(reportUseCase usecase.ReportUseCase, validate *validator.Validate) *ReportHandler {
	return &ReportHandler{
		reportUseCase: reportUseCase,
		validate:      validate,
	}
}

func (h *ReportHandler) List(c *gin.Context) {
	organizationID := c.Param("organizationID")

	if organizationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "organization id is required",
		})
		return
	}

	const (
		Year   = "year"
		Month  = "month"
		Limit  = "limit"
		Offset = "offset"
	)

	filter := input.ReportFilter{
		Limit:  20,
		Offset: 0,
	}

	// Parse year.
	if value := c.Query(Year); value != "" {
		year, err := strconv.Atoi(value)
		if err != nil || year <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid year",
			})
			return
		}
		filter.Year = &year
	}

	// Parse month.
	if value := c.Query(Month); value != "" {
		month, err := strconv.Atoi(value)
		if err != nil || month < 1 || month > 12 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid month",
			})
			return
		}
		filter.Month = &month
	}

	// Parse limit.
	if value := c.Query(Limit); value != "" {
		limit, err := strconv.Atoi(value)
		if err != nil || limit <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid limit",
			})
			return
		}
		filter.Limit = limit
	}

	// Parse offset.
	if value := c.Query(Offset); value != "" {
		offset, err := strconv.Atoi(value)
		if err != nil || offset < 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid offset",
			})
			return
		}
		filter.Offset = offset
	}

	reports, err := h.reportUseCase.List(
		c.Request.Context(),
		organizationID,
		filter,
	)

	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, reports)
}

func (h *ReportHandler) GetDownloadURL(c *gin.Context) {
	organizationID := c.Param("organizationID")
	reportID := c.Param("reportID")

	downloadURL, err := h.reportUseCase.GetDownloadURL(
		c.Request.Context(),
		organizationID,
		reportID,
	)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"download_url": downloadURL,
	})
}
