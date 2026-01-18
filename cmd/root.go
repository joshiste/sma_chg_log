package cmd

import (
	"errors"
	"io"
	"os"
	"strings"
	"time"

	"github.com/joshiste/sma_chg_log/internal/models"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/joshiste/sma_chg_log/internal/log"
)

var cfg = Config{}

type Config struct {
	Host     string
	Username string
	Password string
	Format   string
	Writer   io.Writer
	From     time.Time
	Until    time.Time
}

func (c *Config) Validate() error {
	var errs []error

	if c.Host == "" {
		errs = append(errs, errors.New("host is required (use --host flag or SMA_HOST environment variable)"))
	}

	if !strings.HasPrefix(c.Host, "http://") && !strings.HasPrefix(c.Host, "https://") {
		c.Host = "https://" + c.Host
	}

	if c.Username == "" {
		errs = append(errs, errors.New("username is required (use --username flag or SMA_USERNAME environment variable)"))
	}

	if c.Password == "" {
		errs = append(errs, errors.New("password is required (use --password flag or SMA_PASSWORD environment variable)"))
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
		c.Until = models.TimeMax
	}

	if output := viper.GetString("output"); output == "-" {
		c.Writer = os.Stdout
	} else if f, err := os.Create(output); err == nil {
		c.Writer = f
	} else {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

var rootCmd = &cobra.Command{
	Use:                "sma_chg_log",
	Short:              "Fetch event messages from ennexos device",
	Long:               "A tool to fetch customer messages from SMA device API and output specific events",
	PersistentPreRunE:  persistentPreRunE,
	PersistentPostRunE: persistentPostRunE,
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global persistent flags (available to all subcommands)
	rootCmd.PersistentFlags().StringP("host", "H", "", "Hostname of the SMA device")
	rootCmd.PersistentFlags().StringP("username", "u", "", "Username for authentication")
	rootCmd.PersistentFlags().StringP("password", "p", "", "Password for authentication")
	rootCmd.PersistentFlags().StringP("log-level", "l", "info", "Log level: trace, debug, info, warn, error")
	rootCmd.PersistentFlags().StringP("month", "m", "", "Filter by month (format: YYYY-MM)")
	rootCmd.PersistentFlags().StringP("format", "f", "json", "Output format: json, csv, or pdf")
	rootCmd.PersistentFlags().StringP("output", "o", "-", "Output file path (use '-' for stdout)")

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

func persistentPostRunE(cmd *cobra.Command, args []string) error {
	if f, ok := cfg.Writer.(io.Closer); ok && f != os.Stdout && f != os.Stderr {
		_ = f.Close()
	}
	return nil
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
