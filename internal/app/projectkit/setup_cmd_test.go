package projectkit

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.nhat.io/aferomock"

	"github.com/orbiqd/orbiqd-projectkit/internal/pkg/git"
	agentAPI "github.com/orbiqd/orbiqd-projectkit/pkg/agent"
	aiAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai"
	instructionAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/instruction"
	skillAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/skill"
	projectAPI "github.com/orbiqd/orbiqd-projectkit/pkg/project"
	rulebookAPI "github.com/orbiqd/orbiqd-projectkit/pkg/rulebook"
	sourceAPI "github.com/orbiqd/orbiqd-projectkit/pkg/source"
)

func validSkillFs(t *testing.T, name, description, instructions string) afero.Fs {
	t.Helper()

	fs := afero.NewMemMapFs()

	require.NoError(t, fs.Mkdir(name, 0755))

	metadata := `name: ` + name + `
description: ` + description + `
`
	require.NoError(t, afero.WriteFile(fs, name+"/metadata.yaml", []byte(metadata), 0644))
	require.NoError(t, afero.WriteFile(fs, name+"/instructions.md", []byte(instructions), 0644))

	return fs
}

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

func validRulebookFs(t *testing.T, category string, rules []string, skillName, skillDescription, skillInstructions string) afero.Fs {
	t.Helper()

	fs := afero.NewMemMapFs()

	rulebookContent := `ai:
  instruction:
    sources:
      - uri: rulebook://ai/instructions
  skill:
    sources:
      - uri: rulebook://ai/skills
`
	require.NoError(t, afero.WriteFile(fs, "rulebook.yaml", []byte(rulebookContent), 0644))

	require.NoError(t, fs.MkdirAll("/ai/instructions", 0755))
	instructionContent := "category: " + category + "\nrules:\n"
	for _, rule := range rules {
		instructionContent += "  - " + rule + "\n"
	}
	require.NoError(t, afero.WriteFile(fs, "/ai/instructions/"+category+".yaml", []byte(instructionContent), 0644))

	require.NoError(t, fs.MkdirAll("/ai/skills/"+skillName, 0755))
	skillMetadata := `name: ` + skillName + `
description: ` + skillDescription + `
`
	require.NoError(t, afero.WriteFile(fs, "/ai/skills/"+skillName+"/metadata.yaml", []byte(skillMetadata), 0644))
	require.NoError(t, afero.WriteFile(fs, "/ai/skills/"+skillName+"/instructions.md", []byte(skillInstructions), 0644))

	return fs
}

func TestSetupCmd_loadSkillsFromConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		sources       []skillAPI.SourceConfig
		mockSetup     func(*sourceAPI.MockResolver)
		wantSkillsLen int
	}{
		{
			name:          "WhenNoSources_ThenReturnsNilSkills",
			sources:       []skillAPI.SourceConfig{},
			mockSetup:     func(m *sourceAPI.MockResolver) {},
			wantSkillsLen: 0,
		},
		{
			name: "WhenSingleSourceWithOneSkill_ThenReturnsSkill",
			sources: []skillAPI.SourceConfig{
				{URI: "file://./skills"},
			},
			mockSetup: func(m *sourceAPI.MockResolver) {
				fs := validSkillFs(t, "test-skill", "Test skill description", "Test instructions")
				m.EXPECT().Resolve("file://./skills").Return(fs, nil)
			},
			wantSkillsLen: 1,
		},
		{
			name: "WhenMultipleSources_ThenReturnsCombinedSkills",
			sources: []skillAPI.SourceConfig{
				{URI: "file://./skills1"},
				{URI: "file://./skills2"},
			},
			mockSetup: func(m *sourceAPI.MockResolver) {
				fs1 := validSkillFs(t, "skill-one", "First skill", "First instructions")
				fs2 := validSkillFs(t, "skill-two", "Second skill", "Second instructions")
				m.EXPECT().Resolve("file://./skills1").Return(fs1, nil)
				m.EXPECT().Resolve("file://./skills2").Return(fs2, nil)
			},
			wantSkillsLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockResolver := sourceAPI.NewMockResolver(t)
			tt.mockSetup(mockResolver)

			cmd := &SetupCmd{}
			config := skillAPI.Config{
				Sources: tt.sources,
			}

			skills, err := cmd.loadSkillsFromConfig(config, mockResolver)

			require.NoError(t, err)
			assert.Len(t, skills, tt.wantSkillsLen)
		})
	}
}

func TestSetupCmd_loadSkillsFromConfig_WhenResolverFails_ThenReturnsResolveError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	resolveErr := errors.New("resolver failed")
	mockResolver.EXPECT().Resolve("file://./skills").Return(nil, resolveErr)

	cmd := &SetupCmd{}
	config := skillAPI.Config{
		Sources: []skillAPI.SourceConfig{
			{URI: "file://./skills"},
		},
	}

	skills, err := cmd.loadSkillsFromConfig(config, mockResolver)

	require.Error(t, err)
	assert.Empty(t, skills)
	assert.Contains(t, err.Error(), "resolve: file://./skills")
	assert.ErrorIs(t, err, resolveErr)
}

func TestSetupCmd_loadSkillsFromConfig_WhenLoaderFails_ThenReturnsLoadSkillsError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	emptyFs := afero.NewMemMapFs()
	mockResolver.EXPECT().Resolve("file://./skills").Return(emptyFs, nil)

	cmd := &SetupCmd{}
	config := skillAPI.Config{
		Sources: []skillAPI.SourceConfig{
			{URI: "file://./skills"},
		},
	}

	skills, err := cmd.loadSkillsFromConfig(config, mockResolver)

	require.Error(t, err)
	assert.Empty(t, skills)
	assert.Contains(t, err.Error(), "load skills:")
	assert.ErrorIs(t, err, skillAPI.ErrNoSkillsFound)
}

func TestSetupCmd_loadSkillsFromConfig_WhenSecondSourceResolverFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	fs1 := validSkillFs(t, "skill-one", "First skill", "First instructions")
	resolveErr := errors.New("second resolver failed")

	mockResolver.EXPECT().Resolve("file://./skills1").Return(fs1, nil)
	mockResolver.EXPECT().Resolve("file://./skills2").Return(nil, resolveErr)

	cmd := &SetupCmd{}
	config := skillAPI.Config{
		Sources: []skillAPI.SourceConfig{
			{URI: "file://./skills1"},
			{URI: "file://./skills2"},
		},
	}

	skills, err := cmd.loadSkillsFromConfig(config, mockResolver)

	require.Error(t, err)
	assert.Empty(t, skills)
	assert.Contains(t, err.Error(), "resolve: file://./skills2")
	assert.ErrorIs(t, err, resolveErr)
}

func TestSetupCmd_loadSkillsFromConfig_WhenSecondSourceLoaderFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	fs1 := validSkillFs(t, "skill-one", "First skill", "First instructions")
	emptyFs := afero.NewMemMapFs()

	mockResolver.EXPECT().Resolve("file://./skills1").Return(fs1, nil)
	mockResolver.EXPECT().Resolve("file://./skills2").Return(emptyFs, nil)

	cmd := &SetupCmd{}
	config := skillAPI.Config{
		Sources: []skillAPI.SourceConfig{
			{URI: "file://./skills1"},
			{URI: "file://./skills2"},
		},
	}

	skills, err := cmd.loadSkillsFromConfig(config, mockResolver)

	require.Error(t, err)
	assert.Empty(t, skills)
	assert.Contains(t, err.Error(), "load skills:")
	assert.ErrorIs(t, err, skillAPI.ErrNoSkillsFound)
}

