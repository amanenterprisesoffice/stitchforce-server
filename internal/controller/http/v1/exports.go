package v1

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"stitchforce-server/internal/usecase/output"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
)

func productionEntryExportHeaders() []string {
	return []string{
		"Production Date",
		"Worker",
		"Phone",
		"Bundle Number",
		"Style",
		"Style Number",
		"Operation",
		"Quantity",
		"Rate Per Piece (INR)",
		"Total Amount (INR)",
		"Bundle Status",
		"Remarks",
	}
}

func productionEntryExportRow(entry *output.ProductionEntry) []interface{} {
	return []interface{}{
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
	}
}

func exportProductionEntriesCSV(
	c *gin.Context,
	entries []*output.ProductionEntry,
	filename string,
) error {
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header(
		"Content-Disposition",
		fmt.Sprintf(`attachment; filename="%s.csv"`, filename),
	)

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	if err := writer.Write(productionEntryExportHeaders()); err != nil {
		return err
	}

	for _, entry := range entries {
		row := productionEntryExportRow(entry)

		record := make([]string, len(row))

		for i, value := range row {
			record[i] = fmt.Sprint(value)
		}

		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return writer.Error()
}

func exportProductionEntriesExcel(
	c *gin.Context,
	entries []*output.ProductionEntry,
	filename string,
) error {
	file := excelize.NewFile()
	defer file.Close()

	sheet := "Production Entries"

	index, err := file.NewSheet(sheet)
	if err != nil {
		return err
	}

	file.SetActiveSheet(index)

	headers := productionEntryExportHeaders()

	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		file.SetCellValue(sheet, cell, header)
	}

	headerStyle, err := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Color: "FFFFFF",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"2563EB"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		return err
	}

	lastHeaderCell, _ := excelize.CoordinatesToCellName(len(headers), 1)

	if err := file.SetCellStyle(
		sheet,
		"A1",
		lastHeaderCell,
		headerStyle,
	); err != nil {
		return err
	}

	for i, entry := range entries {
		rowNumber := i + 2
		values := productionEntryExportRow(entry)

		for j, value := range values {
			cell, _ := excelize.CoordinatesToCellName(j+1, rowNumber)

			if err := file.SetCellValue(sheet, cell, value); err != nil {
				return err
			}
		}
	}

	widths := []float64{
		18, 25, 16, 20, 25, 18,
		25, 12, 22, 22, 16, 35,
	}

	for i, width := range widths {
		column, _ := excelize.ColumnNumberToName(i + 1)

		if err := file.SetColWidth(sheet, column, column, width); err != nil {
			return err
		}
	}

	file.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		XSplit:      0,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	})

	var buffer bytes.Buffer

	if err := file.Write(&buffer); err != nil {
		return err
	}

	c.Header(
		"Content-Type",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	)
	c.Header(
		"Content-Disposition",
		fmt.Sprintf(`attachment; filename="%s.xlsx"`, filename),
	)

	c.Data(http.StatusOK, "application/octet-stream", buffer.Bytes())

	return nil
}

func exportProductionEntriesPDF(
	c *gin.Context,
	entries []*output.ProductionEntry,
	filename string,
) error {
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.SetMargins(8, 10, 8)
	pdf.SetAutoPageBreak(true, 10)

	pdf.AddPage()

	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 10, "Production Entries Report", "", 1, "C", false, 0, "")

	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(
		0,
		7,
		fmt.Sprintf("Generated: %s", time.Now().Format("02 Jan 2006, 03:04 PM")),
		"",
		1,
		"L",
		false,
		0,
		"",
	)

	pdf.CellFormat(
		0,
		7,
		fmt.Sprintf("Total entries: %d", len(entries)),
		"",
		1,
		"L",
		false,
		0,
		"",
	)

	pdf.Ln(4)

	headers := []string{
		"Date",
		"Worker",
		"Bundle",
		"Style",
		"Operation",
		"Qty",
		"Rate (INR)",
		"Total (INR)",
	}

	widths := []float64{
		22, 35, 30, 35, 40, 15, 25, 27,
	}

	drawHeader := func() {
		pdf.SetFont("Arial", "B", 8)
		pdf.SetFillColor(37, 99, 235)
		pdf.SetTextColor(255, 255, 255)

		for i, header := range headers {
			pdf.CellFormat(
				widths[i],
				8,
				header,
				"1",
				0,
				"C",
				true,
				0,
				"",
			)
		}

		pdf.Ln(-1)

		pdf.SetFont("Arial", "", 8)
		pdf.SetTextColor(0, 0, 0)
	}

	drawHeader()

	for _, entry := range entries {
		if pdf.GetY()+7 > 190 {
			pdf.AddPage()
			drawHeader()
		}

		values := []string{
			entry.ProductionDate.Format("02/01/2006"),
			entry.Worker.Name,
			entry.Bundle.BundleNumber,
			entry.Style.Name,
			entry.Operation.Name,
			strconv.Itoa(entry.Quantity),
			fmt.Sprintf("%.2f", float64(entry.RatePerPiece)/100),
			fmt.Sprintf("%.2f", float64(entry.TotalAmount)/100),
		}

		for i, value := range values {
			align := "L"

			if i >= 5 {
				align = "R"
			}

			pdf.CellFormat(
				widths[i],
				7,
				value,
				"1",
				0,
				align,
				false,
				0,
				"",
			)
		}

		pdf.Ln(-1)
	}

	var buffer bytes.Buffer

	if err := pdf.Output(&buffer); err != nil {
		return err
	}

	c.Header("Content-Type", "application/pdf")
	c.Header(
		"Content-Disposition",
		fmt.Sprintf(`attachment; filename="%s.pdf"`, filename),
	)

	c.Data(http.StatusOK, "application/pdf", buffer.Bytes())

	return nil
}
