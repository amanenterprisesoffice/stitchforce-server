package v1

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

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

	// Query parameter keys for filtering and pagination
	const (
		Search      = "search"
		StyleID     = "style_id"
		SizeID      = "size_id"
		Status      = "status"
		Limit       = "limit"
		Offset      = "offset"
		CreatedFrom = "created_from"
		CreatedTo   = "created_to"
	)

	filter := input.BundleFilter{
		Search:  strings.TrimSpace(c.Query(Search)),
		StyleID: c.Query(StyleID),
		SizeID:  c.Query(SizeID),
		Status:  c.Query(Status),
		Limit:   20,
		Offset:  0,
	}

	if limit := c.Query(Limit); limit != "" {
		value, err := strconv.Atoi(limit)
		if err != nil || value <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "limit must be a positive integer",
			})
			return
		}
		filter.Limit = value
	}

	if offset := c.Query(Offset); offset != "" {
		value, err := strconv.Atoi(offset)
		if err != nil || value < 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "offset must be a non-negative integer",
			})
			return
		}
		filter.Offset = value
	}

	if createdFrom := c.Query(CreatedFrom); createdFrom != "" {
		date, err := time.Parse(DateLayout, createdFrom)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("created_from must be in %s format", DateLayout),
			})
			return
		}
		filter.CreatedFrom = &date
	}

	if createdTo := c.Query(CreatedTo); createdTo != "" {
		date, err := time.Parse(DateLayout, createdTo)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("created_to must be in %s format", DateLayout),
			})
			return
		}

		// Make the end date inclusive.
		date = date.AddDate(0, 0, 1)
		filter.CreatedTo = &date
	}

	bundles, err := h.bundles.List(
		c.Request.Context(),
		organizationID,
		filter,
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
