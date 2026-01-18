package models

import "time"

// ChargingSession represents a paired charging start/stop event
type ChargingSession struct {
	ChargerName    string    `json:"chargerName"`
	Consumption    float64   `json:"consumption"`
	Authentication string    `json:"authentication,omitzero"`
	Start          time.Time `json:"start,omitzero"`
	End            time.Time `json:"end"`
}
