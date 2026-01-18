package output

import (
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"

	"github.com/joshiste/sma_chg_log/internal/models"
)

const (
	lineSpacing    = 1.15
	dateFormat     = "02.01.2006"
	dateTimeFormat = "02.01.2006 15:04:05"
	headerHeight   = 16.0
	rowHeight      = 16.0
	bodyFontSize   = 10
)

var (
	colWidths  = []float64{25, 30, 45, 45, 45}
	colHeaders = []string{"Record Date", "Consumption (kWh)", "Charger", "Authentication", "Started at\nEnded at"}
)

// PDFFormatter outputs charging sessions
type PDFFormatter struct {
	writer   io.Writer
	sessions []models.ChargingSession
	opts     Options
}

// NewPDFFormatter creates a new PDF formatter
func NewPDFFormatter(w io.Writer) *PDFFormatter {
	return NewPDFFormatterWithOptions(w, Options{})
}

// NewPDFFormatterWithOptions creates a new PDF formatter with options
func NewPDFFormatterWithOptions(w io.Writer, opts Options) *PDFFormatter {
	return &PDFFormatter{
		writer:   w,
		sessions: make([]models.ChargingSession, 0),
		opts:     opts,
	}
}

// WriteHeader is a no-op for PDF format (header is written in Flush)
func (f *PDFFormatter) WriteHeader() error {
	return nil
}

// WriteSession buffers sessions for later PDF generation
func (f *PDFFormatter) WriteSession(session models.ChargingSession) error {
	f.sessions = append(f.sessions, session)
	return nil
}

// Flush generates the PDF document with summary and table
func (f *PDFFormatter) Flush() error {
	pdf := fpdf.New("P", "mm", "A4", "") // Portrait orientation

	// Set up footer with page numbers
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Arial", "", bodyFontSize)
		pageStr := fmt.Sprintf("Page %d of {nb}", pdf.PageNo())
		pdf.CellFormat(0, 10, pageStr, "", 0, "R", false, 0, "")
	})
	pdf.AliasNbPages("")

	pdf.AddPage()

	// Write title
	pdf.SetFont("Arial", "B", 20)
	pdf.Cell(0, 10*lineSpacing, "CHARGING HISTORY OVERVIEW")
	pdf.Ln(20) // More spacing below title

	// Write summary
	f.writeSummary(pdf)

	// Write table
	f.writeTable(pdf)

	return pdf.Output(f.writer)
}

// writeSummary writes the summary section with bold labels
func (f *PDFFormatter) writeSummary(pdf *fpdf.Fpdf) {
	lineHeight := 6 * lineSpacing
	pdf.SetFont("Arial", "", bodyFontSize)

	// Created On
	pdf.SetFontStyle("B")
	pdf.Cell(47, lineHeight, "Created On:")
	pdf.SetFontStyle("")
	pdf.Cell(0, lineHeight, time.Now().Format(dateFormat))
	pdf.Ln(lineHeight)

	// Overview Period
	pdf.SetFontStyle("B")
	pdf.Cell(47, lineHeight, "Overview Period:")
	pdf.SetFontStyle("")

	var fromDate, untilDate time.Time
	if !f.opts.From.IsZero() && !f.opts.Until.Equal(models.TimeMax) {
		fromDate = f.opts.From
		untilDate = f.opts.Until.AddDate(0, 0, -1) // until - 1 day
	} else {
		fromDate = f.opts.From
		untilDate = f.opts.Until
	}
	pdf.Cell(0, lineHeight, fmt.Sprintf("%s - %s", fromDate.Format(dateFormat), untilDate.Format(dateFormat)))
	pdf.Ln(lineHeight)

	// Empty line
	pdf.Ln(lineHeight)

	// Total Charging Records
	pdf.SetFontStyle("B")
	pdf.Cell(47, lineHeight, "Total Charging Records:")
	pdf.SetFontStyle("")
	pdf.Cell(0, lineHeight, strconv.Itoa(len(f.sessions)))
	pdf.Ln(lineHeight)

	// Total Consumption
	pdf.SetFontStyle("B")
	pdf.Cell(47, lineHeight, "Total Consumption:")
	pdf.SetFontStyle("")
	pdf.Cell(0, lineHeight, fmt.Sprintf("%.2f kWh", f.calculateTotalConsumption()))
	pdf.Ln(lineHeight * 2) // Extra space before table
}

// calculateTotalConsumption sums up all consumption values
func (f *PDFFormatter) calculateTotalConsumption() float64 {
	var total float64
	for _, session := range f.sessions {
		total += session.Consumption
	}
	return total
}

// writeTableHeader writes the table header row
func (f *PDFFormatter) writeTableHeader(pdf *fpdf.Fpdf) {
	pdf.SetFontStyle("B")

	//FIXME: do as for the bodies
	x := pdf.GetX()
	y := pdf.GetY()
	for i, header := range colHeaders {
		// Draw border only, no fill
		pdf.Rect(x, y, colWidths[i], headerHeight, "D")
		pdf.SetXY(x, y)
		// Center text vertically by adjusting cell height
		lines := 1
		if i == 1 || i == 4 { // Two-line colHeaders
			lines = 2
		}
		cellLineHeight := headerHeight / float64(lines+1)
		if lines == 1 {
			pdf.SetXY(x, y+(headerHeight-cellLineHeight)/2)
			pdf.CellFormat(colWidths[i], cellLineHeight, header, "", 0, "C", false, 0, "")
		} else {
			pdf.SetXY(x, y+(headerHeight-cellLineHeight*2)/2)
			pdf.MultiCell(colWidths[i], cellLineHeight, header, "", "C", false)
		}
		x += colWidths[i]
	}
	pdf.SetY(y + headerHeight)

	pdf.SetFontStyle("")
}

// writeTable writes the data table to the PDF
func (f *PDFFormatter) writeTable(pdf *fpdf.Fpdf) {
	pdf.SetFont("Arial", "", bodyFontSize)
	pdf.SetCellMargin(2)

	f.writeTableHeader(pdf)

	for _, session := range f.sessions {
		// Check if we need a new page
		if pdf.GetY()+rowHeight > 280 { // Leave space for footer
			pdf.AddPage()
			f.writeTableHeader(pdf)
		}

		start := ""
		if !session.Start.IsZero() {
			start = session.Start.Format(dateTimeFormat)
		}

		x := pdf.GetX()
		y := pdf.GetY()

		cell := func(col int, text, align string) {
			w := colWidths[col]

			lineCount := strings.Count(text, "\n") + 1
			if wrappedLines := int(math.Ceil(pdf.GetStringWidth(text) / w)); wrappedLines > lineCount {
				lineCount = wrappedLines
			}

			offset := 0.0
			for i := 0; i < col && i < len(colWidths); i++ {
				offset += colWidths[i]
			}

			pdf.SetXY(x+offset, y)
			pdf.MultiCell(w, 12.0/float64(lineCount), text, "1", align, false)
		}

		cell(0, session.End.Format(dateFormat), "C")
		cell(1, strconv.FormatFloat(session.Consumption, 'f', 2, 64), "R")
		cell(2, session.ChargerName, "L")
		cell(3, session.Authentication, "L")
		cell(4, fmt.Sprintf("%s\n%s", start, session.End.Format(dateTimeFormat)), "L")
	}
}
