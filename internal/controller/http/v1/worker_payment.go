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

	fromDate, err := parseDateValue(req.FromDate, "from_date")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	toDate, err := parseDateValue(req.ToDate, "to_date")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	paymentDate, err := parseDateValue(req.PaymentDate, "payment_date")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	createInput := &input.CreateWorkerPaymentInput{
		WorkerID:      strings.TrimSpace(req.WorkerID),
		FromDate:      *fromDate,
		ToDate:        *toDate,
		PaymentDate:   *paymentDate,
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

	fromDate, err := parseDate(c, FromDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	filter.FromDate = fromDate

	toDate, err := parseDate(c, ToDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	filter.ToDate = toDate

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

	fromDate, err := parseDate(c, FromDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	filter.FromDate = fromDate

	toDate, err := parseDate(c, ToDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	filter.ToDate = toDate

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
	fromDate, err := parseDate(c, FromDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	filter.FromDate = fromDate

	toDate, err := parseDate(c, ToDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	filter.ToDate = toDate
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

// LIST WORKER PAYMENT BALANCES

func (h *WorkerPaymentHandler) ListWorkerBalances(c *gin.Context) {
	organizationID := c.Param("organizationID")

	if organizationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "organization id is required",
		})
		return
	}

	const (
		Search           = "search"
		BundleID         = "bundle_id"
		WorkerID         = "worker_id"
		StyleID          = "style_id"
		StyleOperationID = "style_operation_id"
		FromDate         = "from_date"
		ToDate           = "to_date"
		Limit            = "limit"
		Offset           = "offset"
	)

	limit := 20
	if value := c.Query(Limit); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "limit must be a positive integer",
			})
			return
		}
		limit = parsed
	}

	offset := 0
	if value := c.Query(Offset); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "offset must be a non-negative integer",
			})
			return
		}
		offset = parsed
	}

	fromDate, err := parseDate(c, FromDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	toDate, err := parseDate(c, ToDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if fromDate != nil && toDate != nil && fromDate.After(*toDate) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "from_date cannot be after to_date",
		})
		return
	}

	filter := input.WorkerPaymentBalanceFilter{
		Search:           c.Query(Search),
		BundleID:         c.Query(BundleID),
		WorkerID:         c.Query(WorkerID),
		StyleID:          c.Query(StyleID),
		StyleOperationID: c.Query(StyleOperationID),
		FromDate:         fromDate,
		ToDate:           toDate,
		Limit:            limit,
		Offset:           offset,
	}

	balances, err := h.payments.ListWorkerBalances(
		c.Request.Context(),
		organizationID,
		filter,
	)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, balances)
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

	const (
		FromDate = "from_date"
		ToDate   = "to_date"
		Format   = "format"
		WorkerID = "worker_id"
	)

	format := strings.ToLower(
		strings.TrimSpace(c.DefaultQuery(Format, "xlsx")),
	)

	if format != "pdf" && format != "xlsx" && format != "csv" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid format; supported formats are pdf, xlsx, csv",
		})
		return
	}

	filter := input.WorkerPaymentFilter{
		WorkerID: strings.TrimSpace(c.Query(WorkerID)),
		Limit:    100,
		Offset:   0,
	}

	if value := strings.TrimSpace(c.Query("status")); value != "" {
		filter.Status = entity.PaymentStatus(value)
	}

	fromDate, err := parseDate(c, FromDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	filter.FromDate = fromDate

	toDate, err := parseDate(c, ToDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	filter.ToDate = toDate

	if filter.FromDate != nil &&
		filter.ToDate != nil &&
		filter.ToDate.Before(*filter.FromDate) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "to_date cannot be before from_date",
		})
		return
	}

	const batchSize = 100

	payments := make([]*output.SlimWorkerPayment, 0)
	offset := 0

	for {
		filter.Limit = batchSize
		filter.Offset = offset

		result, err := h.payments.List(
			c.Request.Context(),
			organizationID,
			filter,
		)
		if err != nil {
			writeError(c, err)
			return
		}

		if result == nil || len(result.Payments) == 0 {
			break
		}

		payments = append(payments, result.Payments...)

		if len(result.Payments) < batchSize {
			break
		}

		offset += batchSize
	}

	data := exports.BuildWorkerPayments(payments)

	data.Filename = fmt.Sprintf(
		"worker-payments-%s",
		time.Now().Format("20060102-150405"),
	)

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

func (h *WorkerPaymentHandler) ExportPendingBalances(c *gin.Context) {
	organizationID := c.Param("organizationID")
	if organizationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "organization id is required",
		})
		return
	}

	format := strings.ToLower(strings.TrimSpace(c.DefaultQuery("format", "xlsx")))

	switch format {
	case "xlsx", "csv", "pdf":
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "unsupported export format",
		})
		return
	}

	fromDate, err := parseDate(c, "from_date")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	toDate, err := parseDate(c, "to_date")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if fromDate != nil && toDate != nil && toDate.Before(*fromDate) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "to_date must be on or after from_date",
		})
		return
	}

	// Use an exclusive upper bound so the entire to_date is included.
	if toDate != nil {
		endDate := toDate.AddDate(0, 0, 1)
		toDate = &endDate
	}

	filter := input.WorkerPaymentBalanceFilter{
		Search:           c.Query("search"),
		BundleID:         c.Query("bundle_id"),
		WorkerID:         c.Query("worker_id"),
		StyleID:          c.Query("style_id"),
		StyleOperationID: c.Query("style_operation_id"),
		FromDate:         fromDate,
		ToDate:           toDate,
		Limit:            100,
		Offset:           0,
	}

	workers := make([]*output.WorkerPaymentBalance, 0)

	for {
		result, err := h.payments.ListWorkerBalances(
			c.Request.Context(),
			organizationID,
			filter,
		)
		if err != nil {
			writeError(c, err)
			return
		}

		if result == nil || len(result.Workers) == 0 {
			break
		}

		workers = append(workers, result.Workers...)

		if len(result.Workers) < filter.Limit {
			break
		}

		filter.Offset += filter.Limit
	}

	data := exports.BuildPendingWorkerBalances(workers)

	filename := fmt.Sprintf(
		"pending-worker-payments-%s",
		time.Now().Format("20060102-150405"),
	)

	data.Filename = filename

	file, contentType, err := exporter.Render(format, data)
	if err != nil {
		writeError(c, err)
		return
	}

	c.Header("Content-Type", contentType)
	c.Header(
		"Content-Disposition",
		fmt.Sprintf(`attachment; filename="%s.%s"`, filename, format),
	)

	c.Data(http.StatusOK, contentType, file)
}
