package output

import (
	"io"
	"time"

	"github.com/joshiste/sma_chg_log/internal/models"
)

// MessageFormatter defines the interface for outputting individual messages
type MessageFormatter interface {
	WriteMessage(msg models.Message) error
	Flush() error
}

// SessionFormatter defines the interface for outputting charging sessions
type SessionFormatter interface {
	WriteHeader() error
	WriteSession(session models.ChargingSession) error
	Flush() error
}

// Options contains options for PDF formatting
type Options struct {
	From  time.Time
	Until time.Time
}

// NewMessageFormatter creates a message formatter (JSON only)
func NewMessageFormatter(w io.Writer) MessageFormatter {
	return NewJSONMessageFormatter(w)
}

// NewSessionFormatter creates a session formatter based on the format type
func NewSessionFormatter(format string, w io.Writer) SessionFormatter {
	return NewSessionFormatterWithOptions(format, w, Options{})
}

// NewSessionFormatterWithOptions creates a session formatter with PDF options
func NewSessionFormatterWithOptions(format string, w io.Writer, opts Options) SessionFormatter {
	switch format {
	case "csv":
		return NewCSVFormatter(w)
	case "pdf":
		return NewPDFFormatterWithOptions(w, opts)
	default:
		return NewJSONSessionFormatter(w)
	}
}
