package cmd

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"sma_event_log/internal/client"
	"sma_event_log/internal/models"
	"sma_event_log/internal/output"
)

var sessionsCmd = &cobra.Command{
	Use:   "sessions",
	Short: "Output charging sessions",
	Long:  "Fetch charging events and output paired charging sessions in JSON, CSV, or PDF format",
	RunE:  runSessions,
}

func init() {
	must(viper.BindPFlags(sessionsCmd.Flags()))

	rootCmd.AddCommand(sessionsCmd)

	// Set sessions as the default command
	rootCmd.RunE = runSessions
}

func runSessions(cmd *cobra.Command, args []string) error {
	if cfg.Format != "" && cfg.Format != "json" && cfg.Format != "csv" && cfg.Format != "pdf" {
		return errors.New("format must be 'json', 'csv', or 'pdf'")
	}

	apiClient := client.New(cfg.URL, cfg.Username, cfg.Password)
	formatter := output.NewSessionFormatter(cfg.Format, os.Stdout)

	if err := formatter.WriteHeader(); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Collect all messages for pairing
	var allMessages []models.Message
	err := apiClient.FetchAllMessages(cfg.From, cfg.Until, func(messages []models.Message) bool {
		allMessages = append(allMessages, filterMessages(messages)...)
		return true
	})
	if err != nil {
		return err
	}

	// Pair messages into sessions and output
	sessions := pairChargingSessions(allMessages)
	for _, session := range sessions {
		if err := formatter.WriteSession(session); err != nil {
			return fmt.Errorf("failed to write session: %w", err)
		}
	}

	return formatter.Flush()
}

// pairChargingSessions pairs charging stopped events with their preceding started events
// Messages are ordered newest to oldest, so a stopped event at index i may pair with started at i+1
func pairChargingSessions(messages []models.Message) []models.ChargingSession {
	var sessions []models.ChargingSession

	for i := 0; i < len(messages); i++ {
		msg := messages[i]

		// Only process charging completed events
		if msg.MessageID != messageIDChargingCompleted {
			continue
		}

		session := models.ChargingSession{
			ChargerName:    msg.DeviceName,
			Consumption:    findConsumption(msg.Arguments),
			Authentication: "",
			End:            msg.Timestamp,
		}

		// Look for preceding charging started event (next in array since newest first)
		if i+1 < len(messages) && messages[i+1].MessageID == messageIDChargingStarted {
			startMsg := messages[i+1]
			session.Authentication = findAuthentication(startMsg.Arguments)
			session.Start = startMsg.Timestamp
			i++ // Skip the start message since it's now paired
		}

		sessions = append(sessions, session)
	}

	return sessions
}

// findConsumption finds the consumption value from message arguments
func findConsumption(args []models.MessageArgument) float64 {
	for _, arg := range args {
		if arg.UnitTag == 8 && arg.DisplayType == "Fix2" {
			if val, err := strconv.ParseFloat(arg.Value, 64); err == nil {
				return val
			}
		}
	}
	return 0
}

// findAuthentication finds the authentication value from message arguments
func findAuthentication(args []models.MessageArgument) string {
	for _, arg := range args {
		if arg.DisplayType == "String" && arg.Position == 0 {
			return arg.Value
		}
	}
	return ""
}
