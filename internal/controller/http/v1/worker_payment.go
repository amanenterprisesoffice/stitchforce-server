package v1

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"stitchforce-server/internal/controller/http/v1/exports"
	"stitchforce-server/internal/controller/http/v1/request"
	"stitchforce-server/internal/entity"
	"stitchforce-server/internal/usecase"
	"stitchforce-server/internal/usecase/input"
	"stitchforce-server/internal/usecase/output"
	"stitchforce-server/pkg/exporter"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type WorkerPaymentHandler struct {
	payments usecase.WorkerPaymentUseCase
	validate *validator.Validate
}

func NewWorkerPaymentHandler(
	payments usecase.WorkerPaymentUseCase,
	validate *validator.Validate,
) *WorkerPaymentHandler {
	return &WorkerPaymentHandler{
		payments: payments,
		validate: validate,
	}
}

// CREATE

func (h *WorkerPaymentHandler) Create(c *gin.Context) {
	organizationID := c.Param("organizationID")

	if organizationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "organization id is required",
		})
		return
	}

	var req request.CreateWorkerPayment

	if err := c.ShouldBindJSON(&req); err != nil {
		writeValidationError(c, err)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeValidationError(c, err)
		return
	}

	fromDate, err := time.Parse(DateLayout, req.FromDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("incorrect from_date format; expected %s", DateLayout),
		})
		return
	}

	toDate, err := time.Parse(DateLayout, req.ToDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("incorrect to_date format; expected %s", DateLayout),
		})
		return
	}

	paymentDate, err := time.Parse(DateLayout, req.PaymentDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("incorrect payment_date format; expected %s", DateLayout),
		})
		return
	}

	createInput := &input.CreateWorkerPaymentInput{
		WorkerID:      strings.TrimSpace(req.WorkerID),
		FromDate:      fromDate,
		ToDate:        toDate,
		PaymentDate:   paymentDate,
		PaymentMethod: entity.PaymentMethod(req.PaymentMethod),
		Reference:     strings.TrimSpace(req.Reference),
		Remarks:       strings.TrimSpace(req.Remarks),
	}

	if err := h.payments.Create(
		c.Request.Context(),
		organizationID,
		createInput,
	); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "worker payment created successfully",
	})
}

// GET BY ID

func (h *WorkerPaymentHandler) Get(c *gin.Context) {
	organizationID := c.Param("organizationID")

	if organizationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "organization id is required",
		})
		return
	}

	payment, err := h.payments.GetByID(
		c.Request.Context(),
		organizationID,
		c.Param("id"),
	)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, payment)
}

// LIST

func (h *WorkerPaymentHandler) List(c *gin.Context) {
	organizationID := c.Param("organizationID")

	if organizationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "organization id is required",
		})
		return
	}

	const (
		WorkerID = "worker_id"
		Status   = "status"
		FromDate = "from_date"
		ToDate   = "to_date"
		Limit    = "limit"
		Offset   = "offset"
	)

	filter := input.WorkerPaymentFilter{
		WorkerID: strings.TrimSpace(c.Query(WorkerID)),
		Limit:    20,
		Offset:   0,
	}

	if value := c.Query(Status); value != "" {
		filter.Status = entity.PaymentStatus(value)
	}

	if value := c.Query(FromDate); value != "" {
		date, err := time.Parse(DateLayout, value)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("incorrect from_date format; expected %s", DateLayout),
			})
			return
		}
		filter.FromDate = &date
	}

	if value := c.Query(ToDate); value != "" {
		date, err := time.Parse(DateLayout, value)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("incorrect to_date format; expected %s", DateLayout),
			})
			return
		}
		filter.ToDate = &date
	}

	if filter.FromDate != nil &&
		filter.ToDate != nil &&
		filter.ToDate.Before(*filter.FromDate) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "to_date cannot be before from_date",
		})
		return
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

	payments, err := h.payments.List(
		c.Request.Context(),
		organizationID,
		filter,
	)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, payments)
}

// GET WORKER PAYMENT SUMMARY

func (h *WorkerPaymentHandler) GetWorkerSummary(c *gin.Context) {
	organizationID := c.Param("organizationID")
	workerID := c.Param("id")

	if organizationID == "" || workerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "organization id and worker id are required",
		})
		return
	}

	const (
		FromDate = "from_date"
		ToDate   = "to_date"
		Status   = "status"
	)
	filter := input.WorkerPaymentSummaryFilter{
		WorkerID: workerID,
	}

	if value := c.Query(FromDate); value != "" {
		date, err := time.Parse(DateLayout, value)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("incorrect from_date format; expected %s", DateLayout),
			})
			return
		}
		filter.FromDate = &date
	}

	if value := c.Query(ToDate); value != "" {
		date, err := time.Parse(DateLayout, value)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("incorrect to_date format; expected %s", DateLayout),
			})
			return
		}
		filter.ToDate = &date
	}

	if value := c.Query(Status); value != "" {
		filter.Status = entity.PaymentStatus(value)
	}

	summary, err := h.payments.GetSummary(
		c.Request.Context(),
		organizationID,
		filter,
	)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, summary)
}

