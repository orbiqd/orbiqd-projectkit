package action

import (
	"errors"
	"io/fs"
	"testing"

	agentAPI "github.com/orbiqd/orbiqd-projectkit/pkg/agent"
	instructionAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/instruction"
	mcpAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/mcp"
	skillAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/skill"
	projectAPI "github.com/orbiqd/orbiqd-projectkit/pkg/project"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.nhat.io/aferomock"
)

func setupMockAgentChain(
	t *testing.T,
	registry *agentAPI.MockRegistry,
	kind string,
	patterns []string,
) (*agentAPI.MockProvider, *agentAPI.MockAgent) {
	t.Helper()

	mockProvider := agentAPI.NewMockProvider(t)
	mockAgent := agentAPI.NewMockAgent(t)

	registry.EXPECT().GetByKind(agentAPI.Kind(kind)).Return(mockProvider, nil)
	mockProvider.EXPECT().NewAgent(nil).Return(mockAgent, nil)
	mockAgent.EXPECT().GetKind().Return(agentAPI.Kind(kind))
	mockAgent.EXPECT().GitIgnorePatterns().Return(patterns)

	return mockProvider, mockAgent
}

func TestRenderAgentActionRun_WhenNoAgents_ThenReturnsNil(t *testing.T) {
	t.Parallel()

	mockRegistry := agentAPI.NewMockRegistry(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockMcpRepo := mcpAPI.NewMockRepository(t)

	gitFs := afero.NewMemMapFs()
	config := projectAPI.Config{}

	action := NewRenderAgentAction(gitFs, config, mockRegistry, mockSkillRepo, mockInstructionRepo, mockMcpRepo)

	err := action.Run()

	require.NoError(t, err)
}

func TestRenderAgentActionRun_WhenSingleAgentNoPatterns_ThenRendersSuccessfully(t *testing.T) {
	t.Parallel()

	mockRegistry := agentAPI.NewMockRegistry(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockMcpRepo := mcpAPI.NewMockRepository(t)

	_, mockAgent := setupMockAgentChain(t, mockRegistry, "test-agent", []string{})

	mockInstructionRepo.EXPECT().GetAll().Return([]instructionAPI.Instructions{}, nil)
	mockAgent.EXPECT().RenderInstructions([]instructionAPI.Instructions{}).Return(nil)
	mockAgent.EXPECT().RebuildSkills(mockSkillRepo).Return(nil)
	mockMcpRepo.EXPECT().GetAll().Return([]mcpAPI.MCPServer{}, nil)
	mockAgent.EXPECT().RenderMCPServers([]mcpAPI.MCPServer{}).Return(nil)

	gitFs := afero.NewMemMapFs()
	config := projectAPI.Config{
		Agents: []agentAPI.Config{
			{Kind: "test-agent"},
		},
	}

	action := NewRenderAgentAction(gitFs, config, mockRegistry, mockSkillRepo, mockInstructionRepo, mockMcpRepo)

	err := action.Run()

	require.NoError(t, err)
}

func TestRenderAgentActionRun_WhenAgentWithNewPattern_ThenAddsGitExclude(t *testing.T) {
	t.Parallel()

	mockRegistry := agentAPI.NewMockRegistry(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockMcpRepo := mcpAPI.NewMockRepository(t)

	_, mockAgent := setupMockAgentChain(t, mockRegistry, "test-agent", []string{".ai"})

	mockInstructionRepo.EXPECT().GetAll().Return([]instructionAPI.Instructions{}, nil)
	mockAgent.EXPECT().RenderInstructions([]instructionAPI.Instructions{}).Return(nil)
	mockAgent.EXPECT().RebuildSkills(mockSkillRepo).Return(nil)
	mockMcpRepo.EXPECT().GetAll().Return([]mcpAPI.MCPServer{}, nil)
	mockAgent.EXPECT().RenderMCPServers([]mcpAPI.MCPServer{}).Return(nil)

	gitFs := afero.NewMemMapFs()
	config := projectAPI.Config{
		Agents: []agentAPI.Config{
			{Kind: "test-agent"},
		},
	}

	action := NewRenderAgentAction(gitFs, config, mockRegistry, mockSkillRepo, mockInstructionRepo, mockMcpRepo)

	err := action.Run()

	require.NoError(t, err)

	content, err := afero.ReadFile(gitFs, "info/exclude")
	require.NoError(t, err)
	assert.Contains(t, string(content), ".ai")
}

func TestRenderAgentActionRun_WhenAgentWithExistingPattern_ThenSkipsExclude(t *testing.T) {
	t.Parallel()

	mockRegistry := agentAPI.NewMockRegistry(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockMcpRepo := mcpAPI.NewMockRepository(t)

	_, mockAgent := setupMockAgentChain(t, mockRegistry, "test-agent", []string{".ai"})

	mockInstructionRepo.EXPECT().GetAll().Return([]instructionAPI.Instructions{}, nil)
	mockAgent.EXPECT().RenderInstructions([]instructionAPI.Instructions{}).Return(nil)
	mockAgent.EXPECT().RebuildSkills(mockSkillRepo).Return(nil)
	mockMcpRepo.EXPECT().GetAll().Return([]mcpAPI.MCPServer{}, nil)
	mockAgent.EXPECT().RenderMCPServers([]mcpAPI.MCPServer{}).Return(nil)

	gitFs := afero.NewMemMapFs()
	require.NoError(t, gitFs.MkdirAll("info", 0755))
	require.NoError(t, afero.WriteFile(gitFs, "info/exclude", []byte(".ai\n"), 0644))

	config := projectAPI.Config{
		Agents: []agentAPI.Config{
			{Kind: "test-agent"},
		},
	}

	action := NewRenderAgentAction(gitFs, config, mockRegistry, mockSkillRepo, mockInstructionRepo, mockMcpRepo)

	err := action.Run()

	require.NoError(t, err)

	content, err := afero.ReadFile(gitFs, "info/exclude")
	require.NoError(t, err)
	assert.Equal(t, ".ai\n", string(content))
}

func TestRenderAgentActionRun_WhenMultipleAgents_ThenRendersAll(t *testing.T) {
	t.Parallel()

	mockRegistry := agentAPI.NewMockRegistry(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockMcpRepo := mcpAPI.NewMockRepository(t)

	_, mockAgent1 := setupMockAgentChain(t, mockRegistry, "agent-one", []string{})
	_, mockAgent2 := setupMockAgentChain(t, mockRegistry, "agent-two", []string{})

	instructions := []instructionAPI.Instructions{}
	mcpServers := []mcpAPI.MCPServer{}
	mockInstructionRepo.EXPECT().GetAll().Return(instructions, nil)
	mockAgent1.EXPECT().RenderInstructions(instructions).Return(nil)
	mockAgent1.EXPECT().RebuildSkills(mockSkillRepo).Return(nil)
	mockMcpRepo.EXPECT().GetAll().Return(mcpServers, nil)
	mockAgent1.EXPECT().RenderMCPServers(mcpServers).Return(nil)
	mockAgent2.EXPECT().RenderInstructions(instructions).Return(nil)
	mockAgent2.EXPECT().RebuildSkills(mockSkillRepo).Return(nil)
	mockMcpRepo.EXPECT().GetAll().Return(mcpServers, nil)
	mockAgent2.EXPECT().RenderMCPServers(mcpServers).Return(nil)

	gitFs := afero.NewMemMapFs()
	config := projectAPI.Config{
		Agents: []agentAPI.Config{
			{Kind: "agent-one"},
			{Kind: "agent-two"},
		},
	}

	action := NewRenderAgentAction(gitFs, config, mockRegistry, mockSkillRepo, mockInstructionRepo, mockMcpRepo)

	err := action.Run()

	require.NoError(t, err)
}

func TestRenderAgentActionRun_WhenLoadAgentsFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockRegistry := agentAPI.NewMockRegistry(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockMcpRepo := mcpAPI.NewMockRepository(t)

	loadErr := errors.New("load agents error")
	mockRegistry.EXPECT().GetByKind(agentAPI.Kind("test-agent")).Return(nil, loadErr)

	gitFs := afero.NewMemMapFs()
	config := projectAPI.Config{
		Agents: []agentAPI.Config{
			{Kind: "test-agent"},
		},
	}

	action := NewRenderAgentAction(gitFs, config, mockRegistry, mockSkillRepo, mockInstructionRepo, mockMcpRepo)

	err := action.Run()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "load agents")
	assert.ErrorIs(t, err, loadErr)
}

func TestRenderAgentActionRun_WhenGetAllInstructionsFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockRegistry := agentAPI.NewMockRegistry(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockMcpRepo := mcpAPI.NewMockRepository(t)

	mockProvider := agentAPI.NewMockProvider(t)
	mockAgent := agentAPI.NewMockAgent(t)

	mockRegistry.EXPECT().GetByKind(agentAPI.Kind("test-agent")).Return(mockProvider, nil)
	mockProvider.EXPECT().NewAgent(nil).Return(mockAgent, nil)
	mockAgent.EXPECT().GetKind().Return(agentAPI.Kind("test-agent"))

	getAllErr := errors.New("get all instructions error")
	mockInstructionRepo.EXPECT().GetAll().Return(nil, getAllErr)

	gitFs := afero.NewMemMapFs()
	config := projectAPI.Config{
		Agents: []agentAPI.Config{
			{Kind: "test-agent"},
		},
	}

	action := NewRenderAgentAction(gitFs, config, mockRegistry, mockSkillRepo, mockInstructionRepo, mockMcpRepo)

	err := action.Run()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "get all instructions")
	assert.ErrorIs(t, err, getAllErr)
}

func TestRenderAgentActionRun_WhenRenderInstructionsFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockRegistry := agentAPI.NewMockRegistry(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockMcpRepo := mcpAPI.NewMockRepository(t)

	mockProvider := agentAPI.NewMockProvider(t)
	mockAgent := agentAPI.NewMockAgent(t)

	mockRegistry.EXPECT().GetByKind(agentAPI.Kind("test-agent")).Return(mockProvider, nil)
	mockProvider.EXPECT().NewAgent(nil).Return(mockAgent, nil)
	mockAgent.EXPECT().GetKind().Return(agentAPI.Kind("test-agent"))

	renderErr := errors.New("render instructions error")
	instructions := []instructionAPI.Instructions{}
	mockInstructionRepo.EXPECT().GetAll().Return(instructions, nil)
	mockAgent.EXPECT().RenderInstructions(instructions).Return(renderErr)

	gitFs := afero.NewMemMapFs()
	config := projectAPI.Config{
		Agents: []agentAPI.Config{
			{Kind: "test-agent"},
		},
	}

	action := NewRenderAgentAction(gitFs, config, mockRegistry, mockSkillRepo, mockInstructionRepo, mockMcpRepo)

	err := action.Run()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "render instructions")
	assert.ErrorIs(t, err, renderErr)
}

func TestRenderAgentActionRun_WhenRebuildSkillsFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockRegistry := agentAPI.NewMockRegistry(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockMcpRepo := mcpAPI.NewMockRepository(t)

	mockProvider := agentAPI.NewMockProvider(t)
	mockAgent := agentAPI.NewMockAgent(t)

	mockRegistry.EXPECT().GetByKind(agentAPI.Kind("test-agent")).Return(mockProvider, nil)
	mockProvider.EXPECT().NewAgent(nil).Return(mockAgent, nil)
	mockAgent.EXPECT().GetKind().Return(agentAPI.Kind("test-agent"))

	rebuildErr := errors.New("rebuild skills error")
	instructions := []instructionAPI.Instructions{}
	mockInstructionRepo.EXPECT().GetAll().Return(instructions, nil)
	mockAgent.EXPECT().RenderInstructions(instructions).Return(nil)
	mockAgent.EXPECT().RebuildSkills(mockSkillRepo).Return(rebuildErr)

	gitFs := afero.NewMemMapFs()
	config := projectAPI.Config{
		Agents: []agentAPI.Config{
			{Kind: "test-agent"},
		},
	}

	action := NewRenderAgentAction(gitFs, config, mockRegistry, mockSkillRepo, mockInstructionRepo, mockMcpRepo)

	err := action.Run()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "rebuild skills")
	assert.ErrorIs(t, err, rebuildErr)
}

func TestRenderAgentActionRun_WhenIsExcludedFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockRegistry := agentAPI.NewMockRegistry(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockMcpRepo := mcpAPI.NewMockRepository(t)

	_, mockAgent := setupMockAgentChain(t, mockRegistry, "test-agent", []string{".ai"})

	instructions := []instructionAPI.Instructions{}
	mockInstructionRepo.EXPECT().GetAll().Return(instructions, nil)
	mockAgent.EXPECT().RenderInstructions(instructions).Return(nil)
	mockAgent.EXPECT().RebuildSkills(mockSkillRepo).Return(nil)
	mockMcpRepo.EXPECT().GetAll().Return([]mcpAPI.MCPServer{}, nil)
	mockAgent.EXPECT().RenderMCPServers([]mcpAPI.MCPServer{}).Return(nil)

	statErr := errors.New("stat error")
	baseFs := afero.NewMemMapFs()
	gitFs := aferomock.OverrideFs(baseFs, aferomock.FsCallbacks{
		StatFunc: func(name string) (fs.FileInfo, error) {
			return nil, statErr
		},
	})

	config := projectAPI.Config{
		Agents: []agentAPI.Config{
			{Kind: "test-agent"},
		},
	}

	action := NewRenderAgentAction(gitFs, config, mockRegistry, mockSkillRepo, mockInstructionRepo, mockMcpRepo)

	err := action.Run()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "check excluded pattern")
}

