package action

import (
	"fmt"
	"log/slog"

	"github.com/orbiqd/orbiqd-projectkit/internal/app/projectkit/loader"
	"github.com/orbiqd/orbiqd-projectkit/internal/pkg/git"
	agentAPI "github.com/orbiqd/orbiqd-projectkit/pkg/agent"
	instructionAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/instruction"
	mcpAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/mcp"
	skillAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/skill"
	projectAPI "github.com/orbiqd/orbiqd-projectkit/pkg/project"
)

type RenderAgentAction struct {
	gitFs                 git.Fs
	config                projectAPI.Config
	agentRegistry         agentAPI.Registry
	skillRepository       skillAPI.Repository
	instructionRepository instructionAPI.Repository
	mcpRepository         mcpAPI.Repository
}

func NewRenderAgentAction(
	gitFs git.Fs,
	config projectAPI.Config,
	agentRegistry agentAPI.Registry,
	skillRepository skillAPI.Repository,
	instructionRepository instructionAPI.Repository,
	mcpRepository mcpAPI.Repository,
) *RenderAgentAction {
	return &RenderAgentAction{
		gitFs:                 gitFs,
		config:                config,
		agentRegistry:         agentRegistry,
		skillRepository:       skillRepository,
		instructionRepository: instructionRepository,
		mcpRepository:         mcpRepository,
	}
}

func (action *RenderAgentAction) Run() error {
	agents, err := loader.LoadAgentsFromConfig(action.config, action.agentRegistry)
	if err != nil {
		return fmt.Errorf("load agents: %w", err)
	}

	if len(agents) == 0 {
		slog.Warn("No agents to render.")
		return nil
	}

	instructions, err := action.instructionRepository.GetAll()
	if err != nil {
		return fmt.Errorf("get all instructions: %w", err)
	}

	for _, agent := range agents {
		agentKind := agent.GetKind()
		logger := slog.Default().With(slog.String("agentKind", string(agentKind)))

		logger.Debug("Setting up agent.")

		err = agent.RenderInstructions(instructions)
		if err != nil {
			return fmt.Errorf("render instructions: %w", err)
		}

		err = agent.RebuildSkills(action.skillRepository)
		if err != nil {
			return fmt.Errorf("rebuild skills: %w", err)
		}

		mcpServers, err := action.mcpRepository.GetAll()
		if err != nil {
			return fmt.Errorf("get all mcp servers: %w", err)
		}

		err = agent.RenderMCPServers(mcpServers)
		if err != nil {
			return fmt.Errorf("render mcp servers: %w", err)
		}

		for _, pattern := range agent.GitIgnorePatterns() {
			isExcluded, err := git.IsExcluded(action.gitFs, pattern)
			if err != nil {
				return fmt.Errorf("check excluded pattern %s: %w", pattern, err)
			}

			if isExcluded {
				continue
			}

			err = git.Exclude(action.gitFs, pattern)
			if err != nil {
				return fmt.Errorf("exclude pattern %s: %w", pattern, err)
			}

			slog.Debug("Added GIT excluding pattern.",
				slog.String("pattern", pattern),
				slog.String("agentKind", string(agentKind)),
			)
		}

		slog.Info("Agent set up finished.", slog.String("agentKind", string(agentKind)))
	}

	slog.Info("All agents configured.", slog.Int("agentsCount", len(agents)))

	return nil
}
