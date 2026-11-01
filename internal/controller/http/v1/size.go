package v1

import (
	"net/http"
	"stitchforce-server/internal/controller/http/v1/request"
	"stitchforce-server/internal/entity"
	"stitchforce-server/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type SizeHandler struct {
	sizes    usecase.SizeUseCase
	validate *validator.Validate
}

func NewSizeHandler(
	sizes usecase.SizeUseCase,
	validate *validator.Validate,
) *SizeHandler {
	return &SizeHandler{
		sizes:    sizes,
		validate: validate,
	}
}

// CREATE

func (h *SizeHandler) Create(c *gin.Context) {
	organizationID := c.Param("organizationID")

	if organizationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "organization id is required",
		})
		return
	}

	var req request.CreateSize

	if err := c.ShouldBind(&req); err != nil {
		writeValidationError(c, err)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeValidationError(c, err)
		return
	}

	createInput := &entity.Size{
		Name: req.Name,
	}

	if err := h.sizes.Create(
		c.Request.Context(),
		organizationID,
		createInput,
	); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "size created successfully",
	})
}

// GET BY ID

func (h *SizeHandler) Get(c *gin.Context) {
	organizationID := c.Param("organizationID")

	size, err := h.sizes.GetByID(
		c.Request.Context(),
		organizationID,
		c.Param("id"),
	)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, size)
}

// LIST

func (h *SizeHandler) List(c *gin.Context) {
	organizationID := c.Param("organizationID")

	sizes, err := h.sizes.List(
		c.Request.Context(),
		organizationID,
	)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, sizes)
}

// UPDATE

func (h *SizeHandler) Update(c *gin.Context) {
	organizationID := c.Param("organizationID")

	var req request.UpdateSize

	if err := c.ShouldBind(&req); err != nil {
		writeValidationError(c, err)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeValidationError(c, err)
		return
	}

	updateInput := &entity.Size{
		ID: c.Param("id"),
		Name: req.Name,
	}

	if err := h.sizes.Update(
		c.Request.Context(),
		organizationID,
		updateInput,
	); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "size updated successfully",
	})
}

// DELETE

func (h *SizeHandler) Delete(c *gin.Context) {
	organizationID := c.Param("organizationID")

	if err := h.sizes.Delete(
		c.Request.Context(),
		organizationID,
		c.Param("id"),
	); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "size deleted successfully",
	})
}