func TestRenderAgentActionRun_WhenExcludeFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockRegistry := agentAPI.NewMockRegistry(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockMcpRepo := mcpAPI.NewMockRepository(t)

	_, mockAgent := setupMockAgentChain(t, mockRegistry, "test-agent", []string{".ai"})

	instructions := []instructionAPI.Instructions{}
	mockInstructionRepo.EXPECT().GetAll().Return(instructions, nil)
	mockAgent.EXPECT().RenderInstructions(instructions).Return(nil)
	mockAgent.EXPECT().RebuildSkills(mockSkillRepo).Return(nil)
	mockMcpRepo.EXPECT().GetAll().Return([]mcpAPI.MCPServer{}, nil)
	mockAgent.EXPECT().RenderMCPServers([]mcpAPI.MCPServer{}).Return(nil)

	mkdirErr := errors.New("mkdir error")
	baseFs := afero.NewMemMapFs()
	gitFs := aferomock.OverrideFs(baseFs, aferomock.FsCallbacks{
		MkdirAllFunc: func(path string, perm fs.FileMode) error {
			return mkdirErr
		},
	})

	config := projectAPI.Config{
		Agents: []agentAPI.Config{
			{Kind: "test-agent"},
		},
	}

	action := NewRenderAgentAction(gitFs, config, mockRegistry, mockSkillRepo, mockInstructionRepo, mockMcpRepo)

	err := action.Run()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "exclude pattern")
}

