package cmd

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/joshiste/sma_chg_log/internal/client"
	"github.com/joshiste/sma_chg_log/internal/models"
	"github.com/joshiste/sma_chg_log/internal/output"
)

var eventsCmd = &cobra.Command{
	Use:   "events",
	Short: "Writer raw charging start/stop events as JSON",
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

	apiClient := client.New(cfg.Host, cfg.Username, cfg.Password)
	formatter := output.NewMessageFormatter(cfg.Writer)

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
