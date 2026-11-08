package v1

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"stitchforce-server/internal/entity"
	"stitchforce-server/internal/usecase"
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
func (h *OrganizationHandler) GetDashboard(c *gin.Context) {
	organizationID := c.Param("organizationID")

	if organizationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "organization id is required",
		})
		return
	}

	dashboard, err := h.organizations.GetDashboard(
		c.Request.Context(),
		organizationID,
	)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, dashboard)
}
