
package v1

import (
	"net/http"

	"stitchforce-server/internal/controller/http/v1/request"
	"stitchforce-server/internal/entity"
	"stitchforce-server/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type GarmentWorkerHandler struct {
	workers  usecase.GarmentWorkerUseCase
	validate *validator.Validate
}

func NewGarmentWorkerHandler(
	workers usecase.GarmentWorkerUseCase,
	validate *validator.Validate,
) *GarmentWorkerHandler {
	return &GarmentWorkerHandler{
		workers:  workers,
		validate: validate,
	}
}

func (h *GarmentWorkerHandler) Create(c *gin.Context) {
	organizationID := c.Param("organizationID")
	if organizationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "organization id is required",
		})
		return
	}

	var req request.CreateGarmentWorker

	if err := c.ShouldBindJSON(&req); err != nil {
		writeValidationError(c, err)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeValidationError(c, err)
		return
	}

	worker := &entity.GarmentWorker{
		OrganizationID: organizationID,
		Name:           req.Name,
		Phone:          req.Phone,
		Gender:         req.Gender,
	}

	if err := h.workers.Create(
		c.Request.Context(),
		organizationID,
		worker,
	); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, worker)
}

func (h *GarmentWorkerHandler) Get(c *gin.Context) {
	organizationID := c.Param("organizationID")

	worker, err := h.workers.GetByID(
		c.Request.Context(),
		organizationID,
		c.Param("id"),
	)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, worker)
}

func (h *GarmentWorkerHandler) List(c *gin.Context) {
	organizationID := c.Param("organizationID")

	workers, err := h.workers.List(
		c.Request.Context(),
		organizationID,
	)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, workers)
}

func (h *GarmentWorkerHandler) Update(c *gin.Context) {
	organizationID := c.Param("organizationID")

	var req request.UpdateGarmentWorker

	if err := c.ShouldBindJSON(&req); err != nil {
		writeValidationError(c, err)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeValidationError(c, err)
		return
	}

	worker := &entity.GarmentWorker{
		ID:             c.Param("id"),
		OrganizationID: organizationID,
		Name:           req.Name,
		Phone:          req.Phone,
		Gender:         req.Gender,
	}

	if err := h.workers.Update(
		c.Request.Context(),
		organizationID,
		worker,
	); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, worker)
}

func (h *GarmentWorkerHandler) Delete(c *gin.Context) {
	organizationID := c.Param("organizationID")

	if err := h.workers.Delete(
		c.Request.Context(),
		organizationID,
		c.Param("id"),
	); err != nil {
		writeError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
