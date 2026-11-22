package report

import (
	"archive/zip"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"stitchforce-server/config"
	"stitchforce-server/internal/entity"
	"stitchforce-server/internal/repo"
	"stitchforce-server/internal/storage"
	appusecase "stitchforce-server/internal/usecase"
	"stitchforce-server/internal/usecase/input"
	"stitchforce-server/internal/usecase/output"
	"stitchforce-server/pkg/email"
	"stitchforce-server/pkg/logger"
	"stitchforce-server/pkg/utils"
)

const (
	archiveRetentionDays = 180
	archiveExpiryYears   = 1
	archiveContentType   = "application/zip"
)

type UseCase struct {
	reports       repo.ReportRepository
	archive       repo.ArchiveRepository
	users         appusecase.UserUseCase
	notifications appusecase.NotificationUseCase
	emailer       email.Sender
	storage       storage.Storage
	frontendCfg   config.FrontendConfig
	log           *logger.Logger
}

func New(
	reports repo.ReportRepository,
	archive repo.ArchiveRepository,
	users appusecase.UserUseCase,
	notifications appusecase.NotificationUseCase,
	emailer email.Sender,
	objectStorage storage.Storage,
	frontendCfg config.FrontendConfig,
	log *logger.Logger,
) *UseCase {
	return &UseCase{
		reports:       reports,
		archive:       archive,
		users:         users,
		notifications: notifications,
		emailer:       emailer,
		storage:       objectStorage,
		frontendCfg:   frontendCfg,
		log:           log,
	}
}

// STREAMING HELPERS

type countingWriter struct {
	writer io.Writer
	size   int64
}

func (w *countingWriter) Write(p []byte) (int, error) {
	n, err := w.writer.Write(p)
	w.size += int64(n)
	return n, err
}

// writeCSV streams rows directly into a ZIP entry.
// It never accumulates all records in memory.
func writeCSV[T any](
	zipWriter *zip.Writer,
	filename string,
	headers []string,
	stream func(func(*T) error) (int, error),
	toRow func(*T) []string,
) (int, error) {
	file, err := zipWriter.Create(filename)
	if err != nil {
		return 0, fmt.Errorf("create %s: %w", filename, err)
	}

	csvWriter := csv.NewWriter(file)

	if err := csvWriter.Write(headers); err != nil {
		return 0, fmt.Errorf("write %s header: %w", filename, err)
	}

	count, streamErr := stream(func(record *T) error {
		return csvWriter.Write(toRow(record))
	})

	csvWriter.Flush()

	if err := csvWriter.Error(); err != nil {
		return count, fmt.Errorf("flush %s: %w", filename, err)
	}

	if streamErr != nil {
		return count, fmt.Errorf("stream %s: %w", filename, streamErr)
	}

	return count, nil
}

// GENERATE ZIP CONTENT

