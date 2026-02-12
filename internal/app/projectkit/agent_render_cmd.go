package projectkit

import (
	"github.com/orbiqd/orbiqd-projectkit/internal/app/projectkit/action"
	"github.com/orbiqd/orbiqd-projectkit/internal/pkg/git"
	agentAPI "github.com/orbiqd/orbiqd-projectkit/pkg/agent"
	instructionAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/instruction"
	skillAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/skill"
	projectAPI "github.com/orbiqd/orbiqd-projectkit/pkg/project"
)

type AgentRenderCmd struct{}

func (cmd *AgentRenderCmd) Run(
	gitFs git.Fs,
	config *projectAPI.Config,
	instructionRepository instructionAPI.Repository,
	skillRepository skillAPI.Repository,
	agentRegistry agentAPI.Registry,
) error {
	return action.NewRenderAgentAction(gitFs, *config, agentRegistry, skillRepository, instructionRepository).Run()
}
