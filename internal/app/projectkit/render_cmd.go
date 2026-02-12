package projectkit

import (
	"github.com/orbiqd/orbiqd-projectkit/internal/app/projectkit/action"
	"github.com/orbiqd/orbiqd-projectkit/internal/pkg/doc/standard"
	"github.com/orbiqd/orbiqd-projectkit/internal/pkg/git"
	agentAPI "github.com/orbiqd/orbiqd-projectkit/pkg/agent"
	instructionAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/instruction"
	mcpAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/mcp"
	skillAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/skill"
	standardAPI "github.com/orbiqd/orbiqd-projectkit/pkg/doc/standard"
	projectAPI "github.com/orbiqd/orbiqd-projectkit/pkg/project"
)

type RenderCmd struct{}

func (cmd *RenderCmd) Run(
	gitFs git.Fs,
	config *projectAPI.Config,
	instructionRepository instructionAPI.Repository,
	skillRepository skillAPI.Repository,
	agentRegistry agentAPI.Registry,
	projectFs projectAPI.Fs,
	standardRepository standardAPI.Repository,
	mcpRepository mcpAPI.Repository,
) error {
	if err := action.NewRenderAgentAction(gitFs, *config, agentRegistry, skillRepository, instructionRepository, mcpRepository).Run(); err != nil {
		return err
	}

	renderers := map[string]standardAPI.Renderer{
		"markdown": standard.NewMarkdownRenderer(),
	}
	return action.NewRenderDocStandardAction(standardRepository, projectFs, config.Docs.Standard.Render, renderers).Run()
}
