package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"stitchforce-server/internal/controller/http/v1/request"
	"stitchforce-server/internal/entity"
	"stitchforce-server/internal/usecase"
	"stitchforce-server/internal/usecase/input"
)

type BundleHandler struct {
	bundles  usecase.BundleUseCase
	validate *validator.Validate
}

func NewBundleHandler(
	bundles usecase.BundleUseCase,
	validate *validator.Validate,
) *BundleHandler {
	return &BundleHandler{
		bundles:  bundles,
		validate: validate,
	}
}

// CREATE

func (h *BundleHandler) Create(c *gin.Context) {
	organizationID := c.Param("organizationID")

	if organizationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "organization id is required",
		})
		return
	}

	var req request.CreateBundle

	if err := c.ShouldBindJSON(&req); err != nil {
		writeValidationError(c, err)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeValidationError(c, err)
		return
	}

	bundle := &entity.Bundle{
		BundleNumber: req.BundleNumber,
		StyleID:      req.StyleID,
		SizeID:       req.SizeID,
		Quantity:     req.Quantity,
		Attributes:   req.Attributes,
		Status:       req.Status,
		Remarks:      req.Remarks,
	}

	if bundle.Status == "" {
		bundle.Status = entity.BundleStatusCreated
	}

	if err := h.bundles.Create(
		c.Request.Context(),
		organizationID,
		bundle,
	); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "bundle created successfully",
	})
}

// GET BY ID

func (h *BundleHandler) Get(c *gin.Context) {
	organizationID := c.Param("organizationID")

	if organizationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "organization id is required",
		})
		return
	}

	bundle, err := h.bundles.GetByID(
		c.Request.Context(),
		organizationID,
		c.Param("id"),
	)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, bundle)
}

// LIST

func (h *BundleHandler) List(c *gin.Context) {
	organizationID := c.Param("organizationID")

	if organizationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "organization id is required",
		})
		return
	}

	bundles, err := h.bundles.List(
		c.Request.Context(),
		organizationID,
	)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, bundles)
}

// UPDATE

func (h *BundleHandler) Update(c *gin.Context) {
	organizationID := c.Param("organizationID")

	if organizationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "organization id is required",
		})
		return
	}

	var req request.UpdateBundle

	if err := c.ShouldBindJSON(&req); err != nil {
		writeValidationError(c, err)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeValidationError(c, err)
		return
	}

	bundle := &input.UpdateBundleInput{
		BundleNumber: req.BundleNumber,
		SizeID:       req.SizeID,
		Quantity:     req.Quantity,
		Attributes:   req.Attributes,
		Status:       req.Status,
		Remarks:      req.Remarks,
	}

	if err := h.bundles.Update(
		c.Request.Context(),
		organizationID,
		c.Param("id"),
		bundle,
	); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "bundle updated successfully",
	})
}

// DELETE

func (h *BundleHandler) Delete(c *gin.Context) {
	organizationID := c.Param("organizationID")

	if organizationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "organization id is required",
		})
		return
	}

	if err := h.bundles.Delete(
		c.Request.Context(),
		organizationID,
		c.Param("id"),
	); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "bundle deleted successfully",
	})
}