func TestSetupCmd_loadInstructionsFromConfig(t *testing.T) {
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

			cmd := &SetupCmd{}
			config := instructionAPI.Config{
				Sources: tt.sources,
			}

			instructions, err := cmd.loadInstructionsFromConfig(config, mockResolver)

			require.NoError(t, err)
			assert.Len(t, instructions, tt.wantInstructionsLen)
		})
	}
}

func TestSetupCmd_loadInstructionsFromConfig_WhenResolverFails_ThenReturnsResolveError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	resolveErr := errors.New("resolver failed")
	mockResolver.EXPECT().Resolve("file://./instructions").Return(nil, resolveErr)

	cmd := &SetupCmd{}
	config := instructionAPI.Config{
		Sources: []instructionAPI.SourceConfig{
			{URI: "file://./instructions"},
		},
	}

	instructions, err := cmd.loadInstructionsFromConfig(config, mockResolver)

	require.Error(t, err)
	assert.Empty(t, instructions)
	assert.Contains(t, err.Error(), "resolve: file://./instructions")
	assert.ErrorIs(t, err, resolveErr)
}

func TestSetupCmd_loadInstructionsFromConfig_WhenLoaderFails_ThenReturnsLoadInstructionsError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	emptyFs := afero.NewMemMapFs()
	mockResolver.EXPECT().Resolve("file://./instructions").Return(emptyFs, nil)

	cmd := &SetupCmd{}
	config := instructionAPI.Config{
		Sources: []instructionAPI.SourceConfig{
			{URI: "file://./instructions"},
		},
	}

	instructions, err := cmd.loadInstructionsFromConfig(config, mockResolver)

	require.Error(t, err)
	assert.Empty(t, instructions)
	assert.Contains(t, err.Error(), "load instructions:")
	assert.ErrorIs(t, err, instructionAPI.ErrNoInstructionsFound)
}

func TestSetupCmd_loadInstructionsFromConfig_WhenSecondSourceResolverFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	fs1 := validInstructionFs(t, "category-one", []string{"First rule"})
	resolveErr := errors.New("second resolver failed")

	mockResolver.EXPECT().Resolve("file://./instructions1").Return(fs1, nil)
	mockResolver.EXPECT().Resolve("file://./instructions2").Return(nil, resolveErr)

	cmd := &SetupCmd{}
	config := instructionAPI.Config{
		Sources: []instructionAPI.SourceConfig{
			{URI: "file://./instructions1"},
			{URI: "file://./instructions2"},
		},
	}

	instructions, err := cmd.loadInstructionsFromConfig(config, mockResolver)

	require.Error(t, err)
	assert.Empty(t, instructions)
	assert.Contains(t, err.Error(), "resolve: file://./instructions2")
	assert.ErrorIs(t, err, resolveErr)
}

func TestSetupCmd_loadInstructionsFromConfig_WhenSecondSourceLoaderFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	fs1 := validInstructionFs(t, "category-one", []string{"First rule"})
	emptyFs := afero.NewMemMapFs()

	mockResolver.EXPECT().Resolve("file://./instructions1").Return(fs1, nil)
	mockResolver.EXPECT().Resolve("file://./instructions2").Return(emptyFs, nil)

	cmd := &SetupCmd{}
	config := instructionAPI.Config{
		Sources: []instructionAPI.SourceConfig{
			{URI: "file://./instructions1"},
			{URI: "file://./instructions2"},
		},
	}

	instructions, err := cmd.loadInstructionsFromConfig(config, mockResolver)

	require.Error(t, err)
	assert.Empty(t, instructions)
	assert.Contains(t, err.Error(), "load instructions:")
	assert.ErrorIs(t, err, instructionAPI.ErrNoInstructionsFound)
}

func TestSetupCmd_loadAgentsFromConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		agents        []agentAPI.Config
		mockSetup     func(*agentAPI.MockRegistry)
		wantAgentsLen int
	}{
		{
			name:          "WhenNoAgents_ThenReturnsNilAgents",
			agents:        []agentAPI.Config{},
			mockSetup:     func(m *agentAPI.MockRegistry) {},
			wantAgentsLen: 0,
		},
		{
			name: "WhenSingleAgent_ThenReturnsOneAgent",
			agents: []agentAPI.Config{
				{Kind: "test-agent"},
			},
			mockSetup: func(m *agentAPI.MockRegistry) {
				mockProvider := agentAPI.NewMockProvider(t)
				mockAgent := agentAPI.NewMockAgent(t)

				m.EXPECT().GetByKind(agentAPI.Kind("test-agent")).Return(mockProvider, nil)
				mockProvider.EXPECT().NewAgent(nil).Return(mockAgent, nil)
				mockAgent.EXPECT().GetKind().Return(agentAPI.Kind("test-agent"))
			},
			wantAgentsLen: 1,
		},
		{
			name: "WhenMultipleAgents_ThenReturnsAllAgents",
			agents: []agentAPI.Config{
				{Kind: "agent-one"},
				{Kind: "agent-two"},
			},
			mockSetup: func(m *agentAPI.MockRegistry) {
				mockProvider1 := agentAPI.NewMockProvider(t)
				mockAgent1 := agentAPI.NewMockAgent(t)
				m.EXPECT().GetByKind(agentAPI.Kind("agent-one")).Return(mockProvider1, nil)
				mockProvider1.EXPECT().NewAgent(nil).Return(mockAgent1, nil)
				mockAgent1.EXPECT().GetKind().Return(agentAPI.Kind("agent-one"))

				mockProvider2 := agentAPI.NewMockProvider(t)
				mockAgent2 := agentAPI.NewMockAgent(t)
				m.EXPECT().GetByKind(agentAPI.Kind("agent-two")).Return(mockProvider2, nil)
				mockProvider2.EXPECT().NewAgent(nil).Return(mockAgent2, nil)
				mockAgent2.EXPECT().GetKind().Return(agentAPI.Kind("agent-two"))
			},
			wantAgentsLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRegistry := agentAPI.NewMockRegistry(t)
			tt.mockSetup(mockRegistry)

			cmd := &SetupCmd{}
			config := projectAPI.Config{
				Agents: tt.agents,
			}

			agents, err := cmd.loadAgentsFromConfig(config, mockRegistry)

			require.NoError(t, err)
			assert.Len(t, agents, tt.wantAgentsLen)
		})
	}
}

func TestSetupCmd_loadAgentsFromConfig_WhenRegistryGetByKindFails_ThenReturnsGetProviderError(t *testing.T) {
	t.Parallel()

	mockRegistry := agentAPI.NewMockRegistry(t)
	getByKindErr := errors.New("registry get by kind failed")
	mockRegistry.EXPECT().GetByKind(agentAPI.Kind("test-agent")).Return(nil, getByKindErr)

	cmd := &SetupCmd{}
	config := projectAPI.Config{
		Agents: []agentAPI.Config{
			{Kind: "test-agent"},
		},
	}

	agents, err := cmd.loadAgentsFromConfig(config, mockRegistry)

	require.Error(t, err)
	assert.Empty(t, agents)
	assert.Contains(t, err.Error(), "get provider: test-agent")
	assert.ErrorIs(t, err, getByKindErr)
}

func TestSetupCmd_loadAgentsFromConfig_WhenProviderNewAgentFails_ThenReturnsCreateAgentError(t *testing.T) {
	t.Parallel()

	mockRegistry := agentAPI.NewMockRegistry(t)
	mockProvider := agentAPI.NewMockProvider(t)
	newAgentErr := errors.New("provider new agent failed")

	mockRegistry.EXPECT().GetByKind(agentAPI.Kind("test-agent")).Return(mockProvider, nil)
	mockProvider.EXPECT().NewAgent(nil).Return(nil, newAgentErr)

	cmd := &SetupCmd{}
	config := projectAPI.Config{
		Agents: []agentAPI.Config{
			{Kind: "test-agent"},
		},
	}

	agents, err := cmd.loadAgentsFromConfig(config, mockRegistry)

	require.Error(t, err)
	assert.Empty(t, agents)
	assert.Contains(t, err.Error(), "create agent: test-agent")
	assert.ErrorIs(t, err, newAgentErr)
}

