package projectkit

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/orbiqd/orbiqd-projectkit/internal/pkg/git"
	"github.com/orbiqd/orbiqd-projectkit/internal/pkg/project"
	agentAPI "github.com/orbiqd/orbiqd-projectkit/pkg/agent"
	instructionAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/instruction"
	skillAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/skill"
	projectAPI "github.com/orbiqd/orbiqd-projectkit/pkg/project"
	rulebookAPI "github.com/orbiqd/orbiqd-projectkit/pkg/rulebook"
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

	return cmd.execute(projectFs, currentDir, configLoader, sourceResolver, instructionRepository, skillRepository, agentRegistry)
}

func (cmd *SetupCmd) execute(
	projectFs afero.Fs,
	currentDir string,
	configLoader projectAPI.ConfigLoader,
	sourceResolver sourceAPI.Resolver,
	instructionRepository instructionAPI.Repository,
	skillRepository skillAPI.Repository,
	agentRegistry agentAPI.Registry,
) error {
	config, err := configLoader.Load()
	if err != nil {
		return fmt.Errorf("config loader: %w", err)
	}
	slog.Debug("Configuration loaded.")

	var instructions []instructionAPI.Instructions
	if config.AI != nil && config.AI.Instruction != nil {
		instructionsSet, err := cmd.loadInstructionsFromConfig(*config.AI.Instruction, sourceResolver)
		if err != nil {
			return fmt.Errorf("load instructions from config: %w", err)
		}

		instructions = append(instructions, instructionsSet...)
	}
	slog.Info("AI instructions loaded.", slog.Int("instructionsCount", len(instructions)))

	var skills []skillAPI.Skill
	if config.AI != nil && config.AI.Skill != nil {
		skillsSet, err := cmd.loadSkillsFromConfig(*config.AI.Skill, sourceResolver)
		if err != nil {
			return fmt.Errorf("load skills from config: %w", err)
		}

		skills = append(skills, skillsSet...)
	}
	slog.Info("AI skills loaded.", slog.Int("skillsCount", len(skills)))

	if config.Rulebook != nil {
		rulebooks, err := cmd.loadRulebooksFromConfig(*config.Rulebook, sourceResolver)
		if err != nil {
			return fmt.Errorf("load rule books from config: %w", err)
		}

		for _, rulebook := range rulebooks {
			instructions = append(instructions, rulebook.AI.Instructions...)
			skills = append(skills, rulebook.AI.Skills...)
		}
	}

	for _, instructionsItem := range instructions {
		err = instructionRepository.AddInstructions(instructionsItem)
		if err != nil {
			return fmt.Errorf("add instructions to repository: %w", err)
		}

		slog.Debug("AI Instructions added to repository.",
			slog.String("instructionsCategory", string(instructionsItem.Category)),
			slog.Int("rulesCount", len(instructionsItem.Rules)),
		)
	}

	for _, skill := range skills {
		err = skillRepository.AddSkill(skill)
		if err != nil {
			return fmt.Errorf("add skill: %w", err)
		}

		slog.Debug("AI Skill added to repository.",
			slog.String("skillName", string(skill.Metadata.Name)),
		)
	}

	agents, err := cmd.loadAgentsFromConfig(*config, agentRegistry)
	if err != nil {
		return fmt.Errorf("load agents: %w", err)
	}

	if len(agents) == 0 {
		slog.Warn("No agents configured.")
		return nil
	}

	for _, agent := range agents {
		err = cmd.setupAgent(projectFs, currentDir, agent, skillRepository, instructionRepository)
		if err != nil {
			return fmt.Errorf("setup agent: %w", err)
		}
	}

	slog.Info("All agents configured.", slog.Int("agentsCount", len(agents)))

	return nil
}

func (cmd *SetupCmd) loadRulebooksFromConfig(config rulebookAPI.Config, sourceResolver sourceAPI.Resolver) ([]rulebookAPI.Rulebook, error) {
	var rulebooks []rulebookAPI.Rulebook

	for _, rulebookSourceConfig := range config.Sources {
		rulebookUri := rulebookSourceConfig.URI

		rulebookSource, err := sourceResolver.Resolve(rulebookUri)
		if err != nil {
			return []rulebookAPI.Rulebook{}, fmt.Errorf("resolve rulebook uri: %w", err)
		}

		rulebook, err := rulebookAPI.NewLoader(rulebookSource).Load()
		if err != nil {
			return []rulebookAPI.Rulebook{}, fmt.Errorf("load rulebook: %w", err)
		}

		rulebooks = append(rulebooks, *rulebook)

		slog.Info("Loaded rulebook.",
			slog.String("sourceUri", rulebookUri),
			slog.Int("aiInstructionsCount", len(rulebook.AI.Instructions)),
			slog.Int("aiSkillsCount", len(rulebook.AI.Skills)),
		)
	}

	return rulebooks, nil
}

