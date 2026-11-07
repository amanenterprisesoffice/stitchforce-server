package workerpayment

import (
    "context"
    "errors"
    "fmt"
    "strings"
    "time"

    "stitchforce-server/internal/entity"
    "stitchforce-server/internal/repo"
    "stitchforce-server/internal/usecase"
    "stitchforce-server/internal/usecase/input"
    "stitchforce-server/internal/usecase/output"
    "stitchforce-server/internal/usecase/utils"
    "stitchforce-server/pkg/logger"
    appUtils "stitchforce-server/pkg/utils"
)

type UseCase struct {
    repository       repo.WorkerPaymentRepository
    workerRepository repo.GarmentWorkerRepository
    logger           *logger.Logger
}

func New(
    repository repo.WorkerPaymentRepository,
    workerRepository repo.GarmentWorkerRepository,
    logger *logger.Logger,
) usecase.WorkerPaymentUseCase {
    return &UseCase{
        repository:       repository,
        workerRepository: workerRepository,
        logger:           logger,
    }
}

// CREATE
func (u *UseCase) Create(
    ctx context.Context,
    organizationID string,
    data *input.CreateWorkerPaymentInput,
) error {
    organizationID, err := utils.ValidateOrganizationID(organizationID)
    if err != nil {
        return err
    }

    if data == nil {
        return errors.New("worker payment is required")
    }

    data.WorkerID = strings.TrimSpace(data.WorkerID)
    data.Reference = strings.TrimSpace(data.Reference)
    data.Remarks = strings.TrimSpace(data.Remarks)

    if data.WorkerID == "" {
        return errors.New("worker ID is required")
    }

    if data.FromDate.IsZero() || data.ToDate.IsZero() {
        return errors.New("payment date range is required")
    }

    if data.ToDate.Before(data.FromDate) {
        return errors.New("to date cannot be before from date")
    }

    if data.PaymentDate.IsZero() {
        return errors.New("payment date is required")
    }

    switch data.PaymentMethod {
    case entity.PaymentMethodCash,
        entity.PaymentMethodBank,
        entity.PaymentMethodUPI,
        entity.PaymentMethodCheque:
    default:
        return errors.New("invalid payment method")
    }

    // Validate worker exists.
    // Inactive workers can still receive payment for previously completed work.
    worker, err := u.workerRepository.GetByID(
        ctx,
        organizationID,
        data.WorkerID,
    )
    if err != nil {
        return fmt.Errorf("get payment worker: %w", err)
    }

    if worker == nil {
        return errors.New("garment worker not found")
    }

    // Generate payment ID.
    id, err := appUtils.NewID()
    if err != nil {
        return fmt.Errorf("generate worker payment ID: %w", err)
    }

    now := time.Now().UTC()

    payment := &entity.WorkerPayment{
        ID:             id,
        OrganizationID: organizationID,
        WorkerID:       data.WorkerID,
        PaymentDate:    data.PaymentDate,
        PaymentMethod:  data.PaymentMethod,
        Reference:      data.Reference,
        Remarks:        data.Remarks,
        Status:         entity.PaymentStatusCompleted,
        CreatedAt:      now,
        UpdatedAt:      now,
    }

    filter := input.PayableEntriesFilter{
        FromDate: data.FromDate,
        ToDate:   data.ToDate,
    }

    u.logger.InfoContext(
        ctx,
        "creating worker payment",
        "organization_id", organizationID,
        "worker_id", data.WorkerID,
        "from_date", data.FromDate,
        "to_date", data.ToDate,
        "payment_date", data.PaymentDate,
        "payment_method", data.PaymentMethod,
    )

    // Repository must perform this atomically:
    // 1. Lock the worker's eligible production entries.
    // 2. Calculate each entry's remaining payable amount.
    // 3. Calculate the total payment amount.
    // 4. Reject if there is no outstanding amount.
    // 5. Insert the payment.
    // 6. Insert allocations for the outstanding amounts.
    // 7. Commit the transaction.
    if err := u.repository.CreateWithAllocations(
        ctx,
        organizationID,
        payment,
        filter,
    ); err != nil {
        return fmt.Errorf("create worker payment: %w", err)
    }

    return nil
}