func TestSetupCmd_loadAgentsFromConfig_WhenSecondAgentGetByKindFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockRegistry := agentAPI.NewMockRegistry(t)
	mockProvider1 := agentAPI.NewMockProvider(t)
	mockAgent1 := agentAPI.NewMockAgent(t)
	getByKindErr := errors.New("second registry get by kind failed")

	mockRegistry.EXPECT().GetByKind(agentAPI.Kind("agent-one")).Return(mockProvider1, nil)
	mockProvider1.EXPECT().NewAgent(nil).Return(mockAgent1, nil)
	mockAgent1.EXPECT().GetKind().Return(agentAPI.Kind("agent-one"))

	mockRegistry.EXPECT().GetByKind(agentAPI.Kind("failing-agent")).Return(nil, getByKindErr)

	cmd := &SetupCmd{}
	config := projectAPI.Config{
		Agents: []agentAPI.Config{
			{Kind: "agent-one"},
			{Kind: "failing-agent"},
		},
	}

	agents, err := cmd.loadAgentsFromConfig(config, mockRegistry)

	require.Error(t, err)
	assert.Empty(t, agents)
	assert.Contains(t, err.Error(), "get provider: failing-agent")
	assert.ErrorIs(t, err, getByKindErr)
}

func TestSetupCmd_loadAgentsFromConfig_WhenSecondAgentNewAgentFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockRegistry := agentAPI.NewMockRegistry(t)
	mockProvider1 := agentAPI.NewMockProvider(t)
	mockAgent1 := agentAPI.NewMockAgent(t)
	mockProvider2 := agentAPI.NewMockProvider(t)
	newAgentErr := errors.New("second provider new agent failed")

	mockRegistry.EXPECT().GetByKind(agentAPI.Kind("agent-one")).Return(mockProvider1, nil)
	mockProvider1.EXPECT().NewAgent(nil).Return(mockAgent1, nil)
	mockAgent1.EXPECT().GetKind().Return(agentAPI.Kind("agent-one"))

	mockRegistry.EXPECT().GetByKind(agentAPI.Kind("failing-agent")).Return(mockProvider2, nil)
	mockProvider2.EXPECT().NewAgent(nil).Return(nil, newAgentErr)

	cmd := &SetupCmd{}
	config := projectAPI.Config{
		Agents: []agentAPI.Config{
			{Kind: "agent-one"},
			{Kind: "failing-agent"},
		},
	}

	agents, err := cmd.loadAgentsFromConfig(config, mockRegistry)

	require.Error(t, err)
	assert.Empty(t, agents)
	assert.Contains(t, err.Error(), "create agent: failing-agent")
	assert.ErrorIs(t, err, newAgentErr)
}

func TestSetupCmd_setupAgentGit(t *testing.T) {
	t.Parallel()

	const currentDir = "/project"

	tests := []struct {
		name      string
		setupFs   func(*testing.T) afero.Fs
		patterns  []string
		mockSetup func(*agentAPI.MockAgent)
		assertFs  func(*testing.T, afero.Fs)
	}{
		{
			name: "WhenNoPatterns_ThenDoesNotModifyExclude",
			setupFs: func(t *testing.T) afero.Fs {
				t.Helper()
				fs := afero.NewMemMapFs()
				require.NoError(t, fs.Mkdir(currentDir, 0755))
				require.NoError(t, fs.Mkdir(currentDir+"/.git", 0755))
				return fs
			},
			patterns: []string{},
			mockSetup: func(m *agentAPI.MockAgent) {
				m.EXPECT().GitIgnorePatterns().Return([]string{})
			},
			assertFs: func(t *testing.T, fs afero.Fs) {
				t.Helper()
				exists, err := afero.Exists(fs, currentDir+"/.git/info/exclude")
				require.NoError(t, err)
				assert.False(t, exists)
			},
		},
		{
			name: "WhenSinglePattern_ThenAddsToExclude",
			setupFs: func(t *testing.T) afero.Fs {
				t.Helper()
				fs := afero.NewMemMapFs()
				require.NoError(t, fs.Mkdir(currentDir, 0755))
				require.NoError(t, fs.Mkdir(currentDir+"/.git", 0755))
				return fs
			},
			patterns: []string{".claude"},
			mockSetup: func(m *agentAPI.MockAgent) {
				m.EXPECT().GitIgnorePatterns().Return([]string{".claude"})
				m.EXPECT().GetKind().Return(agentAPI.Kind("claude")).Once()
			},
			assertFs: func(t *testing.T, fs afero.Fs) {
				t.Helper()
				content, err := afero.ReadFile(fs, currentDir+"/.git/info/exclude")
				require.NoError(t, err)
				assert.Equal(t, ".claude\n", string(content))
			},
		},
		{
			name: "WhenMultiplePatterns_ThenAddsAllToExclude",
			setupFs: func(t *testing.T) afero.Fs {
				t.Helper()
				fs := afero.NewMemMapFs()
				require.NoError(t, fs.Mkdir(currentDir, 0755))
				require.NoError(t, fs.Mkdir(currentDir+"/.git", 0755))
				return fs
			},
			patterns: []string{".claude", ".claude/**"},
			mockSetup: func(m *agentAPI.MockAgent) {
				m.EXPECT().GitIgnorePatterns().Return([]string{".claude", ".claude/**"})
				m.EXPECT().GetKind().Return(agentAPI.Kind("claude")).Times(2)
			},
			assertFs: func(t *testing.T, fs afero.Fs) {
				t.Helper()
				content, err := afero.ReadFile(fs, currentDir+"/.git/info/exclude")
				require.NoError(t, err)
				assert.Equal(t, ".claude\n.claude/**\n", string(content))
			},
		},
		{
			name: "WhenPatternAlreadyExcluded_ThenSkipsPattern",
			setupFs: func(t *testing.T) afero.Fs {
				t.Helper()
				fs := afero.NewMemMapFs()
				require.NoError(t, fs.Mkdir(currentDir, 0755))
				require.NoError(t, fs.Mkdir(currentDir+"/.git", 0755))
				require.NoError(t, fs.Mkdir(currentDir+"/.git/info", 0755))
				require.NoError(t, afero.WriteFile(fs, currentDir+"/.git/info/exclude", []byte(".claude\n"), 0644))
				return fs
			},
			patterns: []string{".claude"},
			mockSetup: func(m *agentAPI.MockAgent) {
				m.EXPECT().GitIgnorePatterns().Return([]string{".claude"})
			},
			assertFs: func(t *testing.T, fs afero.Fs) {
				t.Helper()
				content, err := afero.ReadFile(fs, currentDir+"/.git/info/exclude")
				require.NoError(t, err)
				assert.Equal(t, ".claude\n", string(content))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			projectFs := tt.setupFs(t)
			mockAgent := agentAPI.NewMockAgent(t)
			tt.mockSetup(mockAgent)

			cmd := &SetupCmd{}
			err := cmd.setupAgentGit(projectFs, currentDir, mockAgent)

			require.NoError(t, err)
			tt.assertFs(t, projectFs)
		})
	}
}

