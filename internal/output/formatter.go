package output

import (
	"io"

	"sma_event_log/internal/models"
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

// NewMessageFormatter creates a message formatter (JSON only)
func NewMessageFormatter(w io.Writer) MessageFormatter {
	return NewJSONMessageFormatter(w)
}

// NewSessionFormatter creates a session formatter based on the format type
func NewSessionFormatter(format string, w io.Writer) SessionFormatter {
	switch format {
	case "csv":
		return NewCSVFormatter(w)
	case "pdf":
		return NewPDFFormatter(w)
	default:
		return NewJSONSessionFormatter(w)
	}
}
