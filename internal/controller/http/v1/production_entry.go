package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"stitchforce-server/internal/controller/http/v1/request"
	"stitchforce-server/internal/entity"
	"stitchforce-server/internal/usecase"
)

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
		Quantity:         req.Quantity,
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

	entries, err := h.productionEntries.List(
		c.Request.Context(),
		organizationID,
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
