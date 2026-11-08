
package exporter

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
)

type Column struct {
	Header string
	Width  float64 // Relative width weight
	Align  string  // L, C, R
}

type Data struct {
	Title    string
	Filename string
	Columns  []Column
	Rows     [][]any
}

func (d Data) Validate() error {
	if len(d.Columns) == 0 {
		return errors.New("export columns are required")
	}

	if len(d.Rows) == 0 {
		// Empty reports are valid.
		return nil
	}

	for i, row := range d.Rows {
		if len(row) != len(d.Columns) {
			return fmt.Errorf(
				"row %d has %d values, expected %d",
				i+1,
				len(row),
				len(d.Columns),
			)
		}
	}

	return nil
}

func valueString(value any) string {
	if value == nil {
		return ""
	}

	return fmt.Sprint(value)
}

func columnAlign(align string) string {
	switch strings.ToUpper(align) {
	case "C", "CENTER":
		return "C"
	case "R", "RIGHT":
		return "R"
	default:
		return "L"
	}
}

// Render generates the requested export format.
func Render(format string, data Data) ([]byte, string, error) {
	if err := data.Validate(); err != nil {
		return nil, "", err
	}

	format = strings.ToLower(strings.TrimPrefix(format, "."))

	switch format {
	case "csv":
		result, err := renderCSV(data)
		return result, "text/csv; charset=utf-8", err

	case "xlsx", "excel":
		result, err := renderExcel(data)
		return result,
			"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			err

	case "pdf":
		result, err := renderPDF(data)
		return result, "application/pdf", err

	default:
		return nil, "", errors.New("unsupported export format")
	}
}

// =========================================
// CSV
// =========================================

func renderCSV(data Data) ([]byte, error) {
	var buffer bytes.Buffer

	writer := csv.NewWriter(&buffer)
	defer writer.Flush()

	headers := make([]string, len(data.Columns))

	for i, col := range data.Columns {
		headers[i] = col.Header
	}

	if err := writer.Write(headers); err != nil {
		return nil, err
	}

	for _, row := range data.Rows {
		record := make([]string, len(row))

		for i, value := range row {
			record[i] = valueString(value)
		}

		if err := writer.Write(record); err != nil {
			return nil, err
		}
	}

	writer.Flush()

	if err := writer.Error(); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// =========================================
// EXCEL
// =========================================

func renderExcel(data Data) ([]byte, error) {
	file := excelize.NewFile()
	defer file.Close()

	sheet := "Report"

	if _, err := file.NewSheet(sheet); err != nil {
		return nil, err
	}

	file.DeleteSheet("Sheet1")
	file.SetActiveSheet(0)

	// Header style.
	headerStyle, err := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Color: "FFFFFF",
			Size:  11,
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
		return nil, err
	}

	// Write headers.
	for i, col := range data.Columns {
		cell, err := excelize.CoordinatesToCellName(i+1, 1)
		if err != nil {
			return nil, err
		}

		if err := file.SetCellValue(sheet, cell, col.Header); err != nil {
			return nil, err
		}

		column, err := excelize.ColumnNumberToName(i + 1)
		if err != nil {
			return nil, err
		}

		// Calculate a reasonable width from content.
		maxLength := len([]rune(col.Header))

		for _, row := range data.Rows {
			length := len([]rune(valueString(row[i])))
			if length > maxLength {
				maxLength = length
			}
		}

		width := math.Max(12, math.Min(float64(maxLength+3), 45))

		if err := file.SetColWidth(
			sheet,
			column,
			column,
			width,
		); err != nil {
			return nil, err
		}
	}

	lastHeaderCell, err := excelize.CoordinatesToCellName(
		len(data.Columns),
		1,
	)
	if err != nil {
		return nil, err
	}

	if err := file.SetCellStyle(
		sheet,
		"A1",
		lastHeaderCell,
		headerStyle,
	); err != nil {
		return nil, err
	}

	if err := file.SetRowHeight(sheet, 1, 24); err != nil {
		return nil, err
	}

	// Write data.
	for i, row := range data.Rows {
		rowNumber := i + 2

		for j, value := range row {
			cell, err := excelize.CoordinatesToCellName(
				j+1,
				rowNumber,
			)
			if err != nil {
				return nil, err
			}

			if err := file.SetCellValue(
				sheet,
				cell,
				value,
			); err != nil {
				return nil, err
			}
		}
	}

	// Freeze header.
	if err := file.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	}); err != nil {
		return nil, err
	}

	var buffer bytes.Buffer

	if err := file.Write(&buffer); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// =========================================
