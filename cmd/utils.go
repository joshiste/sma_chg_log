package cmd

import (
	"github.com/joshiste/sma_chg_log/internal/models"
)

const (
	messageIDChargingCompleted = 9813
	messageIDChargingStarted   = 9812
)

// filterMessages filters messages by messageId (charging started/completed only)
func filterMessages(messages []models.Message) []models.Message {
	var filtered []models.Message
	for _, msg := range messages {
		if msg.MessageID == messageIDChargingCompleted || msg.MessageID == messageIDChargingStarted {
			filtered = append(filtered, msg)
		}
	}
	return filtered
}