func (cmd *SetupCmd) setupAgent(projectFs afero.Fs, currentDir string, agent agentAPI.Agent, skillsRepository skillAPI.Repository, instructionRepository instructionAPI.Repository) error {
	agentKind := agent.GetKind()
	logger := slog.Default().With(slog.String("agentKind", string(agent.GetKind())))

	logger.Debug("Setting up agent.")

	instructions, err := instructionRepository.GetAll()
	if err != nil {
		return fmt.Errorf("get all instructions: %w", err)
	}

	err = agent.RenderInstructions(instructions)
	if err != nil {
		return fmt.Errorf("render instructions: %w", err)
	}

	err = agent.RebuildSkills(skillsRepository)
	if err != nil {
		return fmt.Errorf("rebuild skills: %w", err)
	}

	err = cmd.setupAgentGit(projectFs, currentDir, agent)
	if err != nil {
		return fmt.Errorf("git setup: %w", err)
	}

	slog.Info("Agent set up finished.", slog.String("agentKind", string(agentKind)))

	return nil
}

func (cmd *SetupCmd) setupAgentGit(projectFs afero.Fs, currentDir string, agent agentAPI.Agent) error {
	gitFs, err := git.CreateGitFs(projectFs, currentDir)
	if err != nil {
		return fmt.Errorf("git fs creation: %w", err)
	}

	for _, pattern := range agent.GitIgnorePatterns() {
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

		slog.Debug("Added GIT excluding pattern.",
			slog.String("pattern", pattern),
			slog.String("agentKind", string(agent.GetKind())),
		)
	}

	return nil
}

func (cmd *SetupCmd) loadAgentsFromConfig(config projectAPI.Config, agentRegistry agentAPI.Registry) ([]agentAPI.Agent, error) {
	var agents []agentAPI.Agent

	for _, agentConfig := range config.Agents {
		agentKind := agentConfig.Kind

		provider, err := agentRegistry.GetByKind(agentKind)
		if err != nil {
			return []agentAPI.Agent{}, fmt.Errorf("get provider: %s: %w", agentKind, err)
		}

		agent, err := provider.NewAgent(agentConfig.Options)
		if err != nil {
			return []agentAPI.Agent{}, fmt.Errorf("create agent: %s: %w", agentKind, err)
		}

		agents = append(agents, agent)

		slog.Info("Found agent in config.", slog.String("agentKind", string(agent.GetKind())))
	}

	return agents, nil
}

func (cmd *SetupCmd) loadSkillsFromConfig(config skillAPI.Config, sourceResolver sourceAPI.Resolver) ([]skillAPI.Skill, error) {
	var skills []skillAPI.Skill

	for _, skillConfig := range config.Sources {
		skillsUri := skillConfig.URI
		skillsSource, err := sourceResolver.Resolve(skillConfig.URI)
		if err != nil {
			return []skillAPI.Skill{}, fmt.Errorf("resolve: %s: %w", skillsUri, err)
		}

		skillLoader := skillAPI.NewLoader(skillsSource)

		skillsSet, err := skillLoader.Load()
		if err != nil {
			return []skillAPI.Skill{}, fmt.Errorf("load skills: %w", err)
		}

		slog.Debug("Loaded skills from config.",
			slog.String("sourceUri", skillsUri),
			slog.Int("skillsCount", len(skillsSet)),
		)

		skills = append(skills, skillsSet...)
	}

	return skills, nil
}

func (cmd *SetupCmd) loadInstructionsFromConfig(config instructionAPI.Config, sourceResolver sourceAPI.Resolver) ([]instructionAPI.Instructions, error) {
	var instructions []instructionAPI.Instructions

	for _, instructionSourceConfig := range config.Sources {
		instructionsUri := instructionSourceConfig.URI
		instructionSource, err := sourceResolver.Resolve(instructionsUri)
		if err != nil {
			return []instructionAPI.Instructions{}, fmt.Errorf("resolve: %s: %w", instructionsUri, err)
		}

		instructionLoader := instructionAPI.NewLoader(instructionSource)
		instructionsSet, err := instructionLoader.Load()
		if err != nil {
			return []instructionAPI.Instructions{}, fmt.Errorf("load instructions: %w", err)
		}

		instructions = append(instructions, instructionsSet...)
	}

	return instructions, nil
}
