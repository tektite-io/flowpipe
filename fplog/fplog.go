package fplog

import (
	"context"
	"os"
	"strings"

	//nolint:depguard // Wrapper for Zap
	"go.uber.org/zap"

	//nolint:depguard // Wrapper for Zap
	"go.uber.org/zap/zapcore"

	"github.com/turbot/flowpipe/sanitize"
)

type FlowpipeLogger struct {

	// Level is the logging level to use for output
	Level zapcore.Level

	// Format is the logging format to use for output: json or console
	Format string

	// Color is whether to use color in the console output
	Color bool

	// Zap is the Zap logger instance
	Zap *zap.SugaredLogger
}

// LoggerOption defines a type of function to configures the Logger.
type LoggerOption func(*FlowpipeLogger) error

// NewLogger creates a new Logger.
func NewLogger(ctx context.Context, opts ...LoggerOption) (*FlowpipeLogger, error) {
	// Defaults
	c := &FlowpipeLogger{
		Level:  zapcore.InfoLevel,
		Format: "console",
	}
	// Set options
	for _, opt := range opts {
		err := opt(c)
		if err != nil {
			return c, err
		}
	}

	err := c.Initialize()
	if err != nil {
		return nil, err
	}
	return c, nil
}

func WithColor(enabled bool) LoggerOption {
	return func(c *FlowpipeLogger) error {
		c.Color = enabled
		return nil
	}
}

func WithLevelFromEnvironment() LoggerOption {
	return func(c *FlowpipeLogger) error {
		// Get the desired logging level from the FLOWPIPE_LOG_LEVEL environment variable
		logLevelStr := strings.ToLower(os.Getenv("FLOWPIPE_LOG_LEVEL"))
		// If the FLOWPIPE_LOG_LEVEL environment variable is set, parse its value to determine the logging level
		if logLevelStr != "" {
			var err error
			logLevel, err := zapcore.ParseLevel(logLevelStr)
			if err == nil {
				c.Level = logLevel
			}
		}
		return nil
	}
}

func WithFormatFromEnvironment() LoggerOption {
	return func(c *FlowpipeLogger) error {
		// Get the desired logging format from the FLOWPIPE_LOG_FORMAT environment variable
		logFormat := strings.ToLower(os.Getenv("FLOWPIPE_LOG_FORMAT"))
		switch logFormat {
		case "json", "console":
			c.Format = logFormat
		}
		return nil
	}
}

func (c *FlowpipeLogger) Initialize() error {

	// Configure the logging output
	var encoder zapcore.Encoder
	if c.Format == "json" {
		encoder = zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	} else {
		ec := zap.NewDevelopmentEncoderConfig()
		if c.Color {
			ec.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}
		encoder = zapcore.NewConsoleEncoder(ec)
	}
	consoleDebugging := zapcore.Lock(os.Stdout)
	//consoleErrors := zapcore.Lock(os.Stderr)

	// Configure the logging level based on the FLOWPIPE_LOG_LEVEL environment variable
	atomicLevel := zap.NewAtomicLevelAt(c.Level)

	// Create the Zap logger instance
	core := zapcore.NewTee(
		//zapcore.NewCore(encoder, consoleErrors, atomicLevel),
		zapcore.NewCore(encoder, consoleDebugging, atomicLevel),
	)

	c.Zap = zap.New(core).Sugar()

	_, err := zap.RedirectStdLogAt(c.Zap.Desugar(), zapcore.DebugLevel)
	if err != nil {
		return err
	}

	return nil
}

func (c *FlowpipeLogger) Sync() error {
	return c.Zap.Sync()
}

func (c *FlowpipeLogger) Error(msg string, keysAndValues ...interface{}) {
	sanitizedKeysAndValues := sanitize.SanitizeLogEntries(keysAndValues)
	c.Zap.Errorw(msg, sanitizedKeysAndValues...)
}

func (c *FlowpipeLogger) Warn(msg string, keysAndValues ...interface{}) {
	sanitizedKeysAndValues := sanitize.SanitizeLogEntries(keysAndValues)
	c.Zap.Warnw(msg, sanitizedKeysAndValues...)
}

func (c *FlowpipeLogger) Info(msg string, keysAndValues ...interface{}) {
	sanitizedKeysAndValues := sanitize.SanitizeLogEntries(keysAndValues)
	c.Zap.Infow(msg, sanitizedKeysAndValues...)
}

func (c *FlowpipeLogger) Debug(msg string, keysAndValues ...interface{}) {
	sanitizedKeysAndValues := sanitize.SanitizeLogEntries(keysAndValues)
	c.Zap.Debugw(msg, sanitizedKeysAndValues...)
}