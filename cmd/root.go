package cmd

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"sma_event_log/internal/log"
)

var cfg = Config{}

type Config struct {
	URL      string
	Username string
	Password string
	Format   string
	From     time.Time
	Until    time.Time
}

func (c *Config) Validate() error {
	var errs []error

	if c.URL == "" {
		errs = append(errs, errors.New("url is required (use --url flag or SMA_URL environment variable)"))
	}

	if !strings.HasPrefix(c.URL, "http://") && !strings.HasPrefix(c.URL, "https://") {
		c.URL = "https://" + c.URL
	}

	if c.Username == "" {
		errs = append(errs, errors.New("username is required (use --username flag or SMA_USERNAME environment variable)"))
	}

	if c.Password == "" {
		errs = append(errs, errors.New("password is required (use SMA_PASSWORD environment variable)"))
	}

	if month := viper.GetString("month"); month != "" {
		parsed, err := time.Parse("2006-01", month)
		if err != nil {
			errs = append(errs, errors.New("month must be in format YYYY-MM"))
		} else {
			c.From = parsed.UTC()
			c.Until = parsed.AddDate(0, 1, 0).UTC()
		}
	} else {
		c.From = time.Unix(0, 0).UTC()
		c.Until = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
	}

	return errors.Join(errs...)
}

var rootCmd = &cobra.Command{
	Use:               "sma_event_log",
	Short:             "Fetch event messages from ennexos device",
	Long:              "A tool to fetch customer messages from SMA device API and output specific events",
	PersistentPreRunE: persistentPreRunE,
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global persistent flags (available to all subcommands)
	rootCmd.PersistentFlags().String("url", "", "Base URL of the SMA device API")
	rootCmd.PersistentFlags().String("username", "", "Username for authentication")
	rootCmd.PersistentFlags().String("log-level", "info", "Log level: trace, debug, info, warn, error")
	rootCmd.PersistentFlags().String("month", "", "Filter by month (format: YYYY-MM)")
	rootCmd.PersistentFlags().String("format", "json", "Output format: json, csv, or pdf")

	must(viper.BindPFlags(rootCmd.PersistentFlags()))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func initConfig() {
	viper.SetEnvPrefix("SMA")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	must(viper.BindEnv("password"))
	viper.AutomaticEnv()

	log.Init()
}

func persistentPreRunE(cmd *cobra.Command, args []string) error {
	if err := viper.Unmarshal(&cfg); err != nil {
		return err
	}

	if err := cfg.Validate(); err != nil {
		return err
	}

	return nil
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