// GET BY ID
func (u *UseCase) GetByID(
    ctx context.Context,
    organizationID string,
    paymentID string,
) (*entity.WorkerPayment, error) {
    organizationID, err := utils.ValidateOrganizationID(organizationID)
    if err != nil {
        return nil, err
    }

    paymentID = strings.TrimSpace(paymentID)
    if paymentID == "" {
        return nil, errors.New("payment ID is required")
    }

    payment, err := u.repository.GetByID(
        ctx,
        organizationID,
        paymentID,
    )
    if err != nil {
        return nil, fmt.Errorf("get worker payment: %w", err)
    }

    if payment == nil {
        return nil, errors.New("worker payment not found")
    }

    return payment, nil
}

// LIST
func (u *UseCase) List(
    ctx context.Context,
    organizationID string,
    filter input.WorkerPaymentFilter,
) ([]*entity.WorkerPayment, error) {
    organizationID, err := utils.ValidateOrganizationID(organizationID)
    if err != nil {
        return nil, err
    }

    filter.WorkerID = strings.TrimSpace(filter.WorkerID)
    filter.Status = strings.TrimSpace(filter.Status)

    if filter.Limit <= 0 {
        filter.Limit = 20
    }

    if filter.Limit > 100 {
        filter.Limit = 100
    }

    if filter.Offset < 0 {
        filter.Offset = 0
    }

    payments, err := u.repository.List(
        ctx,
        organizationID,
        filter,
    )
    if err != nil {
        return nil, fmt.Errorf("list worker payments: %w", err)
    }

    return payments, nil
}

// GET WORKER PAYMENT SUMMARY
func (u *UseCase) GetSummary(
    ctx context.Context,
    organizationID string,
    workerID string,
) (*output.WorkerPaymentSummary, error) {
    organizationID, err := utils.ValidateOrganizationID(organizationID)
    if err != nil {
        return nil, err
    }

    workerID = strings.TrimSpace(workerID)
    if workerID == "" {
        return nil, errors.New("worker ID is required")
    }

    worker, err := u.workerRepository.GetByID(
        ctx,
        organizationID,
        workerID,
    )
    if err != nil {
        return nil, fmt.Errorf("get payment worker: %w", err)
    }

    if worker == nil {
        return nil, errors.New("garment worker not found")
    }

    summary, err := u.repository.GetSummary(
        ctx,
        organizationID,
        workerID,
    )
    if err != nil {
        return nil, fmt.Errorf("get worker payment summary: %w", err)
    }

    return summary, nil
}

// GET PAYABLE PRODUCTION ENTRIES
func (u *UseCase) GetPayableEntries(
    ctx context.Context,
    organizationID string,
    workerID string,
    filter input.PayableEntriesFilter,
) ([]*output.PayableProductionEntry, error) {
    organizationID, err := utils.ValidateOrganizationID(organizationID)
    if err != nil {
        return nil, err
    }

    workerID = strings.TrimSpace(workerID)
    if workerID == "" {
        return nil, errors.New("worker ID is required")
    }

    if filter.FromDate.IsZero() || filter.ToDate.IsZero() {
        return nil, errors.New("date range is required")
    }

    if filter.ToDate.Before(filter.FromDate) {
        return nil, errors.New("to date cannot be before from date")
    }

    worker, err := u.workerRepository.GetByID(
        ctx,
        organizationID,
        workerID,
    )
    if err != nil {
        return nil, fmt.Errorf("get payment worker: %w", err)
    }

    if worker == nil {
        return nil, errors.New("garment worker not found")
    }

    entries, err := u.repository.ListPayableEntries(
        ctx,
        organizationID,
        workerID,
        filter,
    )
    if err != nil {
        return nil, fmt.Errorf("list payable production entries: %w", err)
    }

    return entries, nil
}

// VOID PAYMENT
func (u *UseCase) Void(
    ctx context.Context,
    organizationID string,
    paymentID string,
) error {
    organizationID, err := utils.ValidateOrganizationID(organizationID)
    if err != nil {
        return err
    }

    paymentID = strings.TrimSpace(paymentID)
    if paymentID == "" {
        return errors.New("payment ID is required")
    }

    payment, err := u.repository.GetByID(
        ctx,
        organizationID,
        paymentID,
    )
    if err != nil {
        return fmt.Errorf("get worker payment: %w", err)
    }

    if payment == nil {
        return errors.New("worker payment not found")
    }

    if payment.Status == entity.PaymentStatusVoided {
        return errors.New("worker payment is already voided")
    }

    if err := u.repository.Void(
        ctx,
        organizationID,
        paymentID,
    ); err != nil {
        return fmt.Errorf("void worker payment: %w", err)
    }

    return nil
}