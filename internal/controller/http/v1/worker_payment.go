package v1

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"stitchforce-server/internal/controller/http/v1/request"
	"stitchforce-server/internal/entity"
	"stitchforce-server/internal/usecase"
	"stitchforce-server/internal/usecase/input"

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
			"error": "incorrect from_date format",
		})
		return
	}

	toDate, err := time.Parse(DateLayout, req.ToDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "incorrect to_date format",
		})
		return
	}

	paymentDate, err := time.Parse(DateLayout, req.PaymentDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "incorrect payment_date format",
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
		Status:   strings.TrimSpace(c.Query(Status)),
		Limit:    20,
		Offset:   0,
	}

	if filter.Status != "" &&
		filter.Status != string(entity.PaymentStatusCompleted) &&
		filter.Status != string(entity.PaymentStatusVoided) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "status must be completed or voided",
		})
		return
	}

	if value := c.Query(FromDate); value != "" {
		date, err := time.Parse(DateLayout, value)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "incorrect from_date format",
			})
			return
		}
		filter.FromDate = &date
	}

	if value := c.Query(ToDate); value != "" {
		date, err := time.Parse(DateLayout, value)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "incorrect to_date format",
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
		if err != nil || value <= 0 || value > MaxPaginationLimit {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf(
					"limit must be between 1 and %d",
					MaxPaginationLimit,
				),
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

func (h *WorkerPaymentHandler) GetSummary(c *gin.Context) {
	organizationID := c.Param("organizationID")
	workerID := c.Param("id")

	if organizationID == "" || workerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "organization id and worker id are required",
		})
		return
	}

	summary, err := h.payments.GetSummary(
		c.Request.Context(),
		organizationID,
		workerID,
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
			"error": "from_date must be in YYYY-MM-DD format",
		})
		return
	}

	toDate, err := time.Parse(DateLayout, toDateValue)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "to_date must be in YYYY-MM-DD format",
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
