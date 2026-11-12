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
	"github.com/gin-gonic/gin/binding"
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

// CREATE

func (h *GarmentWorkerHandler) Create(c *gin.Context) {
	organizationID := c.Param("organizationID")

	if organizationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "organization id is required",
		})
		return
	}

	var req request.CreateGarmentWorker

	if err := c.ShouldBindWith(&req, binding.FormMultipart); err != nil {
		writeValidationError(c, err)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeValidationError(c, err)
		return
	}

	var photo *input.UploadFile

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
			"error": "invalid profile photo",
		})
		return
	}

	createInput := &input.CreateGarmentWorkerInput{
		Name:   req.Name,
		Phone:  req.Phone,
		Gender: req.Gender,
		Photo:  photo,
	}

	if err := h.workers.Create(
		c.Request.Context(),
		organizationID,
		createInput,
	); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "garment worker created successfully",
	})
}

// GET BY ID

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

// LIST

func (h *GarmentWorkerHandler) List(c *gin.Context) {
	organizationID := c.Param("organizationID")

	if organizationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "organization id is required",
		})
		return
	}

	// Query parameter keys for filtering and pagination
	const (
		Search = "search"
		Status = "status"
		Limit  = "limit"
		Offset = "offset"
	)

	filter := input.GarmentWorkerFilter{
		Search: c.Query(Search),
		Status: entity.WorkerStatus(strings.TrimSpace(c.Query(Status))),
		Limit:  20,
		Offset: 0,
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

	workers, err := h.workers.List(
		c.Request.Context(),
		organizationID,
		filter,
	)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, workers)
}

// UPDATE

func (h *GarmentWorkerHandler) Update(c *gin.Context) {
	organizationID := c.Param("organizationID")
	workerID := c.Param("id")

	if organizationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organization id is required"})
		return
	}

	if workerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "garment worker id is required"})
		return
	}

	var req request.UpdateGarmentWorker
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid profile photo"})
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

	updateInput := &input.UpdateGarmentWorkerInput{
		Name:   valueOrEmpty(req.Name),
		Phone:  valueOrEmpty(req.Phone),
		Gender: valueOrEmpty(req.Gender),
		Status: entity.WorkerStatus(valueOrEmpty(req.Status)),
		Photo:  photo,
	}

	if err := h.workers.Update(c.Request.Context(), organizationID, workerID, updateInput); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "garment worker updated successfully"})
}


// DELETE (SOFT DELETE)

func (h *GarmentWorkerHandler) Delete(c *gin.Context) {
	organizationID := c.Param("organizationID")

	if organizationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "organization id is required",
		})
		return
	}

	if err := h.workers.Delete(
		c.Request.Context(),
		organizationID,
		c.Param("id"),
	); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "garment worker deleted successfully",
	})
}


// Helpers
func valueOrEmpty[T any](v *T) T {
	if v == nil {
		var zero T
		return zero
	}
	return *v
}
