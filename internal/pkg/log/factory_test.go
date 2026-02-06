package log

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseLogLevel_WhenEmpty_ReturnsRequiredError(t *testing.T) {
	input := "   "

	level, err := parseLogLevel(input)

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrLogLevelRequired))
	assert.Equal(t, slog.LevelInfo, level)
}

func TestParseLogLevel_WhenKnownValue_ReturnsLevel(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected slog.Level
	}{
		{
			name:     "Debug",
			input:    "  DeBuG ",
			expected: slog.LevelDebug,
		},
		{
			name:     "Info",
			input:    "info",
			expected: slog.LevelInfo,
		},
		{
			name:     "Warn",
			input:    "WARN",
			expected: slog.LevelWarn,
		},
		{
			name:     "Error",
			input:    " error ",
			expected: slog.LevelError,
		},
	}

	for _, testCase := range cases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			level, err := parseLogLevel(testCase.input)

			require.NoError(t, err)
			assert.Equal(t, testCase.expected, level)
		})
	}
}

func TestParseLogLevel_WhenUnknownValue_ReturnsUnknownError(t *testing.T) {
	input := "verbose"

	level, err := parseLogLevel(input)

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrUnknownLogLevel))
	assert.Equal(t, "unknown log level: verbose", err.Error())
	assert.Equal(t, slog.LevelInfo, level)
}

func TestCreateLogHandler_WhenEmptyFormat_ReturnsRequiredError(t *testing.T) {
	level := slog.LevelInfo
	output := &bytes.Buffer{}

	handler, err := createLogHandler(" ", output, level)

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrLogFormatRequired))
	assert.Nil(t, handler)
}

func TestCreateLogHandler_WhenUnknownFormat_ReturnsUnknownError(t *testing.T) {
	level := slog.LevelInfo
	output := &bytes.Buffer{}

	handler, err := createLogHandler("custom", output, level)

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrUnknownLogFormat))
	assert.Equal(t, "unknown log format: custom", err.Error())
	assert.Nil(t, handler)
}

func TestCreateLogHandler_WhenTextNoColor_RespectsLevel(t *testing.T) {
	level := slog.LevelWarn
	output := &bytes.Buffer{}

	handler, err := createLogHandler("text-no-color", output, level)

	require.NoError(t, err)
	assert.NotNil(t, handler)
	assert.False(t, handler.Enabled(context.Background(), slog.LevelInfo))
	assert.True(t, handler.Enabled(context.Background(), slog.LevelWarn))
}

func TestCreateLogHandler_WhenTextColor_RespectsLevel(t *testing.T) {
	level := slog.LevelWarn
	output := &bytes.Buffer{}

	handler, err := createLogHandler("text-color", output, level)

	require.NoError(t, err)
	assert.NotNil(t, handler)
	assert.False(t, handler.Enabled(context.Background(), slog.LevelInfo))
	assert.True(t, handler.Enabled(context.Background(), slog.LevelWarn))
}

func TestCreateLogHandler_WhenJSON_RespectsLevel(t *testing.T) {
	level := slog.LevelWarn
	output := &bytes.Buffer{}

	handler, err := createLogHandler("json", output, level)

	require.NoError(t, err)
	assert.NotNil(t, handler)
	assert.False(t, handler.Enabled(context.Background(), slog.LevelInfo))
	assert.True(t, handler.Enabled(context.Background(), slog.LevelWarn))
}

func TestCreateLoggerFromConfig_WhenInvalidLevel_ReturnsWrappedError(t *testing.T) {
	config := Config{
		Level:  "verbose",
		Format: "text-no-color",
	}

	logger, err := CreateLoggerFromConfig(config)

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrUnknownLogLevel))
	assert.Equal(t, "parse log level: unknown log level: verbose", err.Error())
	assert.Nil(t, logger)
}

func TestCreateLoggerFromConfig_WhenInvalidFormat_ReturnsWrappedError(t *testing.T) {
	config := Config{
		Level:  "info",
		Format: "custom",
	}

	logger, err := CreateLoggerFromConfig(config)

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrUnknownLogFormat))
	assert.Equal(t, "create log handler: unknown log format: custom", err.Error())
	assert.Nil(t, logger)
}

func TestCreateLoggerFromConfig_WhenValidConfig_ReturnsLogger(t *testing.T) {
	config := Config{
		Level:  "info",
		Format: "json",
	}

	logger, err := CreateLoggerFromConfig(config)

	require.NoError(t, err)
	assert.NotNil(t, logger)
}

func TestCreateLoggerFromConfig_WhenQuietFalse_WritesToStderr(t *testing.T) {
	stderr := swapStderr(t)
	config := Config{
		Level:  "info",
		Format: "text-no-color",
		Quiet:  false,
	}

	logger, err := CreateLoggerFromConfig(config)
	require.NoError(t, err)

	logger.Info("Test message.")
	require.NoError(t, stderr.Sync())
	data, readErr := os.ReadFile(stderr.Name())
	require.NoError(t, readErr)

	assert.NotEmpty(t, data)
}

func TestCreateLoggerFromConfig_WhenQuietTrue_DiscardsOutput(t *testing.T) {
	stderr := swapStderr(t)
	config := Config{
		Level:  "info",
		Format: "text-no-color",
		Quiet:  true,
	}

	logger, err := CreateLoggerFromConfig(config)
	require.NoError(t, err)

	logger.Info("Test message.")
	require.NoError(t, stderr.Sync())
	data, readErr := os.ReadFile(stderr.Name())
	require.NoError(t, readErr)

	assert.Empty(t, data)
}

func swapStderr(t *testing.T) *os.File {
	t.Helper()

	file, err := os.CreateTemp("", "stderr-*")
	require.NoError(t, err)

	original := os.Stderr
	os.Stderr = file

	t.Cleanup(func() {
		os.Stderr = original
		assert.NoError(t, file.Close())
		assert.NoError(t, os.Remove(file.Name()))
	})

	return file
}