func (u *UseCase) writeArchive(
	ctx context.Context,
	outputWriter io.Writer,
	organizationID string,
	year int,
	month time.Month,
	from, to time.Time,
) (map[string]int, error) {
	zipWriter := zip.NewWriter(outputWriter)

	counts := make(map[string]int)

	// 1. Production entries.
	count, err := writeCSV(
		zipWriter,
		"production_entries.csv",
		[]string{
			"id",
			"organization_id",
			"bundle_id",
			"worker_id",
			"style_operation_id",
			"quantity",
			"rate_per_piece",
			"total_amount",
			"production_date",
			"remarks",
			"created_at",
			"updated_at",
			"deleted_at",
		},
		func(write func(*entity.ProductionEntry) error) (int, error) {
			return u.archive.StreamProductionEntries(
				ctx,
				organizationID,
				from,
				to,
				write,
			)
		},
		func(entry *entity.ProductionEntry) []string {
			totalAmount := int64(entry.Quantity) * entry.RatePerPiece

			return []string{
				entry.ID,
				entry.OrganizationID,
				entry.BundleID,
				entry.WorkerID,
				entry.StyleOperationID,
				strconv.Itoa(entry.Quantity),
				strconv.FormatInt(entry.RatePerPiece, 10),
				strconv.FormatInt(totalAmount, 10),
				entry.ProductionDate.Format(time.RFC3339Nano),
				entry.Remarks,
				entry.CreatedAt.Format(time.RFC3339Nano),
				entry.UpdatedAt.Format(time.RFC3339Nano),
				formatNullableTime(entry.DeletedAt),
			}
		},
	)
	if err != nil {
		return nil, err
	}
	counts["production_entries"] = count

	// 2. Worker payments.
	count, err = writeCSV(
		zipWriter,
		"worker_payments.csv",
		[]string{
			"id",
			"organization_id",
			"worker_id",
			"amount",
			"payment_date",
			"payment_method",
			"reference",
			"remarks",
			"status",
			"created_at",
			"updated_at",
		},
		func(write func(*entity.WorkerPayment) error) (int, error) {
			return u.archive.StreamWorkerPayments(
				ctx,
				organizationID,
				from,
				to,
				write,
			)
		},
		func(payment *entity.WorkerPayment) []string {
			return []string{
				payment.ID,
				payment.OrganizationID,
				payment.WorkerID,
				strconv.FormatInt(payment.Amount, 10),
				payment.PaymentDate.Format(time.RFC3339Nano),
				string(payment.PaymentMethod),
				payment.Reference,
				payment.Remarks,
				string(payment.Status),
				payment.CreatedAt.Format(time.RFC3339Nano),
				payment.UpdatedAt.Format(time.RFC3339Nano),
			}
		},
	)
	if err != nil {
		return nil, err
	}
	counts["worker_payments"] = count

	// 3. Payment allocations.
	count, err = writeCSV(
		zipWriter,
		"worker_payment_allocations.csv",
		[]string{
			"id",
			"organization_id",
			"worker_payment_id",
			"production_entry_id",
			"amount",
			"created_at",
		},
		func(write func(*output.PaymentAllocation) error) (int, error) {
			return u.archive.StreamPaymentAllocations(
				ctx,
				organizationID,
				from,
				to,
				write,
			)
		},
		func(allocation *output.PaymentAllocation) []string {
			return []string{
				allocation.ID,
				allocation.OrganizationID,
				allocation.WorkerPaymentID,
				allocation.ProductionEntryID,
				strconv.FormatInt(allocation.Amount, 10),
				allocation.CreatedAt.Format(time.RFC3339Nano),
			}
		},
	)
	if err != nil {
		return nil, err
	}
	counts["worker_payment_allocations"] = count

	// 4. Completed bundles.
	count, err = writeCSV(
		zipWriter,
		"completed_bundles.csv",
		[]string{
			"id",
			"organization_id",
			"bundle_number",
			"style_id",
			"size_id",
			"quantity",
			"attributes",
			"status",
			"remarks",
			"created_at",
			"updated_at",
			"deleted_at",
		},
		func(write func(*entity.Bundle) error) (int, error) {
			return u.archive.StreamCompletedBundles(
				ctx,
				organizationID,
				from,
				to,
				write,
			)
		},
		func(bundle *entity.Bundle) []string {
			return []string{
				bundle.ID,
				bundle.OrganizationID,
				bundle.BundleNumber,
				bundle.StyleID,
				bundle.SizeID,
				strconv.Itoa(bundle.Quantity),
				string(bundle.Attributes),
				string(bundle.Status),
				bundle.Remarks,
				bundle.CreatedAt.Format(time.RFC3339Nano),
				bundle.UpdatedAt.Format(time.RFC3339Nano),
				formatNullableTime(bundle.DeletedAt),
			}
		},
	)
	if err != nil {
		return nil, err
	}
	counts["completed_bundles"] = count

	// 5. Manifest.
	manifest := struct {
		OrganizationID string         `json:"organization_id"`
		Year           int            `json:"year"`
		Month          int            `json:"month"`
		From           time.Time      `json:"from"`
		To             time.Time      `json:"to"`
		GeneratedAt    time.Time      `json:"generated_at"`
		RecordCounts   map[string]int `json:"record_counts"`
	}{
		OrganizationID: organizationID,
		Year:           year,
		Month:          int(month),
		From:           from,
		To:             to,
		GeneratedAt:    time.Now(),
		RecordCounts:   counts,
	}

	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal archive manifest: %w", err)
	}

	manifestFile, err := zipWriter.Create("manifest.json")
	if err != nil {
		return nil, fmt.Errorf("create archive manifest: %w", err)
	}

	if _, err := manifestFile.Write(manifestBytes); err != nil {
		return nil, fmt.Errorf("write archive manifest: %w", err)
	}

	// Closing the ZIP writes its central directory.
	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("close archive ZIP: %w", err)
	}

	return counts, nil
}

