package exports

import (
	"stitchforce-server/internal/usecase/output"
	"stitchforce-server/pkg/exporter"
)

func BuildWorkerPayments(
	payments []*output.WorkerPayment,
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