func TestSetupCmd_setupAgentGit_WhenGitRepoNotFound_ThenReturnsGitFsCreationError(t *testing.T) {
	t.Parallel()

	const currentDir = "/project"

	projectFs := afero.NewMemMapFs()
	require.NoError(t, projectFs.Mkdir(currentDir, 0755))

	mockAgent := agentAPI.NewMockAgent(t)

	cmd := &SetupCmd{}
	err := cmd.setupAgentGit(projectFs, currentDir, mockAgent)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "git fs creation")
	assert.ErrorIs(t, err, git.ErrGitRepoNotFound)
}

func TestSetupCmd_setupAgentGit_WhenIsExcludedFails_ThenReturnsCheckExcludedError(t *testing.T) {
	t.Parallel()

	const currentDir = "/project"

	baseFs := afero.NewMemMapFs()
	require.NoError(t, baseFs.Mkdir(currentDir, 0755))
	require.NoError(t, baseFs.Mkdir(currentDir+"/.git", 0755))

	simulatedErr := errors.New("simulated stat error")
	overrideFs := aferomock.OverrideFs(baseFs, aferomock.FsCallbacks{
		StatFunc: func(name string) (os.FileInfo, error) {
			if strings.Contains(name, "info/exclude") {
				return nil, simulatedErr
			}
			return baseFs.Stat(name)
		},
	})

	mockAgent := agentAPI.NewMockAgent(t)
	mockAgent.EXPECT().GitIgnorePatterns().Return([]string{".claude"})

	cmd := &SetupCmd{}
	err := cmd.setupAgentGit(overrideFs, currentDir, mockAgent)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "check excluded pattern")
	assert.ErrorIs(t, err, simulatedErr)
}

func TestSetupCmd_setupAgentGit_WhenExcludeFails_ThenReturnsExcludePatternError(t *testing.T) {
	t.Parallel()

	const currentDir = "/project"

	baseFs := afero.NewMemMapFs()
	require.NoError(t, baseFs.Mkdir(currentDir, 0755))
	require.NoError(t, baseFs.Mkdir(currentDir+"/.git", 0755))

	simulatedErr := errors.New("simulated openfile error")
	overrideFs := aferomock.OverrideFs(baseFs, aferomock.FsCallbacks{
		OpenFileFunc: func(name string, flag int, perm os.FileMode) (afero.File, error) {
			if strings.Contains(name, "info/exclude") {
				return nil, simulatedErr
			}
			return baseFs.OpenFile(name, flag, perm)
		},
	})

	mockAgent := agentAPI.NewMockAgent(t)
	mockAgent.EXPECT().GitIgnorePatterns().Return([]string{".claude"})

	cmd := &SetupCmd{}
	err := cmd.setupAgentGit(overrideFs, currentDir, mockAgent)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "exclude pattern")
	assert.ErrorIs(t, err, simulatedErr)
}

func TestSetupCmd_loadRulebooksFromConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		sources          []rulebookAPI.SourceConfig
		mockSetup        func(*sourceAPI.MockResolver)
		wantRulebooksLen int
	}{
		{
			name:             "WhenNoSources_ThenReturnsNilRulebooks",
			sources:          []rulebookAPI.SourceConfig{},
			mockSetup:        func(m *sourceAPI.MockResolver) {},
			wantRulebooksLen: 0,
		},
		{
			name: "WhenSingleSource_ThenReturnsRulebook",
			sources: []rulebookAPI.SourceConfig{
				{URI: "file://./rulebooks"},
			},
			mockSetup: func(m *sourceAPI.MockResolver) {
				fs := validRulebookFs(t, "test-category", []string{"Rule one"}, "test-skill", "Test skill", "Test instructions")
				m.EXPECT().Resolve("file://./rulebooks").Return(fs, nil)
			},
			wantRulebooksLen: 1,
		},
		{
			name: "WhenMultipleSources_ThenReturnsCombinedRulebooks",
			sources: []rulebookAPI.SourceConfig{
				{URI: "file://./rulebooks1"},
				{URI: "file://./rulebooks2"},
			},
			mockSetup: func(m *sourceAPI.MockResolver) {
				fs1 := validRulebookFs(t, "category-one", []string{"First rule"}, "skill-one", "First skill", "First instructions")
				fs2 := validRulebookFs(t, "category-two", []string{"Second rule"}, "skill-two", "Second skill", "Second instructions")
				m.EXPECT().Resolve("file://./rulebooks1").Return(fs1, nil)
				m.EXPECT().Resolve("file://./rulebooks2").Return(fs2, nil)
			},
			wantRulebooksLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockResolver := sourceAPI.NewMockResolver(t)
			tt.mockSetup(mockResolver)

			cmd := &SetupCmd{}
			config := rulebookAPI.Config{
				Sources: tt.sources,
			}

			rulebooks, err := cmd.loadRulebooksFromConfig(config, mockResolver)

			require.NoError(t, err)
			assert.Len(t, rulebooks, tt.wantRulebooksLen)
		})
	}
}

func TestSetupCmd_loadRulebooksFromConfig_WhenResolverFails_ThenReturnsResolveError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	resolveErr := errors.New("resolver failed")
	mockResolver.EXPECT().Resolve("file://./rulebooks").Return(nil, resolveErr)

	cmd := &SetupCmd{}
	config := rulebookAPI.Config{
		Sources: []rulebookAPI.SourceConfig{
			{URI: "file://./rulebooks"},
		},
	}

	rulebooks, err := cmd.loadRulebooksFromConfig(config, mockResolver)

	require.Error(t, err)
	assert.Empty(t, rulebooks)
	assert.Contains(t, err.Error(), "resolve rulebook uri")
	assert.ErrorIs(t, err, resolveErr)
}

func TestSetupCmd_loadRulebooksFromConfig_WhenLoaderFails_ThenReturnsLoadRulebookError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	emptyFs := afero.NewMemMapFs()
	mockResolver.EXPECT().Resolve("file://./rulebooks").Return(emptyFs, nil)

	cmd := &SetupCmd{}
	config := rulebookAPI.Config{
		Sources: []rulebookAPI.SourceConfig{
			{URI: "file://./rulebooks"},
		},
	}

	rulebooks, err := cmd.loadRulebooksFromConfig(config, mockResolver)

	require.Error(t, err)
	assert.Empty(t, rulebooks)
	assert.Contains(t, err.Error(), "load rulebook")
	assert.ErrorIs(t, err, rulebookAPI.ErrMissingMetadataFile)
}

func TestSetupCmd_loadRulebooksFromConfig_WhenSecondSourceResolverFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	fs1 := validRulebookFs(t, "category-one", []string{"First rule"}, "skill-one", "First skill", "First instructions")
	resolveErr := errors.New("second resolver failed")

	mockResolver.EXPECT().Resolve("file://./rulebooks1").Return(fs1, nil)
	mockResolver.EXPECT().Resolve("file://./rulebooks2").Return(nil, resolveErr)

	cmd := &SetupCmd{}
	config := rulebookAPI.Config{
		Sources: []rulebookAPI.SourceConfig{
			{URI: "file://./rulebooks1"},
			{URI: "file://./rulebooks2"},
		},
	}

	rulebooks, err := cmd.loadRulebooksFromConfig(config, mockResolver)

	require.Error(t, err)
	assert.Empty(t, rulebooks)
	assert.Contains(t, err.Error(), "resolve rulebook uri")
	assert.ErrorIs(t, err, resolveErr)
}

