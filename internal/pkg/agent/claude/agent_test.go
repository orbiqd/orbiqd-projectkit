package claude

import (
	"testing"

	instructionAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/instruction"
	skillAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/skill"
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
		expectedPatterns     []string
	}{
		{
			name:                 "default file name",
			instructionsFileName: "",
			expectedPatterns:     []string{"CLAUDE.md", ".claude"},
		},
		{
			name:                 "custom file name",
			instructionsFileName: "custom-instructions.md",
			expectedPatterns:     []string{"custom-instructions.md", ".claude"},
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

			require.Len(t, patterns, 2)
			assert.Equal(t, tt.expectedPatterns, patterns)
		})
	}
}

func TestAgent_RebuildSkills_WhenNoSkills_ThenDoesNotCreateDirectory(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	agent := NewAgent(Options{}, fs)

	mockRepo := skillAPI.NewMockRepository(t)
	mockRepo.EXPECT().GetAll().Return([]skillAPI.Skill{}, nil)

	err := agent.RebuildSkills(mockRepo)
	require.NoError(t, err)

	exists, err := afero.DirExists(fs, ".claude/skills")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestAgent_RebuildSkills_WhenSingleSkillWithoutScripts_ThenCreatesSkillDirectory(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	agent := NewAgent(Options{}, fs)

	skill := skillAPI.Skill{
		Metadata: skillAPI.Metadata{
			Name:        "test-skill",
			Description: "A test skill",
		},
		Instructions: "Test instructions content",
		Scripts:      nil,
	}

	mockRepo := skillAPI.NewMockRepository(t)
	mockRepo.EXPECT().GetAll().Return([]skillAPI.Skill{skill}, nil)

	err := agent.RebuildSkills(mockRepo)
	require.NoError(t, err)

	content, err := afero.ReadFile(fs, ".claude/skills/test-skill/SKILL.md")
	require.NoError(t, err)

	expectedContent := `---
name: test-skill
description: A test skill
---

Test instructions content`
	assert.Equal(t, expectedContent, string(content))

	exists, err := afero.DirExists(fs, ".claude/skills/test-skill/scripts")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestAgent_RebuildSkills_WhenSkillWithScripts_ThenCreatesScriptsDirectory(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	agent := NewAgent(Options{}, fs)

	skill := skillAPI.Skill{
		Metadata: skillAPI.Metadata{
			Name:        "git-commit",
			Description: "Create a git commit",
		},
		Instructions: "Commit instructions",
		Scripts: map[skillAPI.ScriptName]skillAPI.Script{
			"git-commit.sh": {
				ContentType: "application/x-sh",
				Content:     []byte("#!/bin/bash\necho 'test'"),
			},
		},
	}

	mockRepo := skillAPI.NewMockRepository(t)
	mockRepo.EXPECT().GetAll().Return([]skillAPI.Skill{skill}, nil)

	err := agent.RebuildSkills(mockRepo)
	require.NoError(t, err)

	skillContent, err := afero.ReadFile(fs, ".claude/skills/git-commit/SKILL.md")
	require.NoError(t, err)
	assert.Contains(t, string(skillContent), "git-commit")
	assert.Contains(t, string(skillContent), "Create a git commit")
	assert.Contains(t, string(skillContent), "Commit instructions")

	scriptContent, err := afero.ReadFile(fs, ".claude/skills/git-commit/scripts/git-commit.sh")
	require.NoError(t, err)
	assert.Equal(t, "#!/bin/bash\necho 'test'", string(scriptContent))

	info, err := fs.Stat(".claude/skills/git-commit/scripts/git-commit.sh")
	require.NoError(t, err)
	assert.Equal(t, 0755, int(info.Mode().Perm()))
}

func TestAgent_RebuildSkills_WhenMultipleSkills_ThenCreatesAllSkills(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	agent := NewAgent(Options{}, fs)

	skills := []skillAPI.Skill{
		{
			Metadata: skillAPI.Metadata{
				Name:        "skill-one",
				Description: "First skill",
			},
			Instructions: "First instructions",
		},
		{
			Metadata: skillAPI.Metadata{
				Name:        "skill-two",
				Description: "Second skill",
			},
			Instructions: "Second instructions",
		},
	}

	mockRepo := skillAPI.NewMockRepository(t)
	mockRepo.EXPECT().GetAll().Return(skills, nil)

	err := agent.RebuildSkills(mockRepo)
	require.NoError(t, err)

	content1, err := afero.ReadFile(fs, ".claude/skills/skill-one/SKILL.md")
	require.NoError(t, err)
	assert.Contains(t, string(content1), "skill-one")
	assert.Contains(t, string(content1), "First instructions")

	content2, err := afero.ReadFile(fs, ".claude/skills/skill-two/SKILL.md")
	require.NoError(t, err)
	assert.Contains(t, string(content2), "skill-two")
	assert.Contains(t, string(content2), "Second instructions")
}

func TestAgent_RebuildSkills_WhenExistingSkills_ThenRemovesOldSkills(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	agent := NewAgent(Options{}, fs)

	err := fs.MkdirAll(".claude/skills/old-skill", 0755)
	require.NoError(t, err)
	err = afero.WriteFile(fs, ".claude/skills/old-skill/SKILL.md", []byte("old content"), 0644)
	require.NoError(t, err)

	skill := skillAPI.Skill{
		Metadata: skillAPI.Metadata{
			Name:        "new-skill",
			Description: "New skill",
		},
		Instructions: "New instructions",
	}

	mockRepo := skillAPI.NewMockRepository(t)
	mockRepo.EXPECT().GetAll().Return([]skillAPI.Skill{skill}, nil)

	err = agent.RebuildSkills(mockRepo)
	require.NoError(t, err)

	exists, err := afero.Exists(fs, ".claude/skills/old-skill")
	require.NoError(t, err)
	assert.False(t, exists)

	content, err := afero.ReadFile(fs, ".claude/skills/new-skill/SKILL.md")
	require.NoError(t, err)
	assert.Contains(t, string(content), "new-skill")
}

func TestAgent_RebuildSkills_WhenRepositoryFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	agent := NewAgent(Options{}, fs)

	mockRepo := skillAPI.NewMockRepository(t)
	mockRepo.EXPECT().GetAll().Return(nil, assert.AnError)

	err := agent.RebuildSkills(mockRepo)

	require.Error(t, err)
	assert.ErrorContains(t, err, "skills retrieval")
}

func TestAgent_RebuildSkills_WhenCustomSkillsDir_ThenUsesCustomPath(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	agent := NewAgent(Options{SkillsDirName: "custom-skills"}, fs)

	skill := skillAPI.Skill{
		Metadata: skillAPI.Metadata{
			Name:        "test-skill",
			Description: "Test skill",
		},
		Instructions: "Test instructions",
	}

	mockRepo := skillAPI.NewMockRepository(t)
	mockRepo.EXPECT().GetAll().Return([]skillAPI.Skill{skill}, nil)

	err := agent.RebuildSkills(mockRepo)
	require.NoError(t, err)

	content, err := afero.ReadFile(fs, ".claude/custom-skills/test-skill/SKILL.md")
	require.NoError(t, err)
	assert.Contains(t, string(content), "test-skill")

	exists, err := afero.Exists(fs, ".claude/skills")
	require.NoError(t, err)
	assert.False(t, exists)
}
