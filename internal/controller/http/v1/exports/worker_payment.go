package exports

import (
	"stitchforce-server/internal/usecase/output"
	"stitchforce-server/pkg/exporter"
)

func BuildWorkerPayments(
	payments []*output.SlimWorkerPayment,
) exporter.Data {
	data := exporter.Data{
		Title:    "Worker Payments Report",
		Filename: "worker-payments",

		Columns: []exporter.Column{
			{Header: "Payment Date", Width: 18},
			{Header: "Worker", Width: 25},
			{Header: "Phone", Width: 18},
			{Header: "Payment Method", Width: 18},
			{Header: "Reference", Width: 25},
			{Header: "Amount (INR)", Width: 20, Align: "R"},
			{Header: "Status", Width: 15},
			{Header: "Remarks", Width: 35},
		},

		Rows: make([][]any, 0, len(payments)),
	}

	for _, payment := range payments {
		if payment == nil {
			continue
		}

		data.Rows = append(data.Rows, []any{
			payment.PaymentDate.Format("02 Jan 2006"),
			payment.Worker.Name,
			payment.Worker.Phone,
			string(payment.PaymentMethod),
			payment.Reference,
			float64(payment.Amount) / 100,
			string(payment.Status),
			payment.Remarks,
		})
	}

	return data
}


func BuildPendingWorkerBalances(
	workers []*output.WorkerPaymentBalance,
) exporter.Data {
	data := exporter.Data{
		Title:    "Pending Worker Payments Report",
		Filename: "pending-worker-payments",

		Columns: []exporter.Column{
			{Header: "Worker", Width: 25},
			{Header: "Phone", Width: 18},
			{Header: "Total Earned (INR)", Width: 22, Align: "R"},
			{Header: "Total Paid (INR)", Width: 22, Align: "R"},
			{Header: "Pending Amount (INR)", Width: 25, Align: "R"},
			{Header: "Pending Entries", Width: 18, Align: "R"},
		},

		Rows: make([][]any, 0, len(workers)),
	}

	var totalEarned int64
	var totalPaid int64
	var totalPending int64
	var totalPendingEntries int

	for _, worker := range workers {
		if worker == nil {
			continue
		}

		data.Rows = append(data.Rows, []any{
			worker.Worker.Name,
			worker.Worker.Phone,
			float64(worker.TotalEarned) / 100,
			float64(worker.TotalPaid) / 100,
			float64(worker.TotalPending) / 100,
			worker.PendingEntries,
		})

		totalEarned += worker.TotalEarned
		totalPaid += worker.TotalPaid
		totalPending += worker.TotalPending
		totalPendingEntries += worker.PendingEntries
	}

	// Append a totals row.
	if len(data.Rows) > 0 {
		data.Rows = append(data.Rows, []any{
			"TOTAL",
			"",
			float64(totalEarned) / 100,
			float64(totalPaid) / 100,
			float64(totalPending) / 100,
			totalPendingEntries,
		})
	}

	return data
}