package output

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/go-pdf/fpdf"

	"sma_event_log/internal/models"
)

// PDFFormatter outputs charging sessions
type PDFFormatter struct {
	writer   io.Writer
	sessions []models.ChargingSession
}

// NewPDFFormatter creates a new PDF formatter
func NewPDFFormatter(w io.Writer) *PDFFormatter {
	return &PDFFormatter{
		writer:   w,
		sessions: make([]models.ChargingSession, 0),
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
	pdf := fpdf.New("L", "mm", "A4", "") // Landscape for more columns
	pdf.SetFont("Arial", "", 10)
	pdf.AddPage()

	// Calculate summary statistics
	totalConsumption := f.calculateTotalConsumption()
	chargingCompletedCount := len(f.sessions)

	// Write title
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "Charging Events Report")
	pdf.Ln(15)

	// Write summary
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(0, 8, fmt.Sprintf("Total Charging Records: %d", chargingCompletedCount))
	pdf.Ln(8)
	pdf.Cell(0, 8, fmt.Sprintf("Total Consumption: %.2f kWh", totalConsumption))
	pdf.Ln(15)

	// Write table
	f.writeTable(pdf)

	return pdf.Output(f.writer)
}

// calculateTotalConsumption sums up all consumption values
func (f *PDFFormatter) calculateTotalConsumption() float64 {
	var total float64
	for _, session := range f.sessions {
		total += session.Consumption
	}
	return total
}

// writeTable writes the data table to the PDF
func (f *PDFFormatter) writeTable(pdf *fpdf.Fpdf) {
	// Table header
	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(200, 200, 200)

	colWidths := []float64{30, 45, 55, 50, 50, 35}
	headers := []string{"Record Date", "Charger Name", "Authentication", "Start", "End", "Consumption"}

	for i, header := range headers {
		pdf.CellFormat(colWidths[i], 8, header, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	// Table rows
	pdf.SetFont("Arial", "", 8)
	pdf.SetFillColor(255, 255, 255)

	for _, session := range f.sessions {
		start := ""
		if !session.Start.IsZero() {
			start = session.Start.Format(time.RFC3339)
		}

		authentication := session.Authentication
		if authentication == "" {
			authentication = "no authentication"
		}

		pdf.CellFormat(colWidths[0], 7, session.End.Format("2006-01-02"), "1", 0, "L", false, 0, "")
		pdf.CellFormat(colWidths[1], 7, session.ChargerName, "1", 0, "L", false, 0, "")
		pdf.CellFormat(colWidths[2], 7, authentication, "1", 0, "L", false, 0, "")
		pdf.CellFormat(colWidths[3], 7, start, "1", 0, "L", false, 0, "")
		pdf.CellFormat(colWidths[4], 7, session.End.Format(time.RFC3339), "1", 0, "L", false, 0, "")
		pdf.CellFormat(colWidths[5], 7, strconv.FormatFloat(session.Consumption, 'f', 2, 64), "1", 0, "R", false, 0, "")
		pdf.Ln(-1)
	}
}
