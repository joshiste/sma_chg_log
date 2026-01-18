package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/joshiste/sma_chg_log/internal/client"
	"github.com/joshiste/sma_chg_log/internal/models"
	"github.com/joshiste/sma_chg_log/internal/output"
)

var mapAuthenticationRaw []string

var sessionsCmd = &cobra.Command{
	Use:   "sessions",
	Short: "Writer charging sessions",
	Long:  "Fetch charging events and output paired charging sessions in JSON, CSV, or PDF format",
	RunE:  runSessions,
}

func init() {
	sessionsCmd.Flags().StringArrayVarP(&mapAuthenticationRaw, "map-authentication", "a", nil, "Map authentication values (format: old:new, can be specified multiple times)")
	must(viper.BindPFlags(sessionsCmd.Flags()))

	rootCmd.AddCommand(sessionsCmd)

	rootCmd.RunE = runSessions
}

func parseMapAuthentication(raw []string) map[string]string {
	result := make(map[string]string)
	for _, entry := range raw {
		parts := strings.SplitN(entry, ":", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
}

func runSessions(cmd *cobra.Command, args []string) error {
	if cfg.Format != "" && cfg.Format != "json" && cfg.Format != "csv" && cfg.Format != "pdf" {
		return errors.New("format must be 'json', 'csv', or 'pdf'")
	}

	authMap := parseMapAuthentication(mapAuthenticationRaw)
	slog.Debug("Authentication mapping", "map", authMap)

	apiClient := client.New(cfg.Host, cfg.Username, cfg.Password)

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
	sessions := pairChargingSessions(allMessages, authMap)

	// Calculate date range from sessions if not explicitly set
	opts := output.Options{
		From:  cfg.From,
		Until: cfg.Until,
	}
	if len(sessions) > 0 && opts.From == models.TimeMin {
		opts.From = toDate(sessions[len(sessions)-1].End)
	}
	if len(sessions) > 0 && opts.Until == models.TimeMax {
		opts.Until = toDate(sessions[0].End).AddDate(0, 0, 1)
	}
	if time.Now().Before(opts.Until) {
		opts.Until = toDate(time.Now()).AddDate(0, 0, 1)
	}

	formatter := output.NewSessionFormatterWithOptions(cfg.Format, cfg.Writer, opts)

	if err := formatter.WriteHeader(); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	for _, session := range sessions {
		if err := formatter.WriteSession(session); err != nil {
			return fmt.Errorf("failed to write session: %w", err)
		}
	}

	return formatter.Flush()
}

func toDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// pairChargingSessions pairs charging stopped events with their preceding started events
// Messages are ordered newest to oldest, so a stopped event at index i may pair with started at i+1
func pairChargingSessions(messages []models.Message, authMap map[string]string) []models.ChargingSession {
	var sessions []models.ChargingSession

	for i := 0; i < len(messages); i++ {
		msg := messages[i]

		// Only process charging completed events
		if msg.MessageID != messageIDChargingCompleted {
			continue
		}

		session := models.ChargingSession{
			ChargerName: msg.DeviceName,
			Consumption: findConsumption(msg.Arguments),
			End:         msg.Timestamp,
		}

		// Look for preceding charging started event (next in array since newest first)
		if i+1 < len(messages) && messages[i+1].MessageID == messageIDChargingStarted {
			startMsg := messages[i+1]
			auth := findAuthentication(startMsg.Arguments)
			session.Authentication = auth
			session.Start = startMsg.Timestamp
			i++ // Skip the start message since it's now paired
		}

		// Apply authorization mapping if configured
		if mapped, ok := authMap[session.Authentication]; ok {
			session.Authentication = mapped
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