func TestRenderAgentActionRun_WhenGetAllMCPServersFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockRegistry := agentAPI.NewMockRegistry(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockMcpRepo := mcpAPI.NewMockRepository(t)

	mockProvider := agentAPI.NewMockProvider(t)
	mockAgent := agentAPI.NewMockAgent(t)

	mockRegistry.EXPECT().GetByKind(agentAPI.Kind("test-agent")).Return(mockProvider, nil)
	mockProvider.EXPECT().NewAgent(nil).Return(mockAgent, nil)
	mockAgent.EXPECT().GetKind().Return(agentAPI.Kind("test-agent"))

	getAllErr := errors.New("get all mcp servers error")
	instructions := []instructionAPI.Instructions{}
	mockInstructionRepo.EXPECT().GetAll().Return(instructions, nil)
	mockAgent.EXPECT().RenderInstructions(instructions).Return(nil)
	mockAgent.EXPECT().RebuildSkills(mockSkillRepo).Return(nil)
	mockMcpRepo.EXPECT().GetAll().Return(nil, getAllErr)

	gitFs := afero.NewMemMapFs()
	config := projectAPI.Config{
		Agents: []agentAPI.Config{
			{Kind: "test-agent"},
		},
	}

	action := NewRenderAgentAction(gitFs, config, mockRegistry, mockSkillRepo, mockInstructionRepo, mockMcpRepo)

	err := action.Run()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "get all mcp servers")
	assert.ErrorIs(t, err, getAllErr)
}

