package exports

import (
	"fmt"
	"net/http"
	"stitchforce-server/internal/usecase/output"
	"stitchforce-server/pkg/exporter"

	"github.com/gin-gonic/gin"
)

func BuildProductionEntryExport(
	entries []*output.ProductionEntry,
) exporter.Data {
	return exporter.Data{
		Title:    "Production Entries Report",
		Filename: "production_entries",

		Columns: []exporter.Column{
			{Header: "Production Date", Width: 18},
			{Header: "Worker", Width: 25},
			{Header: "Phone", Width: 16},
			{Header: "Bundle Number", Width: 20},
			{Header: "Style", Width: 25},
			{Header: "Style Number", Width: 18},
			{Header: "Operation", Width: 25},
			{Header: "Quantity", Width: 12, Align: "R"},
			{Header: "Rate Per Piece (INR)", Width: 22, Align: "R"},
			{Header: "Total Amount (INR)", Width: 22, Align: "R"},
			{Header: "Bundle Status", Width: 16},
			{Header: "Remarks", Width: 35},
		},

		Rows: ProductionEntryExportRows(entries),
	}
}

func ProductionEntryExportRows(
	entries []*output.ProductionEntry,
) [][]any {
	rows := make([][]any, 0, len(entries))

	for _, entry := range entries {
		rows = append(rows, []any{
			entry.ProductionDate.Format("02 Jan 2006"),
			entry.Worker.Name,
			entry.Worker.Phone,
			entry.Bundle.BundleNumber,
			entry.Style.Name,
			entry.Style.StyleNumber,
			entry.Operation.Name,
			entry.Quantity,
			float64(entry.RatePerPiece) / 100,
			float64(entry.TotalAmount) / 100,
			entry.Bundle.Status,
			entry.Remarks,
		})
	}

	return rows
}

// Export production entries.
func ExportProductionEntries(
	c *gin.Context,
	entries []*output.ProductionEntry,
) {
	format := c.Param("format")

	data := BuildProductionEntryExport(entries)

	content, contentType, err := exporter.Render(format, data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("export production entries: %v", err),
		})
		return
	}

	c.Header("Content-Type", contentType)
	c.Header(
		"Content-Disposition",
		exporter.ContentDisposition(data.Filename, format),
	)

	c.Data(http.StatusOK, contentType, content)
}