func TestSetupCmd_loadRulebooksFromConfig_WhenSecondSourceLoaderFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	fs1 := validRulebookFs(t, "category-one", []string{"First rule"}, "skill-one", "First skill", "First instructions")
	emptyFs := afero.NewMemMapFs()

	mockResolver.EXPECT().Resolve("file://./rulebooks1").Return(fs1, nil)
	mockResolver.EXPECT().Resolve("file://./rulebooks2").Return(emptyFs, nil)

	cmd := &SetupCmd{}
	config := rulebookAPI.Config{
		Sources: []rulebookAPI.SourceConfig{
			{URI: "file://./rulebooks1"},
			{URI: "file://./rulebooks2"},
		},
	}

	rulebooks, err := cmd.loadRulebooksFromConfig(config, mockResolver)

	require.Error(t, err)
	assert.Empty(t, rulebooks)
	assert.Contains(t, err.Error(), "load rulebook")
	assert.ErrorIs(t, err, rulebookAPI.ErrMissingMetadataFile)
}

func TestSetupCmd_setupAgent(t *testing.T) {
	t.Parallel()

	const currentDir = "/project"

	tests := []struct {
		name      string
		setupFs   func(*testing.T) afero.Fs
		mockSetup func(*agentAPI.MockAgent, *instructionAPI.MockRepository, *skillAPI.MockRepository)
	}{
		{
			name: "WhenAllOperationsSucceed_ThenReturnsNoError",
			setupFs: func(t *testing.T) afero.Fs {
				t.Helper()
				fs := afero.NewMemMapFs()
				require.NoError(t, fs.Mkdir(currentDir, 0755))
				require.NoError(t, fs.Mkdir(currentDir+"/.git", 0755))
				return fs
			},
			mockSetup: func(mockAgent *agentAPI.MockAgent, mockInstructionRepo *instructionAPI.MockRepository, mockSkillRepo *skillAPI.MockRepository) {
				instructions := []instructionAPI.Instructions{
					{Category: "test-category", Rules: []instructionAPI.Rule{"Rule 1"}},
				}

				mockAgent.EXPECT().GetKind().Return(agentAPI.Kind("test-agent")).Maybe()
				mockInstructionRepo.EXPECT().GetAll().Return(instructions, nil)
				mockAgent.EXPECT().RenderInstructions(instructions).Return(nil)
				mockAgent.EXPECT().RebuildSkills(mockSkillRepo).Return(nil)
				mockAgent.EXPECT().GitIgnorePatterns().Return([]string{})
			},
		},
		{
			name: "WhenAgentHasGitPatternsAndInstructions_ThenSetsUpEverything",
			setupFs: func(t *testing.T) afero.Fs {
				t.Helper()
				fs := afero.NewMemMapFs()
				require.NoError(t, fs.Mkdir(currentDir, 0755))
				require.NoError(t, fs.Mkdir(currentDir+"/.git", 0755))
				return fs
			},
			mockSetup: func(mockAgent *agentAPI.MockAgent, mockInstructionRepo *instructionAPI.MockRepository, mockSkillRepo *skillAPI.MockRepository) {
				instructions := []instructionAPI.Instructions{
					{Category: "category-one", Rules: []instructionAPI.Rule{"Rule 1", "Rule 2"}},
					{Category: "category-two", Rules: []instructionAPI.Rule{"Rule 3"}},
				}

				mockAgent.EXPECT().GetKind().Return(agentAPI.Kind("claude")).Maybe()
				mockInstructionRepo.EXPECT().GetAll().Return(instructions, nil)
				mockAgent.EXPECT().RenderInstructions(instructions).Return(nil)
				mockAgent.EXPECT().RebuildSkills(mockSkillRepo).Return(nil)
				mockAgent.EXPECT().GitIgnorePatterns().Return([]string{".claude", ".claude/**"})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			projectFs := tt.setupFs(t)
			mockAgent := agentAPI.NewMockAgent(t)
			mockInstructionRepo := instructionAPI.NewMockRepository(t)
			mockSkillRepo := skillAPI.NewMockRepository(t)
			tt.mockSetup(mockAgent, mockInstructionRepo, mockSkillRepo)

			cmd := &SetupCmd{}
			err := cmd.setupAgent(projectFs, currentDir, mockAgent, mockSkillRepo, mockInstructionRepo)

			require.NoError(t, err)
		})
	}
}

func TestSetupCmd_setupAgent_WhenGetAllInstructionsFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	const currentDir = "/project"

	projectFs := afero.NewMemMapFs()
	require.NoError(t, projectFs.Mkdir(currentDir, 0755))
	require.NoError(t, projectFs.Mkdir(currentDir+"/.git", 0755))

	mockAgent := agentAPI.NewMockAgent(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)

	getAllErr := errors.New("get all instructions failed")

	mockAgent.EXPECT().GetKind().Return(agentAPI.Kind("test-agent")).Maybe()
	mockInstructionRepo.EXPECT().GetAll().Return([]instructionAPI.Instructions{}, getAllErr)

	cmd := &SetupCmd{}
	err := cmd.setupAgent(projectFs, currentDir, mockAgent, mockSkillRepo, mockInstructionRepo)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "get all instructions")
	assert.ErrorIs(t, err, getAllErr)
}

func TestSetupCmd_setupAgent_WhenRenderInstructionsFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	const currentDir = "/project"

	projectFs := afero.NewMemMapFs()
	require.NoError(t, projectFs.Mkdir(currentDir, 0755))
	require.NoError(t, projectFs.Mkdir(currentDir+"/.git", 0755))

	mockAgent := agentAPI.NewMockAgent(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)

	instructions := []instructionAPI.Instructions{
		{Category: "test-category", Rules: []instructionAPI.Rule{"Rule 1"}},
	}
	renderErr := errors.New("render instructions failed")

	mockAgent.EXPECT().GetKind().Return(agentAPI.Kind("test-agent")).Maybe()
	mockInstructionRepo.EXPECT().GetAll().Return(instructions, nil)
	mockAgent.EXPECT().RenderInstructions(instructions).Return(renderErr)

	cmd := &SetupCmd{}
	err := cmd.setupAgent(projectFs, currentDir, mockAgent, mockSkillRepo, mockInstructionRepo)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "render instructions")
	assert.ErrorIs(t, err, renderErr)
}

func TestSetupCmd_setupAgent_WhenRebuildSkillsFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	const currentDir = "/project"

	projectFs := afero.NewMemMapFs()
	require.NoError(t, projectFs.Mkdir(currentDir, 0755))
	require.NoError(t, projectFs.Mkdir(currentDir+"/.git", 0755))

	mockAgent := agentAPI.NewMockAgent(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)

	instructions := []instructionAPI.Instructions{
		{Category: "test-category", Rules: []instructionAPI.Rule{"Rule 1"}},
	}
	rebuildErr := errors.New("rebuild skills failed")

	mockAgent.EXPECT().GetKind().Return(agentAPI.Kind("test-agent")).Maybe()
	mockInstructionRepo.EXPECT().GetAll().Return(instructions, nil)
	mockAgent.EXPECT().RenderInstructions(instructions).Return(nil)
	mockAgent.EXPECT().RebuildSkills(mockSkillRepo).Return(rebuildErr)

	cmd := &SetupCmd{}
	err := cmd.setupAgent(projectFs, currentDir, mockAgent, mockSkillRepo, mockInstructionRepo)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "rebuild skills")
	assert.ErrorIs(t, err, rebuildErr)
}

func TestSetupCmd_setupAgent_WhenGitSetupFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	const currentDir = "/project"

	projectFs := afero.NewMemMapFs()
	require.NoError(t, projectFs.Mkdir(currentDir, 0755))

	mockAgent := agentAPI.NewMockAgent(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)

	instructions := []instructionAPI.Instructions{
		{Category: "test-category", Rules: []instructionAPI.Rule{"Rule 1"}},
	}

	mockAgent.EXPECT().GetKind().Return(agentAPI.Kind("test-agent")).Maybe()
	mockInstructionRepo.EXPECT().GetAll().Return(instructions, nil)
	mockAgent.EXPECT().RenderInstructions(instructions).Return(nil)
	mockAgent.EXPECT().RebuildSkills(mockSkillRepo).Return(nil)

	cmd := &SetupCmd{}
	err := cmd.setupAgent(projectFs, currentDir, mockAgent, mockSkillRepo, mockInstructionRepo)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "git setup")
	assert.ErrorIs(t, err, git.ErrGitRepoNotFound)
}

