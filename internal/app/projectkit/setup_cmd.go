package projectkit

import (
	"fmt"
	"log/slog"

	agentAPI "github.com/orbiqd/orbiqd-projectkit/pkg/agent"
	instructionAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/instruction"
	projectAPI "github.com/orbiqd/orbiqd-projectkit/pkg/project"
	sourceAPI "github.com/orbiqd/orbiqd-projectkit/pkg/source"
)

type SetupCmd struct {
}

func (cmd *SetupCmd) Run(
	configLoader projectAPI.ConfigLoader,
	sourceResolver sourceAPI.Resolver,
	instructionRepository instructionAPI.Repository,
	agentRegistry agentAPI.Registry,
) error {
	config, err := configLoader.Load()
	if err != nil {
		return fmt.Errorf("config loader: %w", err)
	}
	slog.Debug("Configuration loaded.")

	for _, instructionSourceConfig := range config.AI.Instruction.Sources {
		uri := instructionSourceConfig.URI
		instructionSource, err := sourceResolver.Resolve(uri)
		if err != nil {
			return fmt.Errorf("resolve: %s: %w", uri, err)
		}

		instructionLoader := instructionAPI.NewLoader(instructionSource)
		instructionsSet, err := instructionLoader.Load()
		if err != nil {
			return fmt.Errorf("load instructions: %w", err)
		}

		for _, instructions := range instructionsSet {
			err = instructionRepository.AddInstructions(instructions)
			if err != nil {
				return fmt.Errorf("add instructions: %w", err)
			}

			slog.Debug("AI Instructions added to repository.",
				slog.String("sourceUri", uri),
				slog.String("instructionsCategory", string(instructions.Category)),
				slog.Int("rulesCount", len(instructions.Rules)),
			)
		}
	}

	if len(config.Agents) == 0 {
		slog.Warn("No agents configured.")
		return nil
	}

	for _, agentConfig := range config.Agents {
		agentKind := agentConfig.Kind
		slog.Debug("Configuring agent.", slog.String("agentKind", string(agentKind)))

		provider, err := agentRegistry.GetByKind(agentKind)
		if err != nil {
			return fmt.Errorf("get provider: %s: %w", agentKind, err)
		}

		agent, err := provider.NewAgent(agentConfig.Options)
		if err != nil {
			return fmt.Errorf("create agent: %s: %w", agentKind, err)
		}

		instructions, err := instructionRepository.GetAll()
		if err != nil {
			return fmt.Errorf("get all instructions: %s: %w", agentKind, err)
		}

		err = agent.RenderInstructions(instructions)
		if err != nil {
			return fmt.Errorf("render instructions: %s: %w", agentKind, err)
		}
		slog.Debug("Instructions rendered.", slog.String("agentKind", string(agentKind)))

		slog.Info("Agent configured.", slog.String("agentKind", string(agentKind)))
	}

	return nil
}
