package v1

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"stitchforce-server/internal/entity"
	"stitchforce-server/internal/usecase"
	"stitchforce-server/internal/usecase/input"
)

type OrganizationHandler struct {
	organizations usecase.OrganizationUseCase
	validate      *validator.Validate
}

func NewOrganizationHandler(
	organizations usecase.OrganizationUseCase,
	validate *validator.Validate,

) *OrganizationHandler {
	return &OrganizationHandler{
		organizations: organizations,
		validate:      validate,
	}
}

// CREATE
func (h *OrganizationHandler) Create(c *gin.Context) {
	var organization entity.Organization

	if err := c.ShouldBindJSON(&organization); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid organization payload",
		})
		return
	}

	organization.Name = strings.TrimSpace(organization.Name)

	if organization.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "organization name is required",
		})
		return
	}

	if err := h.organizations.Create(
		c.Request.Context(),
		&organization,
	); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, organization)
}

// GET BY ID
func (h *OrganizationHandler) GetByID(c *gin.Context) {
	organizationID := c.Param("organizationID")

	if organizationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "organization id is required",
		})
		return
	}

	organization, err := h.organizations.GetByID(
		c.Request.Context(),
		organizationID,
	)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, organization)
}

// LIST
func (h *OrganizationHandler) List(c *gin.Context) {
	organizations, err := h.organizations.List(
		c.Request.Context(),
	)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, organizations)
}

// UPDATE
func (h *OrganizationHandler) Update(c *gin.Context) {
	organizationID := c.Param("organizationID")

	if organizationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "organization id is required",
		})
		return
	}

	var organization entity.Organization

	if err := c.ShouldBindJSON(&organization); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid organization payload",
		})
		return
	}

	organization.ID = organizationID
	organization.Name = strings.TrimSpace(organization.Name)

	if organization.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "organization name is required",
		})
		return
	}

	if err := h.organizations.Update(
		c.Request.Context(),
		&organization,
	); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, organization)
}

// DELETE
func (h *OrganizationHandler) Delete(c *gin.Context) {
	organizationID := c.Param("organizationID")

	if organizationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "organization id is required",
		})
		return
	}

	if err := h.organizations.Delete(
		c.Request.Context(),
		organizationID,
	); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "organization deleted successfully",
	})
}

// DASHBOARD
func (h *OrganizationHandler) GetDashboardOverview(c *gin.Context) {
	organizationID := c.Param("organizationID")

	if organizationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "organization id is required",
		})
		return
	}

	dashboard, err := h.organizations.GetDashboardOverview(
		c.Request.Context(),
		organizationID,
	)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, dashboard)
}
func (h *OrganizationHandler) GetDashboardTrend(c *gin.Context) {
	organizationID := c.Param("organizationID")

	if organizationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "organization id is required",
		})
		return
	}

	const (
		ToDate   = "to_date"
		FromDate = "from_date"
		GroupBy  = "group_by"
	)

	fromDateValue := c.Query(FromDate)
	toDateValue := c.Query(ToDate)
	groupBy := input.DashboardGroup(c.DefaultQuery(GroupBy, input.DashboardGroupDaily))

	if fromDateValue == "" || toDateValue == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "from_date and to_date are required",
		})
		return
	}

	fromDate, err := time.Parse(DateLayout, fromDateValue)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("from_date must be in %s format", DateLayout),
		})
		return
	}

	toDate, err := time.Parse(DateLayout, toDateValue)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("to_date must be in %s format", DateLayout),
		})
		return
	}

	if toDate.Before(fromDate) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "to_date cannot be before from_date",
		})
		return
	}

	filter := input.DashboardFilter{
		FromDate: fromDate,
		ToDate:   toDate,
		GroupBy:  groupBy,
	}

	dashboard, err := h.organizations.GetDashboardTrend(
		c.Request.Context(),
		organizationID,
		filter,
	)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, dashboard)
}