func TestSetupCmd_execute(t *testing.T) {
	t.Parallel()

	const currentDir = "/project"

	tests := []struct {
		name      string
		setupFs   func(*testing.T) afero.Fs
		mockSetup func(*projectAPI.MockConfigLoader, *sourceAPI.MockResolver, *instructionAPI.MockRepository, *skillAPI.MockRepository, *agentAPI.MockRegistry)
	}{
		{
			name: "WhenNoAgentsConfigured_ThenReturnsNil",
			setupFs: func(t *testing.T) afero.Fs {
				t.Helper()
				fs := afero.NewMemMapFs()
				require.NoError(t, fs.Mkdir(currentDir, 0755))
				require.NoError(t, fs.Mkdir(currentDir+"/.git", 0755))
				return fs
			},
			mockSetup: func(mockConfigLoader *projectAPI.MockConfigLoader, mockResolver *sourceAPI.MockResolver, mockInstructionRepo *instructionAPI.MockRepository, mockSkillRepo *skillAPI.MockRepository, mockRegistry *agentAPI.MockRegistry) {
				config := &projectAPI.Config{
					Agents: []agentAPI.Config{},
				}
				mockConfigLoader.EXPECT().Load().Return(config, nil)
			},
		},
		{
			name: "WhenMinimalConfigWithOneAgent_ThenSetsUpAgent",
			setupFs: func(t *testing.T) afero.Fs {
				t.Helper()
				fs := afero.NewMemMapFs()
				require.NoError(t, fs.Mkdir(currentDir, 0755))
				require.NoError(t, fs.Mkdir(currentDir+"/.git", 0755))
				return fs
			},
			mockSetup: func(mockConfigLoader *projectAPI.MockConfigLoader, mockResolver *sourceAPI.MockResolver, mockInstructionRepo *instructionAPI.MockRepository, mockSkillRepo *skillAPI.MockRepository, mockRegistry *agentAPI.MockRegistry) {
				config := &projectAPI.Config{
					Agents: []agentAPI.Config{
						{Kind: "test-agent"},
					},
				}
				mockConfigLoader.EXPECT().Load().Return(config, nil)

				mockProvider := agentAPI.NewMockProvider(t)
				mockAgent := agentAPI.NewMockAgent(t)
				mockRegistry.EXPECT().GetByKind(agentAPI.Kind("test-agent")).Return(mockProvider, nil)
				mockProvider.EXPECT().NewAgent(nil).Return(mockAgent, nil)
				mockAgent.EXPECT().GetKind().Return(agentAPI.Kind("test-agent")).Maybe()

				mockInstructionRepo.EXPECT().GetAll().Return([]instructionAPI.Instructions{}, nil)
				mockAgent.EXPECT().RenderInstructions([]instructionAPI.Instructions{}).Return(nil)
				mockAgent.EXPECT().RebuildSkills(mockSkillRepo).Return(nil)
				mockAgent.EXPECT().GitIgnorePatterns().Return([]string{})
			},
		},
		{
			name: "WhenConfigWithInstructionsAndSkills_ThenAddsToReposAndSetsUpAgent",
			setupFs: func(t *testing.T) afero.Fs {
				t.Helper()
				fs := afero.NewMemMapFs()
				require.NoError(t, fs.Mkdir(currentDir, 0755))
				require.NoError(t, fs.Mkdir(currentDir+"/.git", 0755))
				return fs
			},
			mockSetup: func(mockConfigLoader *projectAPI.MockConfigLoader, mockResolver *sourceAPI.MockResolver, mockInstructionRepo *instructionAPI.MockRepository, mockSkillRepo *skillAPI.MockRepository, mockRegistry *agentAPI.MockRegistry) {
				instructionConfig := instructionAPI.Config{
					Sources: []instructionAPI.SourceConfig{{URI: "file://./instructions"}},
				}
				skillConfig := skillAPI.Config{
					Sources: []skillAPI.SourceConfig{{URI: "file://./skills"}},
				}
				config := &projectAPI.Config{
					AI: &aiAPI.Config{
						Instruction: &instructionConfig,
						Skill:       &skillConfig,
					},
					Agents: []agentAPI.Config{
						{Kind: "test-agent"},
					},
				}
				mockConfigLoader.EXPECT().Load().Return(config, nil)

				instructionFs := validInstructionFs(t, "test-category", []string{"Rule 1"})
				mockResolver.EXPECT().Resolve("file://./instructions").Return(instructionFs, nil)

				skillFs := validSkillFs(t, "test-skill", "Test skill", "Test instructions")
				mockResolver.EXPECT().Resolve("file://./skills").Return(skillFs, nil)

				mockInstructionRepo.EXPECT().AddInstructions(instructionAPI.Instructions{
					Category: "test-category",
					Rules:    []instructionAPI.Rule{"Rule 1"},
				}).Return(nil)

				mockSkillRepo.EXPECT().AddSkill(skillAPI.Skill{
					Metadata: skillAPI.Metadata{
						Name:        "test-skill",
						Description: "Test skill",
					},
					Instructions: "Test instructions",
					Scripts:      map[skillAPI.ScriptName]skillAPI.Script{},
				}).Return(nil)

				mockProvider := agentAPI.NewMockProvider(t)
				mockAgent := agentAPI.NewMockAgent(t)
				mockRegistry.EXPECT().GetByKind(agentAPI.Kind("test-agent")).Return(mockProvider, nil)
				mockProvider.EXPECT().NewAgent(nil).Return(mockAgent, nil)
				mockAgent.EXPECT().GetKind().Return(agentAPI.Kind("test-agent")).Maybe()

				instructions := []instructionAPI.Instructions{
					{Category: "test-category", Rules: []instructionAPI.Rule{"Rule 1"}},
				}
				mockInstructionRepo.EXPECT().GetAll().Return(instructions, nil)
				mockAgent.EXPECT().RenderInstructions(instructions).Return(nil)
				mockAgent.EXPECT().RebuildSkills(mockSkillRepo).Return(nil)
				mockAgent.EXPECT().GitIgnorePatterns().Return([]string{})
			},
		},
		{
			name: "WhenConfigWithRulebooks_ThenMergesRulebookInstructionsAndSkills",
			setupFs: func(t *testing.T) afero.Fs {
				t.Helper()
				fs := afero.NewMemMapFs()
				require.NoError(t, fs.Mkdir(currentDir, 0755))
				require.NoError(t, fs.Mkdir(currentDir+"/.git", 0755))
				return fs
			},
			mockSetup: func(mockConfigLoader *projectAPI.MockConfigLoader, mockResolver *sourceAPI.MockResolver, mockInstructionRepo *instructionAPI.MockRepository, mockSkillRepo *skillAPI.MockRepository, mockRegistry *agentAPI.MockRegistry) {
				rulebookConfig := rulebookAPI.Config{
					Sources: []rulebookAPI.SourceConfig{{URI: "file://./rulebooks"}},
				}
				config := &projectAPI.Config{
					Rulebook: &rulebookConfig,
					Agents: []agentAPI.Config{
						{Kind: "test-agent"},
					},
				}
				mockConfigLoader.EXPECT().Load().Return(config, nil)

				rulebookFs := validRulebookFs(t, "rulebook-category", []string{"Rulebook rule"}, "rulebook-skill", "Rulebook skill", "Rulebook instructions")
				mockResolver.EXPECT().Resolve("file://./rulebooks").Return(rulebookFs, nil)

				mockInstructionRepo.EXPECT().AddInstructions(instructionAPI.Instructions{
					Category: "rulebook-category",
					Rules:    []instructionAPI.Rule{"Rulebook rule"},
				}).Return(nil)

				mockSkillRepo.EXPECT().AddSkill(skillAPI.Skill{
					Metadata: skillAPI.Metadata{
						Name:        "rulebook-skill",
						Description: "Rulebook skill",
					},
					Instructions: "Rulebook instructions",
					Scripts:      map[skillAPI.ScriptName]skillAPI.Script{},
				}).Return(nil)

				mockProvider := agentAPI.NewMockProvider(t)
				mockAgent := agentAPI.NewMockAgent(t)
				mockRegistry.EXPECT().GetByKind(agentAPI.Kind("test-agent")).Return(mockProvider, nil)
				mockProvider.EXPECT().NewAgent(nil).Return(mockAgent, nil)
				mockAgent.EXPECT().GetKind().Return(agentAPI.Kind("test-agent")).Maybe()

				instructions := []instructionAPI.Instructions{
					{Category: "rulebook-category", Rules: []instructionAPI.Rule{"Rulebook rule"}},
				}
				mockInstructionRepo.EXPECT().GetAll().Return(instructions, nil)
				mockAgent.EXPECT().RenderInstructions(instructions).Return(nil)
				mockAgent.EXPECT().RebuildSkills(mockSkillRepo).Return(nil)
				mockAgent.EXPECT().GitIgnorePatterns().Return([]string{})
			},
		},
		{
			name: "WhenMultipleAgents_ThenSetsUpAllAgents",
			setupFs: func(t *testing.T) afero.Fs {
				t.Helper()
				fs := afero.NewMemMapFs()
				require.NoError(t, fs.Mkdir(currentDir, 0755))
				require.NoError(t, fs.Mkdir(currentDir+"/.git", 0755))
				return fs
			},
			mockSetup: func(mockConfigLoader *projectAPI.MockConfigLoader, mockResolver *sourceAPI.MockResolver, mockInstructionRepo *instructionAPI.MockRepository, mockSkillRepo *skillAPI.MockRepository, mockRegistry *agentAPI.MockRegistry) {
				config := &projectAPI.Config{
					Agents: []agentAPI.Config{
						{Kind: "agent-one"},
						{Kind: "agent-two"},
					},
				}
				mockConfigLoader.EXPECT().Load().Return(config, nil)

				mockProvider1 := agentAPI.NewMockProvider(t)
				mockAgent1 := agentAPI.NewMockAgent(t)
				mockRegistry.EXPECT().GetByKind(agentAPI.Kind("agent-one")).Return(mockProvider1, nil)
				mockProvider1.EXPECT().NewAgent(nil).Return(mockAgent1, nil)
				mockAgent1.EXPECT().GetKind().Return(agentAPI.Kind("agent-one")).Maybe()
				mockInstructionRepo.EXPECT().GetAll().Return([]instructionAPI.Instructions{}, nil)
				mockAgent1.EXPECT().RenderInstructions([]instructionAPI.Instructions{}).Return(nil)
				mockAgent1.EXPECT().RebuildSkills(mockSkillRepo).Return(nil)
				mockAgent1.EXPECT().GitIgnorePatterns().Return([]string{})

				mockProvider2 := agentAPI.NewMockProvider(t)
				mockAgent2 := agentAPI.NewMockAgent(t)
				mockRegistry.EXPECT().GetByKind(agentAPI.Kind("agent-two")).Return(mockProvider2, nil)
				mockProvider2.EXPECT().NewAgent(nil).Return(mockAgent2, nil)
				mockAgent2.EXPECT().GetKind().Return(agentAPI.Kind("agent-two")).Maybe()
				mockInstructionRepo.EXPECT().GetAll().Return([]instructionAPI.Instructions{}, nil)
				mockAgent2.EXPECT().RenderInstructions([]instructionAPI.Instructions{}).Return(nil)
				mockAgent2.EXPECT().RebuildSkills(mockSkillRepo).Return(nil)
				mockAgent2.EXPECT().GitIgnorePatterns().Return([]string{})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			projectFs := tt.setupFs(t)
			mockConfigLoader := projectAPI.NewMockConfigLoader(t)
			mockResolver := sourceAPI.NewMockResolver(t)
			mockInstructionRepo := instructionAPI.NewMockRepository(t)
			mockSkillRepo := skillAPI.NewMockRepository(t)
			mockRegistry := agentAPI.NewMockRegistry(t)

			tt.mockSetup(mockConfigLoader, mockResolver, mockInstructionRepo, mockSkillRepo, mockRegistry)

			cmd := &SetupCmd{}
			err := cmd.execute(projectFs, currentDir, mockConfigLoader, mockResolver, mockInstructionRepo, mockSkillRepo, mockRegistry)

			require.NoError(t, err)
		})
	}
}

