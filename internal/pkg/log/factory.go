package log

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/MatusOllah/slogcolor"
)

func CreateLoggerFromConfig(config Config) (*slog.Logger, error) {
	level, err := parseLogLevel(config.Level)
	if err != nil {
		return nil, fmt.Errorf("parse log level: %w", err)
	}

	output := io.Writer(os.Stderr)
	if config.Quiet {
		output = io.Discard
	}

	handler, err := createLogHandler(config.Format, output, level)
	if err != nil {
		return nil, fmt.Errorf("create log handler: %w", err)
	}

	return slog.New(handler), nil
}

func parseLogLevel(value string) (slog.Level, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return slog.LevelInfo, ErrLogLevelRequired
	}
	switch normalized {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("%w: %s", ErrUnknownLogLevel, value)
	}
}

func createLogHandler(format string, output io.Writer, level slog.Level) (slog.Handler, error) {
	normalizedFormat := strings.ToLower(strings.TrimSpace(format))
	if normalizedFormat == "" {
		return nil, ErrLogFormatRequired
	}

	switch normalizedFormat {
	case "text-no-color":
		return slog.NewTextHandler(output, &slog.HandlerOptions{
			Level: level,
		}), nil
	case "text-color":
		options := slogcolor.DefaultOptions
		options.Level = level
		return slogcolor.NewHandler(output, options), nil
	case "json":
		return slog.NewJSONHandler(output, &slog.HandlerOptions{
			Level: level,
		}), nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownLogFormat, format)
	}
}

var (
	// ErrLogLevelRequired reports a missing log level.
	ErrLogLevelRequired = errors.New("log level is required")

	// ErrUnknownLogLevel reports an unrecognized log level.
	ErrUnknownLogLevel = errors.New("unknown log level")

	// ErrLogFormatRequired reports a missing log format.
	ErrLogFormatRequired = errors.New("log format is required")

	// ErrUnknownLogFormat reports an unrecognized log format.
	ErrUnknownLogFormat = errors.New("unknown log format")
)