func formatNullableTime(value *time.Time) string {
	if value == nil {
		return ""
	}

	return value.Format(time.RFC3339Nano)
}

func hasRecords(counts map[string]int) bool {
	for _, count := range counts {
		if count > 0 {
			return true
		}
	}

	return false
}

// STREAM ZIP DIRECTLY TO OBJECT STORAGE

func (u *UseCase) uploadArchive(
	ctx context.Context,
	organizationID string,
	year int,
	month time.Month,
	from, to time.Time,
	key string,
) (map[string]int, int64, error) {
	reader, writer := io.Pipe()

	counter := &countingWriter{
		writer: writer,
	}

	type generationResult struct {
		counts map[string]int
		err    error
	}

	generationDone := make(chan generationResult, 1)

	go func() {
		counts, err := u.writeArchive(
			ctx,
			counter,
			organizationID,
			year,
			month,
			from,
			to,
		)

		if err != nil {
			_ = writer.CloseWithError(err)
		} else {
			_ = writer.Close()
		}

		generationDone <- generationResult{
			counts: counts,
			err:    err,
		}
	}()

	uploadErr := u.storage.UploadStream(
		ctx,
		key,
		reader,
		archiveContentType,
	)

	if uploadErr != nil {
		// Unblock the ZIP writer if the storage upload fails.
		_ = reader.CloseWithError(uploadErr)
	} else {
		_ = reader.Close()
	}

	generated := <-generationDone

	if uploadErr != nil {
		return nil, 0, fmt.Errorf("upload archive: %w", uploadErr)
	}

	if generated.err != nil {
		return nil, 0, fmt.Errorf("generate archive: %w", generated.err)
	}

	return generated.counts, counter.size, nil
}

// MONTHLY ARCHIVE GENERATION

func (u *UseCase) GenerateMonthlyReports(
	ctx context.Context,
	now time.Time,
) (int, error) {
	organizations, err := u.archive.ListEligibleOrganizations(ctx)
	if err != nil {
		return 0, fmt.Errorf("list archive-eligible organizations: %w", err)
	}

	generated := 0
	var generationErrors []error

	for _, organization := range organizations {
		if organization == nil {
			continue
		}

		count, err := u.generateOrganizationArchives(
			ctx,
			organization,
			now,
		)
		generated += count

		if err != nil {
			generationErrors = append(
				generationErrors,
				fmt.Errorf(
					"organization %s: %w",
					organization.ID,
					err,
				),
			)
		}
	}

	return generated, errors.Join(generationErrors...)
}