func TestSetupCmd_execute_WhenConfigLoaderFails_ThenReturnsConfigLoaderError(t *testing.T) {
	t.Parallel()

	const currentDir = "/project"

	projectFs := afero.NewMemMapFs()
	require.NoError(t, projectFs.Mkdir(currentDir, 0755))

	mockConfigLoader := projectAPI.NewMockConfigLoader(t)
	mockResolver := sourceAPI.NewMockResolver(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockRegistry := agentAPI.NewMockRegistry(t)

	loadErr := errors.New("config loader failed")
	mockConfigLoader.EXPECT().Load().Return(nil, loadErr)

	cmd := &SetupCmd{}
	err := cmd.execute(projectFs, currentDir, mockConfigLoader, mockResolver, mockInstructionRepo, mockSkillRepo, mockRegistry)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "config loader")
	assert.ErrorIs(t, err, loadErr)
}

func TestSetupCmd_execute_WhenLoadInstructionsFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	const currentDir = "/project"

	projectFs := afero.NewMemMapFs()
	require.NoError(t, projectFs.Mkdir(currentDir, 0755))

	mockConfigLoader := projectAPI.NewMockConfigLoader(t)
	mockResolver := sourceAPI.NewMockResolver(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockRegistry := agentAPI.NewMockRegistry(t)

	instructionConfig := instructionAPI.Config{
		Sources: []instructionAPI.SourceConfig{{URI: "file://./instructions"}},
	}
	config := &projectAPI.Config{
		AI: &aiAPI.Config{
			Instruction: &instructionConfig,
		},
		Agents: []agentAPI.Config{
			{Kind: "test-agent"},
		},
	}
	mockConfigLoader.EXPECT().Load().Return(config, nil)

	resolveErr := errors.New("resolve instructions failed")
	mockResolver.EXPECT().Resolve("file://./instructions").Return(nil, resolveErr)

	cmd := &SetupCmd{}
	err := cmd.execute(projectFs, currentDir, mockConfigLoader, mockResolver, mockInstructionRepo, mockSkillRepo, mockRegistry)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "load instructions from config")
	assert.ErrorIs(t, err, resolveErr)
}

