package models

import "time"

var (
	TimeMax = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
)

// ChargingSession represents a paired charging start/stop event
type ChargingSession struct {
	ChargerName    string    `json:"chargerName"`
	Consumption    float64   `json:"consumption"`
	Authentication string    `json:"authentication,omitzero"`
	Start          time.Time `json:"start,omitzero"`
	End            time.Time `json:"end"`
}
