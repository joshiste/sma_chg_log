package output

import (
	"encoding/json"
	"io"

	"github.com/joshiste/sma_chg_log/internal/models"
)

// JSONMessageFormatter outputs messages as JSON
type JSONMessageFormatter struct {
	encoder *json.Encoder
}

// NewJSONMessageFormatter creates a new JSON message formatter
func NewJSONMessageFormatter(w io.Writer) *JSONMessageFormatter {
	return &JSONMessageFormatter{
		encoder: json.NewEncoder(w),
	}
}

// WriteMessage writes the message as JSON (uses original raw JSON via MarshalJSON)
func (f *JSONMessageFormatter) WriteMessage(msg models.Message) error {
	return f.encoder.Encode(msg)
}

// Flush is a no-op for JSON format
func (f *JSONMessageFormatter) Flush() error {
	return nil
}

// JSONSessionFormatter outputs charging sessions as JSON
type JSONSessionFormatter struct {
	encoder *json.Encoder
}

// NewJSONSessionFormatter creates a new JSON session formatter
func NewJSONSessionFormatter(w io.Writer) *JSONSessionFormatter {
	return &JSONSessionFormatter{
		encoder: json.NewEncoder(w),
	}
}

// WriteHeader is a no-op for JSON format
func (f *JSONSessionFormatter) WriteHeader() error {
	return nil
}

// WriteSession writes a charging session as JSON
func (f *JSONSessionFormatter) WriteSession(session models.ChargingSession) error {
	return f.encoder.Encode(session)
}

// Flush is a no-op for JSON format
func (f *JSONSessionFormatter) Flush() error {
	return nil
}