func (u *UseCase) generateOrganizationArchives(
	ctx context.Context,
	organization *output.ArchiveOrganization,
	now time.Time,
) (int, error) {
	location, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		return 0, fmt.Errorf("load archive timezone: %w", err)
	}

	now = now.In(location)
	cutoff := now.Add(-archiveRetentionDays * 24 * time.Hour)

	createdAt := organization.CreatedAt.In(location)

	firstMonth := time.Date(
		createdAt.Year(),
		createdAt.Month(),
		1,
		0, 0, 0, 0,
		location,
	)

	// Only fully completed months whose end is older than the
	// retention cutoff are eligible.
	eligibleTo := time.Date(
		cutoff.Year(),
		cutoff.Month(),
		1,
		0, 0, 0, 0,
		location,
	)

	if !firstMonth.Before(eligibleTo) {
		return 0, nil
	}

	// Load existing reports once.
	existingReports, err := u.reports.List(ctx, organization.ID, input.ReportFilter{})
	if err != nil {
		return 0, fmt.Errorf("list existing reports: %w", err)
	}

	existing := make(map[string]*entity.Report)

	for _, report := range existingReports {
		if report == nil || report.Type != entity.ReportMonthlyArchive {
			continue
		}

		key := fmt.Sprintf("%04d-%02d", report.Year, report.Month)
		existing[key] = report
	}

	// One database query discovers every month containing archive data.
	archivedPeriods, err := u.archive.ListArchivedPeriods(
		ctx,
		organization.ID,
		firstMonth,
		eligibleTo,
	)
	if err != nil {
		return 0, fmt.Errorf("list archived periods: %w", err)
	}

	// Merge periods that need cleanup retries.
	// These must be included even if their source data was already
	// deleted but marking cleanup as completed previously failed.
	periodMap := make(map[string]output.ArchivePeriod)

	for _, period := range archivedPeriods {
		key := fmt.Sprintf("%04d-%02d", period.Year, period.Month)
		periodMap[key] = period
	}

	for _, report := range existing {
		if report.CleanupCompleted {
			continue
		}

		if report.Month < 1 || report.Month > 12 {
			continue
		}

		monthStart := time.Date(
			report.Year,
			time.Month(report.Month),
			1,
			0, 0, 0, 0,
			location,
		)

		if monthStart.Before(firstMonth) || !monthStart.Before(eligibleTo) {
			continue
		}

		key := fmt.Sprintf("%04d-%02d", report.Year, report.Month)

		periodMap[key] = output.ArchivePeriod{
			Year:  report.Year,
			Month: report.Month,
		}
	}

	// Convert the map to a sorted list.
	periods := make([]output.ArchivePeriod, 0, len(periodMap))

	for _, period := range periodMap {
		periods = append(periods, period)
	}

	sort.Slice(periods, func(i, j int) bool {
		if periods[i].Year != periods[j].Year {
			return periods[i].Year < periods[j].Year
		}

		return periods[i].Month < periods[j].Month
	})

	generated := 0
	var archiveErrors []error

	for _, period := range periods {
		year := period.Year
		month := time.Month(period.Month)

		from := time.Date(
			year,
			month,
			1,
			0, 0, 0, 0,
			location,
		)

		to := from.AddDate(0, 1, 0)

		periodKey := fmt.Sprintf("%04d-%02d", year, int(month))

		// Existing report: retry cleanup if necessary.
		if existingReport, found := existing[periodKey]; found {
			if existingReport.CleanupCompleted {
				continue
			}

			if err := u.archive.DeleteArchivedData(
				ctx,
				organization.ID,
				from,
				to,
			); err != nil {
				archiveErrors = append(
					archiveErrors,
					fmt.Errorf("retry cleanup for %s: %w", periodKey, err),
				)
				continue
			}

			if err := u.reports.MarkCleanupCompleted(
				ctx,
				organization.ID,
				existingReport.ID,
			); err != nil {
				archiveErrors = append(
					archiveErrors,
					fmt.Errorf("mark cleanup completed for %s: %w", periodKey, err),
				)
				continue
			}

			existingReport.CleanupCompleted = true
			u.notifyReportReady(ctx, organization.ID, existingReport)

			continue
		}

		// No existing report: generate the archive.
		created, err := u.generateArchive(
			ctx,
			organization.ID,
			year,
			month,
			from,
			to,
			now,
		)
		if err != nil {
			archiveErrors = append(
				archiveErrors,
				fmt.Errorf("generate archive for %s: %w", periodKey, err),
			)
			continue
		}

		if created {
			generated++
		}
	}

	return generated, errors.Join(archiveErrors...)
}

