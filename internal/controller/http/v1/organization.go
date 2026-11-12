package v1

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"stitchforce-server/internal/controller/http/v1/request"
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

// UPDATE
func (h *OrganizationHandler) Update(c *gin.Context) {
	organizationID := c.Param("organizationID")

	if organizationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "organization id is required",
		})
		return
	}

	var req request.UpdateOrganization
	var photo *input.UploadFile

	if strings.HasPrefix(c.GetHeader("Content-Type"), "multipart/form-data") {
		if err := c.ShouldBindWith(&req, binding.FormMultipart); err != nil {
			writeValidationError(c, err)
			return
		}

		file, header, err := c.Request.FormFile("photo")
		if err == nil {
			defer file.Close()

			photo = &input.UploadFile{
				Reader:      file,
				Size:        header.Size,
				ContentType: header.Header.Get("Content-Type"),
				Filename:    header.Filename,
			}
		} else if err != http.ErrMissingFile {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid organization photo",
			})
			return
		}
	} else {
		if err := c.ShouldBindJSON(&req); err != nil {
			writeValidationError(c, err)
			return
		}
	}

	if err := h.validate.Struct(req); err != nil {
		writeValidationError(c, err)
		return
	}

	updateInput := &input.UpdateOrganizationInput{
		Name:         req.Name,
		AddressLine1: req.AddressLine1,
		AddressLine2: req.AddressLine2,
		City:         req.City,
		State:        req.State,
		PostalCode:   req.PostalCode,
		Country:      req.Country,
		Photo:        photo,
	}

	if err := h.organizations.Update(
		c.Request.Context(),
		organizationID,
		updateInput,
	); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "organization updated successfully",
	})
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

func (h *OrganizationHandler) Invite(c *gin.Context) {
	var req request.InviteUser

	if err := c.ShouldBindJSON(&req); err != nil {
		writeValidationError(c, err)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeValidationError(c, err)
		return
	}

	organizationID := c.Param("organizationID")
	if organizationID == "" {
		writeError(c, errors.New("organization ID is required"))
		return
	}

	err := h.organizations.Invite(
		c.Request.Context(),
		organizationID,
		req.Email,
		req.Role,
	)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Invitation sent successfully",
	})
}

func (h *OrganizationHandler) AcceptInvitation(c *gin.Context) {
	var req request.AcceptInvitation

	if err := c.ShouldBindJSON(&req); err != nil {
		writeValidationError(c, err)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeValidationError(c, err)
		return
	}

	u, err := h.organizations.AcceptInvitation(
		c.Request.Context(),
		req.Token,
		req.Name,
		req.Password,
	)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Invitation accepted successfully",
		"user":    u,
	})
}

func (h *OrganizationHandler) ListInvitations(c *gin.Context) {
	organizationID := c.Param("organizationID")

	invitations, err := h.organizations.ListInvitations(
		c.Request.Context(),
		organizationID,
	)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"message":     "Invitations retrieved successfully",
		"invitations": invitations,
	})
}

func (h *OrganizationHandler) ResendInvitation(c *gin.Context) {
	organizationID := c.Param("organizationID")
	invitationID := c.Param("id")

	if err := h.organizations.ResendInvitation(
		c.Request.Context(),
		organizationID,
		invitationID,
	); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Invitation resent successfully",
	})
}

func (h *OrganizationHandler) CancelInvitation(c *gin.Context) {
	organizationID := c.Param("organizationID")
	invitationID := c.Param("id")

	if err := h.organizations.CancelInvitation(
		c.Request.Context(),
		organizationID,
		invitationID,
	); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Invitation cancelled successfully",
	})
}