// PDF
// =========================================

func renderPDF(data Data) ([]byte, error) {
	pdf := gofpdf.New("L", "mm", "A4", "")

	pdf.SetMargins(8, 10, 8)
	pdf.SetAutoPageBreak(true, 10)

	pdf.AddPage()

	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(
		0,
		10,
		data.Title,
		"",
		1,
		"C",
		false,
		0,
		"",
	)

	pdf.SetFont("Arial", "", 8)
	pdf.CellFormat(
		0,
		6,
		"Generated report",
		"",
		1,
		"L",
		false,
		0,
		"",
	)

	pdf.Ln(3)

	pageWidth, _ := pdf.GetPageSize()
	leftMargin, _, rightMargin, _ := pdf.GetMargins()

	availableWidth := pageWidth - leftMargin - rightMargin

	// Calculate proportional column widths.
	totalWeight := 0.0

	for _, col := range data.Columns {
		weight := col.Width
		if weight <= 0 {
			weight = 1
		}
		totalWeight += weight
	}

	widths := make([]float64, len(data.Columns))

	for i, col := range data.Columns {
		weight := col.Width
		if weight <= 0 {
			weight = 1
		}

		widths[i] = availableWidth * weight / totalWeight
	}

	const (
		fontSize = 7.5
		lineH    = 4.2
		padding  = 1.5
	)

	drawHeader := func() {
		pdf.SetFont("Arial", "B", fontSize)
		pdf.SetFillColor(37, 99, 235)
		pdf.SetTextColor(255, 255, 255)

		for i, col := range data.Columns {
			x, y := pdf.GetXY()

			pdf.Rect(
				x,
				y,
				widths[i],
				8,
				"F",
			)

			pdf.SetXY(x+padding, y+2)

			pdf.CellFormat(
				widths[i]-2*padding,
				4,
				col.Header,
				"",
				0,
				"C",
				false,
				0,
				"",
			)

			pdf.SetXY(x+widths[i], y)
		}

		pdf.Ln(8)

		pdf.SetFont("Arial", "", fontSize)
		pdf.SetTextColor(0, 0, 0)
	}

	drawHeader()

	for _, row := range data.Rows {
		lines := make([][]string, len(row))
		maxLines := 1

		// Calculate wrapped text for every cell.
		for i, value := range row {
			text := valueString(value)

			lines[i] = pdf.SplitText(
				text,
				widths[i]-2*padding,
			)

			if len(lines[i]) > maxLines {
				maxLines = len(lines[i])
			}
		}

		rowHeight := float64(maxLines)*lineH + 2*padding

		// New page when required.
		if pdf.GetY()+rowHeight > 190 {
			pdf.AddPage()
			drawHeader()
		}

		startX := leftMargin
		startY := pdf.GetY()

		for i, cellLines := range lines {
			cellWidth := widths[i]

			pdf.Rect(
				startX,
				startY,
				cellWidth,
				rowHeight,
				"D",
			)

			for lineIndex, line := range cellLines {
				textWidth := pdf.GetStringWidth(line)

				x := startX + padding

				switch columnAlign(data.Columns[i].Align) {
				case "C":
					x = startX + (cellWidth-textWidth)/2
				case "R":
					x = startX + cellWidth - padding - textWidth
				}

				y := startY + padding +
					float64(lineIndex+1)*lineH - 1

				pdf.Text(x, y, line)
			}

			startX += cellWidth
		}

		pdf.SetXY(leftMargin, startY+rowHeight)
	}

	var buffer bytes.Buffer

	if err := pdf.Output(&buffer); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// =========================================
// HTTP RESPONSE
// =========================================

func ContentDisposition(filename, format string) string {
	filename = strings.TrimSpace(filename)
	filename = strings.ReplaceAll(filename, `"`, "")

	if filename == "" {
		filename = "export"
	}

	extension := strings.ToLower(
		strings.TrimPrefix(format, "."),
	)

	if extension == "excel" {
		extension = "xlsx"
	}

	return fmt.Sprintf(
		`attachment; filename="%s.%s"`,
		filename,
		extension,
	)
}

// Prevent accidental unused import changes when adapting formats.
var _ io.Writer
var _ = strconv.Itoa
