package projectkit

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	agentAPI "github.com/orbiqd/orbiqd-projectkit/pkg/agent"
	instructionAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/instruction"
	mcpAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/mcp"
	skillAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/skill"
	workflowAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/workflow"
	docAPI "github.com/orbiqd/orbiqd-projectkit/pkg/doc"
	standardAPI "github.com/orbiqd/orbiqd-projectkit/pkg/doc/standard"
	projectAPI "github.com/orbiqd/orbiqd-projectkit/pkg/project"
	sourceAPI "github.com/orbiqd/orbiqd-projectkit/pkg/source"
)

func TestUpdateCmdRun_WhenEmptyConfig_ThenReturnsNoError(t *testing.T) {
	t.Setenv("BRIEFKIT_BINARY_PATH", "/test/bin/projectkit")

	mockResolver := sourceAPI.NewMockResolver(t)
	mockInstRepo := instructionAPI.NewMockRepository(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockWorkflowRepo := workflowAPI.NewMockRepository(t)
	mockMcpRepo := mcpAPI.NewMockRepository(t)
	mockStandardRepo := standardAPI.NewMockRepository(t)

	mockStandardRepo.EXPECT().RemoveAll().Return(nil)
	mockInstRepo.EXPECT().RemoveAll().Return(nil)
	mockSkillRepo.EXPECT().RemoveAll().Return(nil)
	mockWorkflowRepo.EXPECT().RemoveAllWorkflows().Return(nil)
	mockMcpRepo.EXPECT().RemoveAll().Return(nil)
	mockMcpRepo.EXPECT().AddMCPServer(mock.AnythingOfType("mcp.MCPServer")).Return(nil)

	config := &projectAPI.Config{}
	cmd := UpdateCmd{}

	err := cmd.Run(config, mockResolver, mockInstRepo, mockSkillRepo, mockWorkflowRepo, mockMcpRepo, mockStandardRepo)

	require.NoError(t, err)
}

func TestAgentRenderCmdRun_WhenNoAgents_ThenReturnsNoError(t *testing.T) {
	mockRegistry := agentAPI.NewMockRegistry(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockInstRepo := instructionAPI.NewMockRepository(t)
	mockMcpRepo := mcpAPI.NewMockRepository(t)
	gitFs := afero.NewMemMapFs()

	config := &projectAPI.Config{}
	cmd := AgentRenderCmd{}

	err := cmd.Run(gitFs, config, mockInstRepo, mockSkillRepo, mockRegistry, mockMcpRepo)

	require.NoError(t, err)
}

func TestDocStandardRenderCmdRun_WhenNoStandards_ThenReturnsNoError(t *testing.T) {
	mockStandardRepo := standardAPI.NewMockRepository(t)
	projectFs := afero.NewMemMapFs()

	mockStandardRepo.EXPECT().GetAll().Return([]standardAPI.Standard{}, nil)

	config := &projectAPI.Config{
		Docs: &docAPI.Config{
			Standard: &standardAPI.Config{
				Render: []standardAPI.RenderConfig{},
			},
		},
	}
	cmd := DocStandardRenderCmd{}

	err := cmd.Run(config, projectFs, mockStandardRepo)

	require.NoError(t, err)
}

func TestRenderCmdRun_WhenEmptyConfig_ThenReturnsNoError(t *testing.T) {
	mockRegistry := agentAPI.NewMockRegistry(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockInstRepo := instructionAPI.NewMockRepository(t)
	mockMcpRepo := mcpAPI.NewMockRepository(t)
	mockStandardRepo := standardAPI.NewMockRepository(t)
	gitFs := afero.NewMemMapFs()
	projectFs := afero.NewMemMapFs()

	mockStandardRepo.EXPECT().GetAll().Return([]standardAPI.Standard{}, nil)

	config := &projectAPI.Config{
		Docs: &docAPI.Config{
			Standard: &standardAPI.Config{
				Render: []standardAPI.RenderConfig{},
			},
		},
	}
	cmd := RenderCmd{}

	err := cmd.Run(gitFs, config, mockInstRepo, mockSkillRepo, mockRegistry, projectFs, mockStandardRepo, mockMcpRepo)

	require.NoError(t, err)
}
