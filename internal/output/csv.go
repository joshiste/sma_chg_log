package output

import (
	"encoding/csv"
	"io"
	"strconv"
	"time"

	"github.com/joshiste/sma_chg_log/internal/models"
)

// CSVFormatter outputs charging session as CSV
type CSVFormatter struct {
	writer *csv.Writer
}

// NewCSVFormatter creates a new CSV formatter
func NewCSVFormatter(w io.Writer) *CSVFormatter {
	return &CSVFormatter{
		writer: csv.NewWriter(w),
	}
}

// WriteHeader writes the CSV header row
func (f *CSVFormatter) WriteHeader() error {
	return f.writer.Write([]string{
		"record date",
		"charger name",
		"authentication",
		"start",
		"end",
		"consumption",
	})
}

// WriteSession writes a charging session as a CSV row
func (f *CSVFormatter) WriteSession(session models.ChargingSession) error {
	start := ""
	if !session.Start.IsZero() {
		start = session.Start.Format(time.RFC3339)
	}

	return f.writer.Write([]string{
		session.End.Format("2006-01-02"),
		session.ChargerName,
		session.Authentication,
		start,
		session.End.Format(time.RFC3339),
		strconv.FormatFloat(session.Consumption, 'f', 2, 64),
	})
}

// Flush ensures all buffered data is written
func (f *CSVFormatter) Flush() error {
	f.writer.Flush()
	return f.writer.Error()
}
