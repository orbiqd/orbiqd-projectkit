package projectkit

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"

	agentAPI "github.com/orbiqd/orbiqd-projectkit/pkg/agent"
	instructionAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/instruction"
	skillAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/skill"
	workflowAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/workflow"
	docAPI "github.com/orbiqd/orbiqd-projectkit/pkg/doc"
	"github.com/orbiqd/orbiqd-projectkit/pkg/doc/standard"
	projectAPI "github.com/orbiqd/orbiqd-projectkit/pkg/project"
	sourceAPI "github.com/orbiqd/orbiqd-projectkit/pkg/source"
)

func TestUpdateCmdRun_WhenEmptyConfig_ThenReturnsNoError(t *testing.T) {
	mockResolver := sourceAPI.NewMockResolver(t)
	mockInstRepo := instructionAPI.NewMockRepository(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockWorkflowRepo := workflowAPI.NewMockRepository(t)
	mockStandardRepo := standard.NewMockRepository(t)

	mockInstRepo.EXPECT().RemoveAll().Return(nil)
	mockSkillRepo.EXPECT().RemoveAll().Return(nil)
	mockWorkflowRepo.EXPECT().RemoveAllWorkflows().Return(nil)
	mockStandardRepo.EXPECT().RemoveAll().Return(nil)

	config := &projectAPI.Config{}
	cmd := UpdateCmd{}

	err := cmd.Run(config, mockResolver, mockInstRepo, mockSkillRepo, mockWorkflowRepo, mockStandardRepo)

	require.NoError(t, err)
}

func TestAgentRenderCmdRun_WhenNoAgents_ThenReturnsNoError(t *testing.T) {
	mockRegistry := agentAPI.NewMockRegistry(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockInstRepo := instructionAPI.NewMockRepository(t)
	gitFs := afero.NewMemMapFs()

	config := &projectAPI.Config{}
	cmd := AgentRenderCmd{}

	err := cmd.Run(gitFs, config, mockInstRepo, mockSkillRepo, mockRegistry)

	require.NoError(t, err)
}

func TestDocStandardRenderCmdRun_WhenNoStandards_ThenReturnsNoError(t *testing.T) {
	mockStandardRepo := standard.NewMockRepository(t)
	projectFs := afero.NewMemMapFs()

	mockStandardRepo.EXPECT().GetAll().Return([]standard.Standard{}, nil)

	config := &projectAPI.Config{
		Docs: &docAPI.Config{
			Standard: &standard.Config{
				Render: []standard.RenderConfig{},
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
	mockStandardRepo := standard.NewMockRepository(t)
	gitFs := afero.NewMemMapFs()
	projectFs := afero.NewMemMapFs()

	mockStandardRepo.EXPECT().GetAll().Return([]standard.Standard{}, nil)

	config := &projectAPI.Config{
		Docs: &docAPI.Config{
			Standard: &standard.Config{
				Render: []standard.RenderConfig{},
			},
		},
	}
	cmd := RenderCmd{}

	err := cmd.Run(gitFs, config, mockInstRepo, mockSkillRepo, mockRegistry, projectFs, mockStandardRepo)

	require.NoError(t, err)
}