func (u *UseCase) generateArchive(
	ctx context.Context,
	organizationID string,
	year int,
	month time.Month,
	from, to time.Time,
	now time.Time,
) (bool, error) {
	key := fmt.Sprintf(
		"organizations/%s/archives/%04d/%02d/archive.zip",
		organizationID,
		year,
		int(month),
	)

	// 1. Generate and stream ZIP to object storage.
	counts, fileSize, err := u.uploadArchive(
		ctx,
		organizationID,
		year,
		month,
		from,
		to,
		key,
	)
	if err != nil {
		return false, err
	}

	// Avoid creating empty monthly reports.
	if !hasRecords(counts) {
		if err := u.storage.Delete(ctx, key); err != nil {
			return false, fmt.Errorf("delete empty archive: %w", err)
		}

		return false, nil
	}

	// 2. Save metadata after the object upload succeeds.
	reportID, err := utils.NewID()
	if err != nil {
		_ = u.storage.Delete(ctx, key)
		return false, fmt.Errorf("generate report ID: %w", err)
	}

	report := &entity.Report{
		ID:             reportID,
		OrganizationID: organizationID,
		Type:           entity.ReportMonthlyArchive,
		Year:           year,
		Month:          int(month),
		FileKey:        key,
		FileSize:       fileSize,
		RecordCounts:   counts,
		ArchivedAt:     now,
		ExpiresAt:      now.AddDate(0, 0, 0).AddDate(archiveExpiryYears, 0, 0),
	}

	if err := u.reports.Create(ctx, report); err != nil {
		// Keep original database records. Remove the uploaded object
		// because its metadata could not be persisted.
		if deleteErr := u.storage.Delete(ctx, key); deleteErr != nil {
			return false, errors.Join(
				fmt.Errorf("create report metadata: %w", err),
				fmt.Errorf("remove unregistered archive: %w", deleteErr),
			)
		}

		return false, fmt.Errorf("create report metadata: %w", err)
	}

	// 3. Delete source records only after upload and metadata succeed.
	if err := u.archive.DeleteArchivedData(
		ctx,
		organizationID,
		from,
		to,
	); err != nil {
		// The report exists. The next run retries only the cleanup.
		return true, fmt.Errorf("delete archived source data: %w", err)
	}

	if err := u.reports.MarkCleanupCompleted(ctx, organizationID, report.ID); err != nil {
		return true, fmt.Errorf("mark archive cleanup completed: %w", err)
	}
	report.CleanupCompleted = true

	// Notify users only after source cleanup is complete.
	u.notifyReportReady(ctx, organizationID, report)

	return true, nil
}

// NOTIFICATIONS

func (u *UseCase) notifyReportReady(
	ctx context.Context,
	organizationID string,
	report *entity.Report,
) {
	users, err := u.users.List(
		ctx,
		organizationID,
		input.UserFilter{},
	)
	if err != nil {
		u.logArchiveError(ctx, "list users for report notification", err)
		return
	}

	notifications := make([]*entity.Notification, 0, len(users))
	for _, user := range users {
		if user == nil {
			continue
		}

		notificationID, err := utils.NewID()
		if err != nil {
			u.logArchiveError(ctx, "generate notification ID", err)
			return
		}

		notifications = append(notifications, &entity.Notification{
			ID:             notificationID,
			OrganizationID: organizationID,
			UserID:         user.ID,
			Type:           entity.NotificationReportReady,
			Title:          "Monthly archive is ready",
			Message: fmt.Sprintf(
				"Your archive for %s %d is ready to download.",
				time.Month(report.Month).String(),
				report.Year,
			),
			Data: map[string]any{
				"report_id": report.ID,
				"year":      report.Year,
				"month":     report.Month,
			},
			ActionURL: "/reports",
		})
	}

	if len(notifications) == 0 {
		return
	}

	if err := u.notifications.CreateBulk(ctx, notifications); err != nil {
		u.logArchiveError(ctx, "create report notifications", err)
	}
}

