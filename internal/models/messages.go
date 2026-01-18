package models

import (
	"encoding/json"
	"time"
)

// SearchRequest represents the POST body for the messages search endpoint
type SearchRequest struct {
	ComponentID      string   `json:"componentId"`
	From             *string  `json:"from"`
	Until            *string  `json:"until"`
	MessageGroupTags []int    `json:"messageGroupTags"`
	TraceLevels      []string `json:"traceLevels"`
	Marker           string   `json:"marker"`
	Offset           int      `json:"offset"`
}

// MessageArgument represents an argument within a message
type MessageArgument struct {
	DisplayType string          `json:"displayType"`
	Position    int             `json:"position"`
	UnitTag     int             `json:"unitTag"`
	Value       string          `json:"value"`
	RawJSON     json.RawMessage `json:"-"`
}

// UnmarshalJSON implements custom unmarshaling to capture raw JSON
func (a *MessageArgument) UnmarshalJSON(data []byte) error {
	type Alias MessageArgument
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(a),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	a.RawJSON = data
	return nil
}

// MarshalJSON returns the original raw JSON
func (a *MessageArgument) MarshalJSON() ([]byte, error) {
	return a.RawJSON, nil
}

// Message represents a single message from the API response
type Message struct {
	Arguments          []MessageArgument `json:"arguments"`
	DeviceID           string            `json:"deviceId"`
	DeviceName         string            `json:"deviceName"`
	DeviceSerialnumber string            `json:"deviceSerialnumber"`
	EscalationLevel    int               `json:"escalationLevel"`
	EventTypeExtension string            `json:"eventTypeExtension"`
	Marker             string            `json:"marker"`
	MessageGroupTag    int               `json:"messageGroupTag"`
	MessageID          int               `json:"messageId"`
	MessageTag         int               `json:"messageTag"`
	Timestamp          time.Time         `json:"timestamp"`
	TraceLevel         string            `json:"traceLevel"`
	RawJSON            json.RawMessage   `json:"-"`
}

// UnmarshalJSON implements custom unmarshalling to capture raw JSON
func (m *Message) UnmarshalJSON(data []byte) error {
	type Alias Message
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(m),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	m.RawJSON = data
	return nil
}

// MarshalJSON returns the original raw JSON
func (m *Message) MarshalJSON() ([]byte, error) {
	return m.RawJSON, nil
}

// TokenResponse represents the response from the token endpoint
type TokenResponse struct {
	AccessToken string `json:"access_token"`
}
