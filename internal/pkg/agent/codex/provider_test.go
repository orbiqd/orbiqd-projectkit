package codex

import (
	"testing"

	agentAPI "github.com/orbiqd/orbiqd-projectkit/pkg/agent"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvider_NewAgent_WhenValidOptions_ThenReturnsAgent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                        string
		options                     any
		expectedInstructionsFileName string
	}{
		{
			name:                        "Options struct",
			options:                     Options{InstructionsFileName: "custom.md"},
			expectedInstructionsFileName: "custom.md",
		},
		{
			name:                        "map with matching key",
			options:                     map[string]any{"instructionsFileName": "custom.md"},
			expectedInstructionsFileName: "custom.md",
		},
		{
			name:                        "nil options",
			options:                     nil,
			expectedInstructionsFileName: "AGENTS.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fs := afero.NewMemMapFs()
			provider := NewProvider(fs)

			result, err := provider.NewAgent(tt.options)

			require.NoError(t, err)
			agent, ok := result.(*Agent)
			require.True(t, ok, "result should be *Agent")
			assert.Equal(t, tt.expectedInstructionsFileName, agent.options.InstructionsFileName)
		})
	}
}

func TestProvider_NewAgent_WhenInvalidOptions_ThenReturnsError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		options     any
		errContains string
	}{
		{
			name:        "un-marshallable value",
			options:     func() {},
			errContains: "json marshal",
		},
		{
			name:        "type mismatch",
			options:     map[string]any{"instructionsFileName": []string{"a"}},
			errContains: "json unmarshal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fs := afero.NewMemMapFs()
			provider := NewProvider(fs)

			result, err := provider.NewAgent(tt.options)

			require.Error(t, err)
			assert.Nil(t, result)
			assert.ErrorIs(t, err, ErrInvalidOptions)
			assert.ErrorContains(t, err, tt.errContains)
		})
	}
}

func TestProvider_GetKind_ThenReturnsCodexKind(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	provider := NewProvider(fs)

	kind := provider.GetKind()

	assert.Equal(t, agentAPI.Kind("codex"), kind)
}
