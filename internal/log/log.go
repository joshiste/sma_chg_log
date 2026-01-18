package log

import (
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/viper"
)

const (
	// LevelTrace is a custom log level below Debug for detailed tracing
	LevelTrace = slog.Level(-8)
)

func Init() {
	level, err := parseLogLevel(viper.GetString("log-level"))
	if err != nil {
		slog.Warn("invalid log level", "error", err)
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey {
				if a.Value.Any().(slog.Level) == LevelTrace {
					a.Value = slog.StringValue("TRACE")
				}
			}
			return a
		},
	})))
}

func parseLogLevel(level string) (slog.Level, error) {
	if strings.EqualFold(level, "trace") {
		return LevelTrace, nil
	}

	var l slog.Level
	if err := l.UnmarshalText([]byte(level)); err != nil {
		return slog.LevelInfo, err
	}
	return l, nil
}
