package claude

import (
	"testing"

	instructionAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/instruction"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgent_RenderInstructions_WhenInstructionsProvided_ThenWritesMarkdownFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		instructions []instructionAPI.Instructions
		expected     string
	}{
		{
			name:         "empty slice",
			instructions: []instructionAPI.Instructions{},
			expected:     "# Claude Code Instructions\n\n",
		},
		{
			name: "single instruction, single rule",
			instructions: []instructionAPI.Instructions{
				{
					Category: "general",
					Rules:    []instructionAPI.Rule{"Use proper formatting"},
				},
			},
			expected: "# Claude Code Instructions\n\n## General\n\n- Use proper formatting\n\n",
		},
		{
			name: "single instruction, multiple rules",
			instructions: []instructionAPI.Instructions{
				{
					Category: "general",
					Rules: []instructionAPI.Rule{
						"Use proper formatting",
						"Write clear code",
						"Add documentation",
					},
				},
			},
			expected: "# Claude Code Instructions\n\n## General\n\n- Use proper formatting\n- Write clear code\n- Add documentation\n\n",
		},
		{
			name: "multiple instructions",
			instructions: []instructionAPI.Instructions{
				{
					Category: "general",
					Rules: []instructionAPI.Rule{
						"Use proper formatting",
						"Write clear code",
					},
				},
				{
					Category: "testing",
					Rules: []instructionAPI.Rule{
						"Write unit tests",
						"Use table-driven tests",
					},
				},
			},
			expected: "# Claude Code Instructions\n\n## General\n\n- Use proper formatting\n- Write clear code\n\n## Testing\n\n- Write unit tests\n- Use table-driven tests\n\n",
		},
		{
			name: "kebab-case category",
			instructions: []instructionAPI.Instructions{
				{
					Category: "user-communication",
					Rules:    []instructionAPI.Rule{"Be clear and concise"},
				},
			},
			expected: "# Claude Code Instructions\n\n## User Communication\n\n- Be clear and concise\n\n",
		},
		{
			name: "camelCase category",
			instructions: []instructionAPI.Instructions{
				{
					Category: "codingStyle",
					Rules:    []instructionAPI.Rule{"Follow conventions"},
				},
			},
			expected: "# Claude Code Instructions\n\n## Coding Style\n\n- Follow conventions\n\n",
		},
		{
			name: "snake_case category",
			instructions: []instructionAPI.Instructions{
				{
					Category: "unit_tests",
					Rules:    []instructionAPI.Rule{"Test all edge cases"},
				},
			},
			expected: "# Claude Code Instructions\n\n## Unit Tests\n\n- Test all edge cases\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fs := afero.NewMemMapFs()

			agent := NewAgent(Options{}, fs)

			err := agent.RenderInstructions(tt.instructions)
			require.NoError(t, err)

			content, err := afero.ReadFile(fs, "CLAUDE.md")
			require.NoError(t, err)

			assert.Equal(t, tt.expected, string(content))
		})
	}
}

func TestAgent_RenderInstructions_WhenFileSystemFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	fs := afero.NewReadOnlyFs(afero.NewMemMapFs())

	agent := NewAgent(Options{}, fs)

	instructions := []instructionAPI.Instructions{
		{
			Category: "general",
			Rules:    []instructionAPI.Rule{"Test rule"},
		},
	}

	err := agent.RenderInstructions(instructions)

	require.Error(t, err)
	assert.ErrorContains(t, err, "instructions file write")
}

func TestAgent_RenderInstructions_WhenCustomFileName_ThenWritesToCustomFile(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()

	agent := NewAgent(Options{InstructionsFileName: "custom-instructions.md"}, fs)

	instructions := []instructionAPI.Instructions{
		{
			Category: "general",
			Rules:    []instructionAPI.Rule{"Test rule"},
		},
	}

	err := agent.RenderInstructions(instructions)
	require.NoError(t, err)

	content, err := afero.ReadFile(fs, "custom-instructions.md")
	require.NoError(t, err)
	assert.Contains(t, string(content), "Test rule")

	exists, err := afero.Exists(fs, "CLAUDE.md")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestAgent_GitIgnorePatterns_ThenReturnsInstructionsFileName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		instructionsFileName string
		expectedPattern      string
	}{
		{
			name:                 "default file name",
			instructionsFileName: "",
			expectedPattern:      "CLAUDE.md",
		},
		{
			name:                 "custom file name",
			instructionsFileName: "custom-instructions.md",
			expectedPattern:      "custom-instructions.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fs := afero.NewMemMapFs()

			var agent *Agent
			if tt.instructionsFileName == "" {
				agent = NewAgent(Options{}, fs)
			} else {
				agent = NewAgent(Options{InstructionsFileName: tt.instructionsFileName}, fs)
			}

			patterns := agent.GitIgnorePatterns()

			require.Len(t, patterns, 1)
			assert.Equal(t, tt.expectedPattern, patterns[0])
		})
	}
}