// EXPIRED REPORT CLEANUP

func (u *UseCase) DeleteExpiredReports(
	ctx context.Context,
	now time.Time,
) (int, error) {
	reports, err := u.reports.ListExpired(ctx, now)
	if err != nil {
		return 0, fmt.Errorf("list expired reports: %w", err)
	}

	deleted := 0
	var deletionErrors []error

	for _, report := range reports {
		if report == nil {
			continue
		}

		// Delete the object first.
		if err := u.storage.Delete(ctx, report.FileKey); err != nil {
			deletionErrors = append(
				deletionErrors,
				fmt.Errorf(
					"delete report object %s: %w",
					report.ID,
					err,
				),
			)
			continue
		}

		// Only remove metadata after object deletion succeeds.
		if err := u.reports.Delete(
			ctx,
			report.OrganizationID,
			report.ID,
		); err != nil {
			deletionErrors = append(
				deletionErrors,
				fmt.Errorf(
					"delete report metadata %s: %w",
					report.ID,
					err,
				),
			)
			continue
		}

		deleted++
	}

	return deleted, errors.Join(deletionErrors...)
}

// REPORT CRUD

func (u *UseCase) Create(
	ctx context.Context,
	report *entity.Report,
) error {
	return u.reports.Create(ctx, report)
}

func (u *UseCase) List(
	ctx context.Context,
	organizationID string,
	filter input.ReportFilter,
) ([]*entity.Report, error) {
	return u.reports.List(ctx, organizationID, filter)
}

func (u *UseCase) GetByID(
	ctx context.Context,
	organizationID string,
	reportID string,
) (*entity.Report, error) {
	return u.reports.GetByID(ctx, organizationID, reportID)
}

func (u *UseCase) GetDownloadURL(
	ctx context.Context,
	organizationID string,
	reportID string,
) (string, error) {
	report, err := u.reports.GetByID(
		ctx,
		organizationID,
		reportID,
	)
	if err != nil {
		return "", err
	}

	return u.storage.SignedURL(
		ctx,
		report.FileKey,
		24*time.Hour,
	)
}

func (u *UseCase) Delete(
	ctx context.Context,
	organizationID string,
	reportID string,
) error {
	report, err := u.reports.GetByID(
		ctx,
		organizationID,
		reportID,
	)
	if err != nil {
		return err
	}

	if err := u.storage.Delete(ctx, report.FileKey); err != nil {
		return fmt.Errorf("delete report object: %w", err)
	}

	if err := u.reports.Delete(
		ctx,
		organizationID,
		reportID,
	); err != nil {
		return fmt.Errorf("delete report metadata: %w", err)
	}

	return nil
}

