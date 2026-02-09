package projectkit

import (
	"errors"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	agentAPI "github.com/orbiqd/orbiqd-projectkit/pkg/agent"
	instructionAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/instruction"
	skillAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/skill"
	projectAPI "github.com/orbiqd/orbiqd-projectkit/pkg/project"
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
