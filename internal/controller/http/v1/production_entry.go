package v1

import (
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

const DateLayout = "2006-01-02"

type ProductionEntryHandler struct {
	productionEntries usecase.ProductionEntryUseCase
	validate          *validator.Validate
}

func NewProductionEntryHandler(
	productionEntries usecase.ProductionEntryUseCase,
	validate *validator.Validate,
) *ProductionEntryHandler {
	return &ProductionEntryHandler{
		productionEntries: productionEntries,
		validate:          validate,
	}
}

// CREATE

func (h *ProductionEntryHandler) Create(c *gin.Context) {
	organizationID := c.Param("organizationID")

	if organizationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "organization id is required",
		})
		return
	}

	var req request.CreateProductionEntry

	if err := c.ShouldBindJSON(&req); err != nil {
		writeValidationError(c, err)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeValidationError(c, err)
		return
	}

	entry := &entity.ProductionEntry{
		BundleID:         req.BundleID,
		WorkerID:         req.WorkerID,
		StyleOperationID: req.StyleOperationID,
		ProductionDate:   req.ProductionDate,
		Remarks:          req.Remarks,
	}

	if err := h.productionEntries.Create(
		c.Request.Context(),
		organizationID,
		entry,
	); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "production entry created successfully",
		"id":      entry.ID,
	})
}

// GET BY ID

func (h *ProductionEntryHandler) Get(c *gin.Context) {
	organizationID := c.Param("organizationID")

	if organizationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "organization id is required",
		})
		return
	}

	entry, err := h.productionEntries.GetByID(
		c.Request.Context(),
		organizationID,
		c.Param("id"),
	)

	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, entry)
}

// LIST

func (h *ProductionEntryHandler) List(c *gin.Context) {
	organizationID := c.Param("organizationID")

	if organizationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "organization id is required",
		})
		return
	}

	// Query parameter keys for filtering and pagination
	const (
		Search           = "search"
		BundleID         = "bundle_id"
		WorkerID         = "worker_id"
		StyleID          = "style_id"
		StyleOperationID = "style_operation_id"
		Limit            = "limit"
		Offset           = "offset"
		FromDate         = "from_date"
		ToDate           = "to_date"
	)

	filter := input.ProductionEntryFilter{
		Search:           strings.TrimSpace(c.Query(Search)),
		BundleID:         c.Query(BundleID),
		WorkerID:         c.Query(WorkerID),
		StyleID:          c.Query(StyleID),
		StyleOperationID: c.Query(StyleOperationID),
		Limit:            20,
		Offset:           0,
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

	// Parse from_date.
	if value := c.Query(FromDate); value != "" {
		date, err := time.Parse(DateLayout, value)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid from_date; expected YYYY-MM-DD",
			})
			return
		}

		filter.FromDate = &date
	}

	// Parse to_date (inclusive).
	if value := c.Query(ToDate); value != "" {
		date, err := time.Parse(DateLayout, value)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid to_date; expected YYYY-MM-DD",
			})
			return
		}

		date = date.AddDate(0, 0, 1)
		filter.ToDate = &date
	}

	entries, err := h.productionEntries.List(
		c.Request.Context(),
		organizationID,
		filter,
	)

	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, entries)
}

// UPDATE

func (h *ProductionEntryHandler) Update(c *gin.Context) {
	organizationID := c.Param("organizationID")

	if organizationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "organization id is required",
		})
		return
	}

	var req request.UpdateProductionEntry

	if err := c.ShouldBindJSON(&req); err != nil {
		writeValidationError(c, err)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeValidationError(c, err)
		return
	}

	entry := &entity.ProductionEntry{
		ID:             c.Param("id"),
		BundleID:       req.BundleID,
		WorkerID:       req.WorkerID,
		Quantity:       req.Quantity,
		ProductionDate: req.ProductionDate,
		Remarks:        req.Remarks,
	}

	if err := h.productionEntries.Update(
		c.Request.Context(),
		organizationID,
		entry,
	); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "production entry updated successfully",
	})
}

// DELETE

func (h *ProductionEntryHandler) Delete(c *gin.Context) {
	organizationID := c.Param("organizationID")

	if organizationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "organization id is required",
		})
		return
	}

	if err := h.productionEntries.Delete(
		c.Request.Context(),
		organizationID,
		c.Param("id"),
	); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "production entry deleted successfully",
	})
}
