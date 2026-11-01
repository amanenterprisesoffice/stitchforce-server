package v1

import (
	"net/http"

	"stitchforce-server/internal/controller/http/v1/request"
	"stitchforce-server/internal/entity"
	"stitchforce-server/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type StyleHandler struct {
	styles   usecase.StyleUseCase
	validate *validator.Validate
}

func NewStyleHandler(
	styles usecase.StyleUseCase,
	validate *validator.Validate,
) *StyleHandler {
	return &StyleHandler{
		styles:   styles,
		validate: validate,
	}
}

func (h *StyleHandler) Create(c *gin.Context) {
	organizationID := c.Param("organizationID")
	if organizationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "organization id is required",
		})
		return
	}

	var req request.CreateStyle

	if err := c.ShouldBindJSON(&req); err != nil {
		writeValidationError(c, err)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeValidationError(c, err)
		return
	}

	style := &entity.Style{
		Name:        req.Name,
		StyleNumber: req.StyleNumber,
		Description: req.Description,
	}

	if err := h.styles.Create(
		c.Request.Context(),
		organizationID,
		style,
	); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, style)
}

func (h *StyleHandler) Get(c *gin.Context) {
	organizationID := c.Param("organizationID")

	style, err := h.styles.GetByID(
		c.Request.Context(),
		organizationID,
		c.Param("styleID"),
	)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, style)
}

func (h *StyleHandler) List(c *gin.Context) {
	organizationID := c.Param("organizationID")

	styles, err := h.styles.List(
		c.Request.Context(),
		organizationID,
	)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, styles)
}

func (h *StyleHandler) Update(c *gin.Context) {
	organizationID := c.Param("organizationID")

	var req request.UpdateStyle

	if err := c.ShouldBindJSON(&req); err != nil {
		writeValidationError(c, err)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeValidationError(c, err)
		return
	}

	style := &entity.Style{
		ID:             c.Param("styleID"),
		OrganizationID: organizationID,
		Name:           req.Name,
		StyleNumber:    req.StyleNumber,
		Description:    req.Description,
	}

	if err := h.styles.Update(
		c.Request.Context(),
		organizationID,
		style,
	); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, style)
}

func (h *StyleHandler) Delete(c *gin.Context) {
	organizationID := c.Param("organizationID")

	if err := h.styles.Delete(
		c.Request.Context(),
		organizationID,
		c.Param("styleID"),
	); err != nil {
		writeError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
