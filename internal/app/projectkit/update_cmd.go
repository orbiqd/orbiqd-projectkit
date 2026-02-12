package projectkit

import (
	"github.com/orbiqd/orbiqd-projectkit/internal/app/projectkit/action"
	instructionAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/instruction"
	mcpAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/mcp"
	skillAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/skill"
	workflowAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/workflow"
	standardAPI "github.com/orbiqd/orbiqd-projectkit/pkg/doc/standard"
	projectAPI "github.com/orbiqd/orbiqd-projectkit/pkg/project"
	sourceAPI "github.com/orbiqd/orbiqd-projectkit/pkg/source"
)

type UpdateCmd struct{}

func (cmd *UpdateCmd) Run(
	config *projectAPI.Config,
	sourceResolver sourceAPI.Resolver,
	instructionRepository instructionAPI.Repository,
	skillRepository skillAPI.Repository,
	workflowRepository workflowAPI.Repository,
	mcpRepository mcpAPI.Repository,
	standardRepository standardAPI.Repository,
) error {
	return action.NewUpdateAction(*config, sourceResolver, instructionRepository, skillRepository, workflowRepository, mcpRepository, standardRepository).Run()
}