func (h *WorkerPaymentHandler) GetSummary(c *gin.Context) {
	organizationID := c.Param("organizationID")

	if organizationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "organization id is required",
		})
		return
	}

	const (
		FromDate = "from_date"
		ToDate   = "to_date"
		Status   = "status"
	)
	filter := input.WorkerPaymentSummaryFilter{}

	if value := c.Query(FromDate); value != "" {
		date, err := time.Parse(DateLayout, value)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("incorrect from_date format; expected %s", DateLayout),
			})
			return
		}
		filter.FromDate = &date
	}

	if value := c.Query(ToDate); value != "" {
		date, err := time.Parse(DateLayout, value)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("incorrect to_date format; expected %s", DateLayout),
			})
			return
		}
		filter.ToDate = &date
	}

	if value := c.Query(Status); value != "" {
		filter.Status = entity.PaymentStatus(value)
	}

	summary, err := h.payments.GetSummary(
		c.Request.Context(),
		organizationID,
		filter,
	)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GET PAYABLE PRODUCTION ENTRIES

func (h *WorkerPaymentHandler) GetPayableEntries(c *gin.Context) {
	organizationID := c.Param("organizationID")
	workerID := c.Param("id")

	if organizationID == "" || workerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "organization id and worker id are required",
		})
		return
	}

	fromDateValue := c.Query("from_date")
	toDateValue := c.Query("to_date")

	if fromDateValue == "" || toDateValue == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "from_date and to_date are required",
		})
		return
	}

	fromDate, err := time.Parse(DateLayout, fromDateValue)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("from_date must be in %s format", DateLayout),
		})
		return
	}

	toDate, err := time.Parse(DateLayout, toDateValue)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("to_date must be in %s format", DateLayout),
		})
		return
	}

	if toDate.Before(fromDate) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "to_date cannot be before from_date",
		})
		return
	}

	filter := input.PayableEntriesFilter{
		FromDate: fromDate,
		ToDate:   toDate,
	}

	entries, err := h.payments.GetPayableEntries(
		c.Request.Context(),
		organizationID,
		workerID,
		filter,
	)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, entries)
}

// VOID PAYMENT

func (h *WorkerPaymentHandler) Void(c *gin.Context) {
	organizationID := c.Param("organizationID")
	paymentID := c.Param("id")

	if organizationID == "" || paymentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "organization id and payment id are required",
		})
		return
	}

	if err := h.payments.Void(
		c.Request.Context(),
		organizationID,
		paymentID,
	); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "worker payment voided successfully",
	})
}



func (h *WorkerPaymentHandler) Export(c *gin.Context) {
	organizationID := c.Param("organizationID")

	if organizationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "organization id is required",
		})
		return
	}

	format := strings.ToLower(
		strings.TrimSpace(c.DefaultQuery("format", "xlsx")),
	)

	if format != "pdf" && format != "xlsx" && format != "csv" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid format; supported formats are pdf, xlsx, csv",
		})
		return
	}

	filter := input.WorkerPaymentFilter{
		WorkerID: strings.TrimSpace(c.Query("worker_id")),
		Limit:    100,
		Offset:   0,
	}

	if value := strings.TrimSpace(c.Query("status")); value != "" {
		filter.Status = entity.PaymentStatus(value)
	}

	if value := c.Query("from_date"); value != "" {
		date, err := time.Parse(DateLayout, value)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf(
					"incorrect from_date format; expected %s",
					DateLayout,
				),
			})
			return
		}

		filter.FromDate = &date
	}

	if value := c.Query("to_date"); value != "" {
		date, err := time.Parse(DateLayout, value)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf(
					"incorrect to_date format; expected %s",
					DateLayout,
				),
			})
			return
		}

		filter.ToDate = &date
	}

	if filter.FromDate != nil &&
		filter.ToDate != nil &&
		filter.ToDate.Before(*filter.FromDate) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "to_date cannot be before from_date",
		})
		return
	}

	// Fetch all matching payments.
	const batchSize = 100

	payments := make([]*output.WorkerPayment, 0)
	offset := 0

	for {
		filter.Limit = batchSize
		filter.Offset = offset

		batch, err := h.payments.List(
			c.Request.Context(),
			organizationID,
			filter,
		)
		if err != nil {
			writeError(c, err)
			return
		}

		payments = append(payments, batch...)

		if len(batch) < batchSize {
			break
		}

		offset += batchSize
	}

	// Build export data.
	data := exports.BuildWorkerPayments(payments)

	data.Filename = fmt.Sprintf(
		"worker-payments-%s",
		time.Now().Format("20060102-150405"),
	)

	// Generate file.
	content, contentType, err := exporter.Render(format, data)
	if err != nil {
		writeError(c, err)
		return
	}

	c.Header("Content-Type", contentType)
	c.Header(
		"Content-Disposition",
		exporter.ContentDisposition(data.Filename, format),
	)

	c.Data(http.StatusOK, contentType, content)
}