func TestRenderAgentActionRun_WhenRenderMCPServersFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockRegistry := agentAPI.NewMockRegistry(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockMcpRepo := mcpAPI.NewMockRepository(t)

	mockProvider := agentAPI.NewMockProvider(t)
	mockAgent := agentAPI.NewMockAgent(t)

	mockRegistry.EXPECT().GetByKind(agentAPI.Kind("test-agent")).Return(mockProvider, nil)
	mockProvider.EXPECT().NewAgent(nil).Return(mockAgent, nil)
	mockAgent.EXPECT().GetKind().Return(agentAPI.Kind("test-agent"))

	renderErr := errors.New("render mcp servers error")
	instructions := []instructionAPI.Instructions{}
	mcpServers := []mcpAPI.MCPServer{}
	mockInstructionRepo.EXPECT().GetAll().Return(instructions, nil)
	mockAgent.EXPECT().RenderInstructions(instructions).Return(nil)
	mockAgent.EXPECT().RebuildSkills(mockSkillRepo).Return(nil)
	mockMcpRepo.EXPECT().GetAll().Return(mcpServers, nil)
	mockAgent.EXPECT().RenderMCPServers(mcpServers).Return(renderErr)

	gitFs := afero.NewMemMapFs()
	config := projectAPI.Config{
		Agents: []agentAPI.Config{
			{Kind: "test-agent"},
		},
	}

	action := NewRenderAgentAction(gitFs, config, mockRegistry, mockSkillRepo, mockInstructionRepo, mockMcpRepo)

	err := action.Run()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "render mcp servers")
	assert.ErrorIs(t, err, renderErr)
}
