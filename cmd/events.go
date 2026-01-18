package cmd

import (
	"errors"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"sma_event_log/internal/client"
	"sma_event_log/internal/models"
	"sma_event_log/internal/output"
)

var eventsCmd = &cobra.Command{
	Use:   "events",
	Short: "Output raw charging start/stop events as JSON",
	Long:  "Fetch and output raw charging event messages in JSON format",
	RunE:  runEvents,
}

func init() {
	must(viper.BindPFlags(eventsCmd.Flags()))

	rootCmd.AddCommand(eventsCmd)
}

func runEvents(cmd *cobra.Command, args []string) error {
	if cfg.Format != "" && cfg.Format != "json" {
		return errors.New("only 'json' fromat supported for events command")
	}

	apiClient := client.New(cfg.URL, cfg.Username, cfg.Password)
	formatter := output.NewMessageFormatter(os.Stdout)

	err := apiClient.FetchAllMessages(cfg.From, cfg.Until, func(messages []models.Message) bool {
		for _, msg := range filterMessages(messages) {
			if err := formatter.WriteMessage(msg); err != nil {
				return false
			}
		}
		return true
	})
	if err != nil {
		return err
	}

	return formatter.Flush()
}
