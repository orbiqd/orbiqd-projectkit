package loader

import (
	"errors"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	instructionAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/instruction"
	sourceAPI "github.com/orbiqd/orbiqd-projectkit/pkg/source"
)

func validInstructionFs(t *testing.T, category string, rules []string) afero.Fs {
	t.Helper()

	fs := afero.NewMemMapFs()

	content := "category: " + category + "\nrules:\n"
	for _, rule := range rules {
		content += "  - " + rule + "\n"
	}

	require.NoError(t, afero.WriteFile(fs, category+".yaml", []byte(content), 0644))

	return fs
}

func TestLoadAiInstructionsFromConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		sources             []instructionAPI.SourceConfig
		mockSetup           func(*sourceAPI.MockResolver)
		wantInstructionsLen int
	}{
		{
			name:                "WhenNoSources_ThenReturnsNilInstructions",
			sources:             []instructionAPI.SourceConfig{},
			mockSetup:           func(m *sourceAPI.MockResolver) {},
			wantInstructionsLen: 0,
		},
		{
			name: "WhenSingleSourceWithOneInstruction_ThenReturnsInstruction",
			sources: []instructionAPI.SourceConfig{
				{URI: "file://./instructions"},
			},
			mockSetup: func(m *sourceAPI.MockResolver) {
				fs := validInstructionFs(t, "test-category", []string{"Rule one", "Rule two"})
				m.EXPECT().Resolve("file://./instructions").Return(fs, nil)
			},
			wantInstructionsLen: 1,
		},
		{
			name: "WhenMultipleSources_ThenReturnsCombinedInstructions",
			sources: []instructionAPI.SourceConfig{
				{URI: "file://./instructions1"},
				{URI: "file://./instructions2"},
			},
			mockSetup: func(m *sourceAPI.MockResolver) {
				fs1 := validInstructionFs(t, "category-one", []string{"First rule"})
				fs2 := validInstructionFs(t, "category-two", []string{"Second rule"})
				m.EXPECT().Resolve("file://./instructions1").Return(fs1, nil)
				m.EXPECT().Resolve("file://./instructions2").Return(fs2, nil)
			},
			wantInstructionsLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockResolver := sourceAPI.NewMockResolver(t)
			tt.mockSetup(mockResolver)

			config := instructionAPI.Config{
				Sources: tt.sources,
			}

			instructions, err := LoadAiInstructionsFromConfig(config, mockResolver)

			require.NoError(t, err)
			assert.Len(t, instructions, tt.wantInstructionsLen)
		})
	}
}

func TestLoadAiInstructionsFromConfig_WhenResolverFails_ThenReturnsResolveError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	resolveErr := errors.New("resolver failed")
	mockResolver.EXPECT().Resolve("file://./instructions").Return(nil, resolveErr)

	config := instructionAPI.Config{
		Sources: []instructionAPI.SourceConfig{
			{URI: "file://./instructions"},
		},
	}

	instructions, err := LoadAiInstructionsFromConfig(config, mockResolver)

	require.Error(t, err)
	assert.Nil(t, instructions)
	assert.Contains(t, err.Error(), "resolve: file://./instructions")
	assert.ErrorIs(t, err, resolveErr)
}

func TestLoadAiInstructionsFromConfig_WhenLoaderFails_ThenReturnsLoadInstructionsError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	emptyFs := afero.NewMemMapFs()
	mockResolver.EXPECT().Resolve("file://./instructions").Return(emptyFs, nil)

	config := instructionAPI.Config{
		Sources: []instructionAPI.SourceConfig{
			{URI: "file://./instructions"},
		},
	}

	instructions, err := LoadAiInstructionsFromConfig(config, mockResolver)

	require.Error(t, err)
	assert.Nil(t, instructions)
	assert.Contains(t, err.Error(), "load instructions:")
	assert.ErrorIs(t, err, instructionAPI.ErrNoInstructionsFound)
}

func TestLoadAiInstructionsFromConfig_WhenSecondSourceResolverFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	fs1 := validInstructionFs(t, "category-one", []string{"First rule"})
	resolveErr := errors.New("second resolver failed")

	mockResolver.EXPECT().Resolve("file://./instructions1").Return(fs1, nil)
	mockResolver.EXPECT().Resolve("file://./instructions2").Return(nil, resolveErr)

	config := instructionAPI.Config{
		Sources: []instructionAPI.SourceConfig{
			{URI: "file://./instructions1"},
			{URI: "file://./instructions2"},
		},
	}

	instructions, err := LoadAiInstructionsFromConfig(config, mockResolver)

	require.Error(t, err)
	assert.Nil(t, instructions)
	assert.Contains(t, err.Error(), "resolve: file://./instructions2")
	assert.ErrorIs(t, err, resolveErr)
}

func TestLoadAiInstructionsFromConfig_WhenSecondSourceLoaderFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	fs1 := validInstructionFs(t, "category-one", []string{"First rule"})
	emptyFs := afero.NewMemMapFs()

	mockResolver.EXPECT().Resolve("file://./instructions1").Return(fs1, nil)
	mockResolver.EXPECT().Resolve("file://./instructions2").Return(emptyFs, nil)

	config := instructionAPI.Config{
		Sources: []instructionAPI.SourceConfig{
			{URI: "file://./instructions1"},
			{URI: "file://./instructions2"},
		},
	}

	instructions, err := LoadAiInstructionsFromConfig(config, mockResolver)

	require.Error(t, err)
	assert.Nil(t, instructions)
	assert.Contains(t, err.Error(), "load instructions:")
	assert.ErrorIs(t, err, instructionAPI.ErrNoInstructionsFound)
}