func (u *UseCase) SendExpirationReminders(
	ctx context.Context,
	now time.Time,
) (int, error) {
	reports, err := u.reports.ListExpiring(ctx, now)
	if err != nil {
		return 0, fmt.Errorf("list expiring reports: %w", err)
	}

	sent := 0
	var failures []error

	location, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		return 0, fmt.Errorf("load timezone: %w", err)
	}

	for _, report := range reports {
		if report == nil {
			continue
		}

		users, err := u.users.List(
			ctx,
			report.OrganizationID,
			input.UserFilter{
				Limit:  1000,
				Offset: 0,
			},
		)
		if err != nil {
			failures = append(failures, fmt.Errorf(
				"report %s: list organization users: %w",
				report.ID,
				err,
			))
			continue
		}

		recipients := make([]string, 0)
		notifications := make([]*entity.Notification, 0)

		expirationDate := report.ExpiresAt.In(location).Format("02 January 2006")
		period := fmt.Sprintf("%s %d", time.Month(report.Month), report.Year)

		for _, user := range users {
			if user == nil {
				continue
			}

			// Email reminders only go to organization administrators.
			if user.Role == entity.UserRoleOrgAdmin {
				address := strings.TrimSpace(user.Email)
				if address != "" {
					recipients = append(recipients, address)
				}
			}

			// In-app notifications go to every organization user.
			notifications = append(notifications, &entity.Notification{
				OrganizationID: report.OrganizationID,
				UserID:         user.ID,
				Type:           entity.NotificationSystem,
				Title:          "Archived data expiring soon",
				Message: fmt.Sprintf(
					"Your %s archive is scheduled for permanent deletion on %s. Download it before it expires.",
					period,
					expirationDate,
				),
				Data: map[string]any{
					"report_id": report.ID,
					"year":      report.Year,
					"month":     report.Month,
					"expires_at": report.ExpiresAt.Format(time.RFC3339),
				},
				ActionURL: "/reports/" + report.ID,
			})
		}

		if len(recipients) == 0 {
			failures = append(failures, fmt.Errorf(
				"report %s: no organization admin email recipients found",
				report.ID,
			))
			continue
		}

		downloadURL, err := u.GetDownloadURL(
			ctx,
			report.OrganizationID,
			report.ID,
		)
		if err != nil {
			failures = append(failures, fmt.Errorf(
				"report %s: get download URL: %w",
				report.ID,
				err,
			))
			continue
		}

		body, err := u.buildExpirationReminderEmail(report, downloadURL)
		if err != nil {
			failures = append(failures, fmt.Errorf(
				"report %s: build expiration email: %w",
				report.ID,
				err,
			))
			continue
		}

		message := email.Message{
			To:      recipients,
			Subject: "StitchForce - Your archived data will be deleted soon",
			HTML:    body,
		}

		if err := u.emailer.Send(ctx, message); err != nil {
			failures = append(failures, fmt.Errorf(
				"report %s: send expiration reminder: %w",
				report.ID,
				err,
			))
			continue
		}

		if len(notifications) > 0 {
			if err := u.notifications.CreateBulk(ctx, notifications); err != nil {
				failures = append(failures, fmt.Errorf(
					"report %s: create expiration notifications: %w",
					report.ID,
					err,
				))
				continue
			}
		}

		if err := u.reports.MarkExpirationReminderSent(
			ctx,
			report.OrganizationID,
			report.ID,
		); err != nil {
			failures = append(failures, fmt.Errorf(
				"report %s: mark expiration reminder sent: %w",
				report.ID,
				err,
			))
			continue
		}

		sent++
	}

	return sent, errors.Join(failures...)
}

// HELPERS

func (u *UseCase) logArchiveError(
	ctx context.Context,
	message string,
	err error,
) {
	if u.log == nil || err == nil {
		return
	}

	u.log.ErrorContext(ctx, message, "error", err)
}