func TestSetupCmd_execute_WhenLoadSkillsFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	const currentDir = "/project"

	projectFs := afero.NewMemMapFs()
	require.NoError(t, projectFs.Mkdir(currentDir, 0755))

	mockConfigLoader := projectAPI.NewMockConfigLoader(t)
	mockResolver := sourceAPI.NewMockResolver(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockRegistry := agentAPI.NewMockRegistry(t)

	skillConfig := skillAPI.Config{
		Sources: []skillAPI.SourceConfig{{URI: "file://./skills"}},
	}
	config := &projectAPI.Config{
		AI: &aiAPI.Config{
			Skill: &skillConfig,
		},
		Agents: []agentAPI.Config{
			{Kind: "test-agent"},
		},
	}
	mockConfigLoader.EXPECT().Load().Return(config, nil)

	resolveErr := errors.New("resolve skills failed")
	mockResolver.EXPECT().Resolve("file://./skills").Return(nil, resolveErr)

	cmd := &SetupCmd{}
	err := cmd.execute(projectFs, currentDir, mockConfigLoader, mockResolver, mockInstructionRepo, mockSkillRepo, mockRegistry)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "load skills from config")
	assert.ErrorIs(t, err, resolveErr)
}

func TestSetupCmd_execute_WhenLoadRulebooksFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	const currentDir = "/project"

	projectFs := afero.NewMemMapFs()
	require.NoError(t, projectFs.Mkdir(currentDir, 0755))

	mockConfigLoader := projectAPI.NewMockConfigLoader(t)
	mockResolver := sourceAPI.NewMockResolver(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockRegistry := agentAPI.NewMockRegistry(t)

	rulebookConfig := rulebookAPI.Config{
		Sources: []rulebookAPI.SourceConfig{{URI: "file://./rulebooks"}},
	}
	config := &projectAPI.Config{
		Rulebook: &rulebookConfig,
		Agents: []agentAPI.Config{
			{Kind: "test-agent"},
		},
	}
	mockConfigLoader.EXPECT().Load().Return(config, nil)

	resolveErr := errors.New("resolve rulebooks failed")
	mockResolver.EXPECT().Resolve("file://./rulebooks").Return(nil, resolveErr)

	cmd := &SetupCmd{}
	err := cmd.execute(projectFs, currentDir, mockConfigLoader, mockResolver, mockInstructionRepo, mockSkillRepo, mockRegistry)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "load rule books from config")
	assert.ErrorIs(t, err, resolveErr)
}

func TestSetupCmd_execute_WhenAddInstructionsFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	const currentDir = "/project"

	projectFs := afero.NewMemMapFs()
	require.NoError(t, projectFs.Mkdir(currentDir, 0755))

	mockConfigLoader := projectAPI.NewMockConfigLoader(t)
	mockResolver := sourceAPI.NewMockResolver(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockRegistry := agentAPI.NewMockRegistry(t)

	instructionConfig := instructionAPI.Config{
		Sources: []instructionAPI.SourceConfig{{URI: "file://./instructions"}},
	}
	config := &projectAPI.Config{
		AI: &aiAPI.Config{
			Instruction: &instructionConfig,
		},
		Agents: []agentAPI.Config{
			{Kind: "test-agent"},
		},
	}
	mockConfigLoader.EXPECT().Load().Return(config, nil)

	instructionFs := validInstructionFs(t, "test-category", []string{"Rule 1"})
	mockResolver.EXPECT().Resolve("file://./instructions").Return(instructionFs, nil)

	addErr := errors.New("add instructions failed")
	mockInstructionRepo.EXPECT().AddInstructions(instructionAPI.Instructions{
		Category: "test-category",
		Rules:    []instructionAPI.Rule{"Rule 1"},
	}).Return(addErr)

	cmd := &SetupCmd{}
	err := cmd.execute(projectFs, currentDir, mockConfigLoader, mockResolver, mockInstructionRepo, mockSkillRepo, mockRegistry)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "add instructions to repository")
	assert.ErrorIs(t, err, addErr)
}

func TestSetupCmd_execute_WhenAddSkillFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	const currentDir = "/project"

	projectFs := afero.NewMemMapFs()
	require.NoError(t, projectFs.Mkdir(currentDir, 0755))

	mockConfigLoader := projectAPI.NewMockConfigLoader(t)
	mockResolver := sourceAPI.NewMockResolver(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockRegistry := agentAPI.NewMockRegistry(t)

	skillConfig := skillAPI.Config{
		Sources: []skillAPI.SourceConfig{{URI: "file://./skills"}},
	}
	config := &projectAPI.Config{
		AI: &aiAPI.Config{
			Skill: &skillConfig,
		},
		Agents: []agentAPI.Config{
			{Kind: "test-agent"},
		},
	}
	mockConfigLoader.EXPECT().Load().Return(config, nil)

	skillFs := validSkillFs(t, "test-skill", "Test skill", "Test instructions")
	mockResolver.EXPECT().Resolve("file://./skills").Return(skillFs, nil)

	addErr := errors.New("add skill failed")
	mockSkillRepo.EXPECT().AddSkill(skillAPI.Skill{
		Metadata: skillAPI.Metadata{
			Name:        "test-skill",
			Description: "Test skill",
		},
		Instructions: "Test instructions",
		Scripts:      map[skillAPI.ScriptName]skillAPI.Script{},
	}).Return(addErr)

	cmd := &SetupCmd{}
	err := cmd.execute(projectFs, currentDir, mockConfigLoader, mockResolver, mockInstructionRepo, mockSkillRepo, mockRegistry)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "add skill")
	assert.ErrorIs(t, err, addErr)
}

func TestSetupCmd_execute_WhenLoadAgentsFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	const currentDir = "/project"

	projectFs := afero.NewMemMapFs()
	require.NoError(t, projectFs.Mkdir(currentDir, 0755))

	mockConfigLoader := projectAPI.NewMockConfigLoader(t)
	mockResolver := sourceAPI.NewMockResolver(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockRegistry := agentAPI.NewMockRegistry(t)

	config := &projectAPI.Config{
		Agents: []agentAPI.Config{
			{Kind: "test-agent"},
		},
	}
	mockConfigLoader.EXPECT().Load().Return(config, nil)

	getByKindErr := errors.New("get by kind failed")
	mockRegistry.EXPECT().GetByKind(agentAPI.Kind("test-agent")).Return(nil, getByKindErr)

	cmd := &SetupCmd{}
	err := cmd.execute(projectFs, currentDir, mockConfigLoader, mockResolver, mockInstructionRepo, mockSkillRepo, mockRegistry)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "load agents")
	assert.ErrorIs(t, err, getByKindErr)
}

func TestSetupCmd_execute_WhenSetupAgentFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	const currentDir = "/project"

	projectFs := afero.NewMemMapFs()
	require.NoError(t, projectFs.Mkdir(currentDir, 0755))

	mockConfigLoader := projectAPI.NewMockConfigLoader(t)
	mockResolver := sourceAPI.NewMockResolver(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockRegistry := agentAPI.NewMockRegistry(t)

	config := &projectAPI.Config{
		Agents: []agentAPI.Config{
			{Kind: "test-agent"},
		},
	}
	mockConfigLoader.EXPECT().Load().Return(config, nil)

	mockProvider := agentAPI.NewMockProvider(t)
	mockAgent := agentAPI.NewMockAgent(t)
	mockRegistry.EXPECT().GetByKind(agentAPI.Kind("test-agent")).Return(mockProvider, nil)
	mockProvider.EXPECT().NewAgent(nil).Return(mockAgent, nil)
	mockAgent.EXPECT().GetKind().Return(agentAPI.Kind("test-agent")).Maybe()

	mockInstructionRepo.EXPECT().GetAll().Return([]instructionAPI.Instructions{}, nil)
	mockAgent.EXPECT().RenderInstructions([]instructionAPI.Instructions{}).Return(nil)
	mockAgent.EXPECT().RebuildSkills(mockSkillRepo).Return(nil)

	cmd := &SetupCmd{}
	err := cmd.execute(projectFs, currentDir, mockConfigLoader, mockResolver, mockInstructionRepo, mockSkillRepo, mockRegistry)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "setup agent")
	assert.ErrorIs(t, err, git.ErrGitRepoNotFound)
}
