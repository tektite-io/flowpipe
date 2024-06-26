package log

import (
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/sanitize"

	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/constants"
)

func FlowpipeLogger() *slog.Logger {
	handlerOptions := &slog.HandlerOptions{
		Level: GetLogLevel(),
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			sanitized := sanitize.Instance.SanitizeKeyValue(a.Key, a.Value.Any())

			return slog.Attr{
				Key:   a.Key,
				Value: slog.AnyValue(sanitized),
			}
		},
	}

	if handlerOptions.Level == constants.LogLevelOff {
		return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	}
	return slog.New(slog.NewJSONHandler(os.Stderr, handlerOptions))
}

func FlowpipeLoggerWithLevelAndWriter(level slog.Leveler, w io.Writer) *slog.Logger {
	handlerOptions := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			sanitized := sanitize.Instance.SanitizeKeyValue(a.Key, a.Value.Any())

			return slog.Attr{
				Key:   a.Key,
				Value: slog.AnyValue(sanitized),
			}
		},
	}

	if handlerOptions.Level == constants.LogLevelOff {
		return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	}

	return slog.New(slog.NewJSONHandler(os.Stderr, handlerOptions))
}

func SetDefaultLogger() {
	logger := FlowpipeLogger()
	slog.SetDefault(logger)
}

func GetLogLevel() slog.Leveler {
	logLevel := os.Getenv(app_specific.EnvLogLevel)

	if logLevel == "" {
		logLevel = viper.GetString(constants.ArgLogLevel)
	}

	switch strings.ToLower(logLevel) {
	case "trace":
		return constants.LogLevelTrace
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	case "off":
		return constants.LogLevelOff
	default:
		return constants.LogLevelOff
	}
}