func (u *UseCase) buildExpirationReminderEmail(
	report *entity.Report,
	downloadURL string,
) (string, error) {
	if report == nil {
		return "", errors.New("report is required")
	}

	if downloadURL == "" {
		return "", errors.New("download URL is required")
	}

	if u.frontendCfg.URL == "" {
		return "", errors.New("frontend URL is not configured")
	}

	location, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		return "", fmt.Errorf("load timezone: %w", err)
	}

	period := time.Month(report.Month).String() + " " +
		fmt.Sprintf("%d", report.Year)

	expirationDate := report.ExpiresAt.In(location).Format("02 January 2006")

	reportsURL := strings.TrimRight(u.frontendCfg.URL, "/") + "/reports"
	reportDetailsURL := reportsURL + "/" + report.ID

	escapedPeriod := html.EscapeString(period)
	escapedReportID := html.EscapeString(report.ID)
	escapedExpirationDate := html.EscapeString(expirationDate)
	escapedDownloadURL := html.EscapeString(downloadURL)
	escapedReportDetailsURL := html.EscapeString(reportDetailsURL)

	body := fmt.Sprintf(`
		<div>
			<div style="
				margin-bottom: 28px;
				padding-bottom: 20px;
				border-bottom: 1px solid #e4e4e7;
			">
				<h2 style="
					margin: 0;
					font-size: 24px;
					font-weight: 700;
				">
					Your archived data will be deleted soon
				</h2>

				<p style="
					margin: 8px 0 0;
					color: #71717a;
					font-size: 14px;
				">
					StitchForce - Garment Manufacturing Management System
				</p>
			</div>

			<p style="font-size: 15px; line-height: 1.7;">
				This is a reminder that your archived factory data
				is scheduled for permanent deletion soon.
			</p>

			<div style="
				margin: 24px 0;
				padding: 20px;
				background: #f4f4f5;
				border: 1px solid #e4e4e7;
				border-radius: 8px;
			">
				<p style="
					margin: 0 0 6px;
					color: #71717a;
					font-size: 12px;
					font-weight: 600;
					text-transform: uppercase;
				">
					Archive period
				</p>

				<h3 style="
					margin: 0;
					font-size: 20px;
					font-weight: 700;
				">
					%s
				</h3>

				<p style="
					margin: 8px 0 0;
					color: #71717a;
					font-size: 12px;
				">
					Report ID: %s
				</p>

				<div style="
					margin-top: 18px;
					padding-top: 14px;
					border-top: 1px solid #e4e4e7;
				">
					<p style="
						margin: 0 0 5px;
						color: #71717a;
						font-size: 12px;
						font-weight: 600;
						text-transform: uppercase;
					">
						Scheduled deletion date
					</p>

					<p style="
						margin: 0;
						font-size: 16px;
						font-weight: 600;
					">
						%s
					</p>
				</div>
			</div>

			<p style="font-size: 14px; line-height: 1.7;">
				Please download and securely store this archive
				if you need to retain the data.
			</p>

			<div style="margin: 28px 0; text-align: center;">
				<a href="%s"
					style="
						display: inline-block;
						padding: 13px 28px;
						background: #18181b;
						color: #ffffff;
						text-decoration: none;
						border-radius: 7px;
						font-size: 14px;
						font-weight: 600;
					">
					Download Archive
				</a>
			</div>

			<div style="
				margin: 20px 0;
				padding: 15px;
				background: #fefce8;
				border: 1px solid #fde68a;
				border-radius: 8px;
			">
				<p style="
					margin: 0;
					color: #854d0e;
					font-size: 13px;
					line-height: 1.7;
				">
					<strong>Important:</strong> This download link
					expires in 24 hours. If it expires, visit this
					report's details page to generate a fresh download link.
				</p>

				<p style="margin: 10px 0 0;">
					<a href="%s"
						style="
							color: #2563eb;
							font-size: 13px;
							font-weight: 600;
							text-decoration: underline;
						">
						View Report Details
					</a>
				</p>
			</div>

			<p style="
				color: #71717a;
				font-size: 13px;
				line-height: 1.6;
			">
				After the scheduled deletion date, this archive
				will no longer be available for download.
			</p>

			<div style="
				margin: 28px 0;
				text-align: center;
			">
				<a href="%s"
					style="
						display: inline-block;
						padding: 12px 26px;
						background: #ffffff;
						color: #18181b;
						border: 1px solid #d4d4d8;
						text-decoration: none;
						border-radius: 7px;
						font-size: 14px;
						font-weight: 600;
					">
					View Report
				</a>
			</div>

			<div style="
				margin-top: 30px;
				padding-top: 18px;
				border-top: 1px solid #e4e4e7;
				color: #a1a1aa;
				font-size: 12px;
				line-height: 1.6;
			">
				This is an automated data-retention reminder from StitchForce.
			</div>
		</div>
	`,
		escapedPeriod,
		escapedReportID,
		escapedExpirationDate,
		escapedDownloadURL,
		escapedReportDetailsURL,
		escapedReportDetailsURL,
	)

	return email.Render(body)
}