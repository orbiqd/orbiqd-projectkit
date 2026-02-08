package projectkit

import (
	"fmt"
	"log/slog"
	"maps"
	"os"
	"slices"

	"github.com/orbiqd/orbiqd-projectkit/internal/pkg/git"
	"github.com/orbiqd/orbiqd-projectkit/internal/pkg/project"
	agentAPI "github.com/orbiqd/orbiqd-projectkit/pkg/agent"
	instructionAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/instruction"
	skillAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/skill"
	projectAPI "github.com/orbiqd/orbiqd-projectkit/pkg/project"
	sourceAPI "github.com/orbiqd/orbiqd-projectkit/pkg/source"
	"github.com/spf13/afero"
)

type SetupCmd struct {
}

func (cmd *SetupCmd) Run(
	configLoader projectAPI.ConfigLoader,
	sourceResolver sourceAPI.Resolver,
	instructionRepository instructionAPI.Repository,
	skillRepository skillAPI.Repository,
	agentRegistry agentAPI.Registry,
) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	slog.Debug("Discovered current working directory.", slog.String("workingDirectoryPath", currentDir))

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get user home directory: %w", err)
	}

	projectFs, err := project.CreateProjectFs(afero.NewOsFs(), currentDir, homeDir)
	if err != nil {
		return fmt.Errorf("create project fs: %w", err)
	}

	slog.Info("Project found. Setting up.")

	config, err := configLoader.Load()
	if err != nil {
		return fmt.Errorf("config loader: %w", err)
	}
	slog.Debug("Configuration loaded.")

	for _, instructionSourceConfig := range config.AI.Instruction.Sources {
		instructionsUri := instructionSourceConfig.URI
		instructionSource, err := sourceResolver.Resolve(instructionsUri)
		if err != nil {
			return fmt.Errorf("resolve: %s: %w", instructionsUri, err)
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
				slog.String("sourceUri", instructionsUri),
				slog.String("instructionsCategory", string(instructions.Category)),
				slog.Int("rulesCount", len(instructions.Rules)),
			)
		}
	}

	for _, skillConfig := range config.AI.Skill.Sources {
		skillsUri := skillConfig.URI
		skillsSource, err := sourceResolver.Resolve(skillConfig.URI)
		if err != nil {
			return fmt.Errorf("resolve: %s: %w", skillsUri, err)
		}

		skillLoader := skillAPI.NewLoader(skillsSource)

		skills, err := skillLoader.Load()
		if err != nil {
			return fmt.Errorf("load skills: %w", err)
		}

		for _, skill := range skills {
			err = skillRepository.AddSkill(skill)
			if err != nil {
				return fmt.Errorf("add skill: %w", err)
			}

			slog.Debug("AI Skill added to repository.",
				slog.String("sourceUri", skillsUri),
				slog.String("skillName", string(skill.Metadata.Name)),
			)
		}

		slog.Debug("AI Skills added to repository.",
			slog.String("sourceUri", skillsUri),
			slog.Int("skillsCount", len(skills)),
		)
	}

	if len(config.Agents) == 0 {
		slog.Warn("No agents configured.")
		return nil
	}

	agents := make(map[agentAPI.Kind]agentAPI.Agent)

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

		agents[agentKind] = agent
	}

	for agentKind, agent := range agents {
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

	err = cmd.setupGit(projectFs, currentDir, slices.Collect(maps.Values(agents)))
	if err != nil {
		return fmt.Errorf("setup git: %w", err)
	}

	return nil
}

func (cmd *SetupCmd) setupGit(projectFs afero.Fs, currentDir string, agents []agentAPI.Agent) error {
	gitFs, err := git.CreateGitFs(projectFs, currentDir)
	if err != nil {
		return fmt.Errorf("create git fs: %w", err)
	}

	var patterns []string

	for _, agent := range agents {
		patterns = append(patterns, agent.GitIgnorePatterns()...)
	}

	for _, pattern := range patterns {
		isExcluded, err := git.IsExcluded(gitFs, pattern)
		if err != nil {
			return fmt.Errorf("check excluded pattern %s: %w", pattern, err)
		}

		if isExcluded {
			continue
		}

		err = git.Exclude(gitFs, pattern)
		if err != nil {
			return fmt.Errorf("exclude pattern %s: %w", pattern, err)
		}

		slog.Debug("Added GIT excluding pattern.", slog.String("pattern", pattern))
	}

	return nil
}
