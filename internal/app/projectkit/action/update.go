package action

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/orbiqd/orbiqd-projectkit/internal/app/projectkit/loader"
	instructionAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/instruction"
	mcpAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/mcp"
	skillAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/skill"
	workflowAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/workflow"
	standardAPI "github.com/orbiqd/orbiqd-projectkit/pkg/doc/standard"
	projectAPI "github.com/orbiqd/orbiqd-projectkit/pkg/project"
	sourceAPI "github.com/orbiqd/orbiqd-projectkit/pkg/source"
)

type UpdateAction struct {
	config                projectAPI.Config
	sourceResolver        sourceAPI.Resolver
	instructionRepository instructionAPI.Repository
	skillRepository       skillAPI.Repository
	workflowRepository    workflowAPI.Repository
	mcpRepository         mcpAPI.Repository
	standardRepository    standardAPI.Repository
}

func NewUpdateAction(
	config projectAPI.Config,
	sourceResolver sourceAPI.Resolver,
	instructionRepository instructionAPI.Repository,
	skillRepository skillAPI.Repository,
	workflowRepository workflowAPI.Repository,
	mcpRepository mcpAPI.Repository,
	standardRepository standardAPI.Repository,
) *UpdateAction {
	return &UpdateAction{
		config:                config,
		sourceResolver:        sourceResolver,
		instructionRepository: instructionRepository,
		skillRepository:       skillRepository,
		workflowRepository:    workflowRepository,
		mcpRepository:         mcpRepository,
		standardRepository:    standardRepository,
	}
}

func (action *UpdateAction) Run() error {
	var instructions []instructionAPI.Instructions
	if action.config.AI != nil && action.config.AI.Instruction != nil {
		instructionsSet, err := loader.LoadAiInstructionsFromConfig(*action.config.AI.Instruction, action.sourceResolver)
		if err != nil {
			return fmt.Errorf("load instructions from config: %w", err)
		}

		instructions = append(instructions, instructionsSet...)
	}

	var skills []skillAPI.Skill
	if action.config.AI != nil && action.config.AI.Skill != nil {
		skillsSet, err := loader.LoadAiSkillsFromConfig(*action.config.AI.Skill, action.sourceResolver)
		if err != nil {
			return fmt.Errorf("load skills from config: %w", err)
		}

		skills = append(skills, skillsSet...)
	}

	var workflows []workflowAPI.Workflow
	if action.config.AI != nil && action.config.AI.Workflows != nil {
		workflowsSet, err := loader.LoadWorkflowsFromConfig(*action.config.AI.Workflows, action.sourceResolver)
		if err != nil {
			return fmt.Errorf("load workflows from config: %w", err)
		}

		workflows = append(workflows, workflowsSet...)
	}

	var mcpServers []mcpAPI.MCPServer
	if action.config.AI != nil && action.config.AI.MCP != nil {
		mcpServersSet, err := loader.LoadAiMCPServersFromConfig(*action.config.AI.MCP, action.sourceResolver)
		if err != nil {
			return fmt.Errorf("load mcp servers from config: %w", err)
		}

		mcpServers = append(mcpServers, mcpServersSet...)
	}

	var standards []standardAPI.Standard
	if action.config.Docs != nil && action.config.Docs.Standard != nil {
		standardsSet, err := loader.LoadDocStandardsFromConfig(*action.config.Docs.Standard, action.sourceResolver)
		if err != nil {
			return fmt.Errorf("load standards from config: %w", err)
		}

		standards = append(standards, standardsSet...)
	}

	if action.config.Rulebook != nil {
		rulebooks, err := loader.LoadRulebooksFromConfig(*action.config.Rulebook, action.sourceResolver)
		if err != nil {
			return fmt.Errorf("load rule books from config: %w", err)
		}

		for _, rulebook := range rulebooks {
			instructions = append(instructions, rulebook.AI.Instructions...)
			skills = append(skills, rulebook.AI.Skills...)
			workflows = append(workflows, rulebook.AI.Workflows...)
			mcpServers = append(mcpServers, rulebook.AI.MCPServers...)
			standards = append(standards, rulebook.Doc.Standards...)
		}
	}

	executablePath, err := action.resolveExecutablePath()
	if err != nil {
		return err
	}

	mcpServers = append(mcpServers, mcpAPI.MCPServer{
		Name: "projectkit",
		STDIO: &mcpAPI.STDIOMCPServer{
			ExecutablePath: executablePath,
			Arguments:      []string{"mcp", "server"},
		},
	})

	err = action.standardRepository.RemoveAll()
	if err != nil {
		return fmt.Errorf("remove all standards from repository: %w", err)
	}
	for _, standardItem := range standards {
		err := action.standardRepository.AddStandard(standardItem)
		if err != nil {
			return fmt.Errorf("add standard to repository: %w", err)
		}
	}
	slog.Info("Standards added to repository.", slog.Int("count", len(standards)))

	err = action.instructionRepository.RemoveAll()
	if err != nil {
		return fmt.Errorf("remove all instructions from repository: %w", err)
	}
	for _, instructionsItem := range instructions {
		err := action.instructionRepository.AddInstructions(instructionsItem)
		if err != nil {
			return fmt.Errorf("add instructions to repository: %w", err)
		}
	}
	slog.Info("Instructions added to repository.", slog.Int("count", len(instructions)))

	err = action.skillRepository.RemoveAll()
	if err != nil {
		return fmt.Errorf("remove all skills from repository: %w", err)
	}
	for _, skill := range skills {
		err := action.skillRepository.AddSkill(skill)
		if err != nil {
			return fmt.Errorf("add skill: %w", err)
		}
	}
	slog.Info("Skills added to repository.", slog.Int("count", len(skills)))

	err = action.workflowRepository.RemoveAllWorkflows()
	if err != nil {
		return fmt.Errorf("remove all workflows from repository: %w", err)
	}
	for _, workflow := range workflows {
		err := action.workflowRepository.AddWorkflow(workflow)
		if err != nil {
			return fmt.Errorf("add workflow: %w", err)
		}
	}
	slog.Info("Workflows added to repository.", slog.Int("count", len(workflows)))

	err = action.mcpRepository.RemoveAll()
	if err != nil {
		return fmt.Errorf("remove all mcp servers from repository: %w", err)
	}
	for _, mcpServer := range mcpServers {
		err := action.mcpRepository.AddMCPServer(mcpServer)
		if err != nil {
			return fmt.Errorf("add mcp server: %w", err)
		}
	}
	slog.Info("MCP servers added to repository.", slog.Int("count", len(mcpServers)))

	return nil
}

func (action *UpdateAction) resolveExecutablePath() (string, error) {
	if envPath := os.Getenv("BRIEFKIT_BINARY_PATH"); envPath != "" {
		return envPath, nil
	}

	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolve executable path: %w", err)
	}

	return execPath, nil
}
