package v1

import (
	"net/http"
	"strconv"
	"strings"

	"stitchforce-server/internal/controller/http/v1/request"
	"stitchforce-server/internal/entity"
	"stitchforce-server/internal/usecase"
	"stitchforce-server/internal/usecase/input"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type StyleOperationHandler struct {
	operations usecase.StyleOperationUseCase
	validate   *validator.Validate
}

func NewStyleOperationHandler(
	operations usecase.StyleOperationUseCase,
	validate *validator.Validate,
) *StyleOperationHandler {
	return &StyleOperationHandler{
		operations: operations,
		validate:   validate,
	}
}

func (h *StyleOperationHandler) Create(c *gin.Context) {
	organizationID := c.Param("organizationID")
	styleID := c.Param("styleID")

	if organizationID == "" || styleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "organization id and style id are required",
		})
		return
	}

	var req request.CreateStyleOperation

	if err := c.ShouldBindJSON(&req); err != nil {
		writeValidationError(c, err)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeValidationError(c, err)
		return
	}

	operation := &entity.StyleOperation{
		OrganizationID: organizationID,
		StyleID:        styleID,
		Name:           req.Name,
		RatePerPiece:   req.RatePerPiece,
	}

	if err := h.operations.Create(
		c.Request.Context(),
		organizationID,
		styleID,
		operation,
	); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, operation)
}

func (h *StyleOperationHandler) Get(c *gin.Context) {
	organizationID := c.Param("organizationID")
	styleID := c.Param("styleID")
	operationID := c.Param("id")

	operation, err := h.operations.GetByID(
		c.Request.Context(),
		organizationID,
		styleID,
		operationID,
	)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, operation)
}

func (h *StyleOperationHandler) List(c *gin.Context) {
	organizationID := c.Param("organizationID")
	styleID := c.Param("styleID")

	if organizationID == "" || styleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "organization id and style id are required",
		})
		return
	}

    // Style Ops Query parameter keys for filtering style operations
	const (
		Search         = "search"
		Limit            = "limit"
		Offset           = "offset"
	)
	filter := input.StyleOperationFilter{
		Search: strings.TrimSpace(c.Query(Search)),
		Limit:  20,
		Offset: 0,
	}

	if limit := c.Query(Limit); limit != "" {
		value, err := strconv.Atoi(limit)
		if err != nil || value <= 0 || value > 100 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "limit must be between 1 and 100",
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

	operations, err := h.operations.List(
		c.Request.Context(),
		organizationID,
		styleID,
		filter,
	)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, operations)
}

func (h *StyleOperationHandler) Update(c *gin.Context) {
	organizationID := c.Param("organizationID")
	styleID := c.Param("styleID")

	var req request.UpdateStyleOperation

	if err := c.ShouldBindJSON(&req); err != nil {
		writeValidationError(c, err)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeValidationError(c, err)
		return
	}

	operation := &entity.StyleOperation{
		ID:             c.Param("id"),
		OrganizationID: organizationID,
		StyleID:        styleID,
		Name:           req.Name,
		RatePerPiece:   req.RatePerPiece,
	}

	if err := h.operations.Update(
		c.Request.Context(),
		organizationID,
		styleID,
		operation,
	); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, operation)
}

func (h *StyleOperationHandler) Delete(c *gin.Context) {
	organizationID := c.Param("organizationID")
	styleID := c.Param("styleID")
	operationID := c.Param("id")

	if err := h.operations.Delete(
		c.Request.Context(),
		organizationID,
		styleID,
		operationID,
	); err != nil {
		writeError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
