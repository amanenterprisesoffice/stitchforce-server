package v1

import (
	"net/http"

	"stitchforce-server/internal/controller/http/v1/request"
	"stitchforce-server/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type LicenseHandler struct {
	useCase  usecase.LicenseUseCase
	validate *validator.Validate
}

func NewLicenseHandler(
	licenseUseCase usecase.LicenseUseCase,
	validate *validator.Validate,
) *LicenseHandler {
	return &LicenseHandler{
		useCase:  licenseUseCase,
		validate: validate,
	}
}


func (h *LicenseHandler) Activate(c *gin.Context) {
	organizationID := c.Param("organizationID")

	var req request.ActivateLicenseRequest

	// use validate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "License key is required",
		})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeValidationError(c, err)
		return
	}

	if err := h.validate.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	license, err := h.useCase.Activate(
		c.Request.Context(),
		organizationID,
		req.LicenseKey,
	)

	if err != nil {
		status := http.StatusInternalServerError
		c.JSON(status, gin.H{
			"success": false,
			"error":   err,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "License activated successfully",
		"data":    license,
	})
}
