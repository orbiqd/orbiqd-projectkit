package main

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/alecthomas/kong"
	"github.com/orbiqd/orbiqd-projectkit/internal/app/projectkit"
	"github.com/orbiqd/orbiqd-projectkit/internal/pkg/agent"
	"github.com/orbiqd/orbiqd-projectkit/internal/pkg/agent/claude"
	"github.com/orbiqd/orbiqd-projectkit/internal/pkg/agent/codex"
	"github.com/orbiqd/orbiqd-projectkit/internal/pkg/ai/instruction"
	"github.com/orbiqd/orbiqd-projectkit/internal/pkg/ai/skill"
	"github.com/orbiqd/orbiqd-projectkit/internal/pkg/log"
	_ "github.com/orbiqd/orbiqd-projectkit/internal/pkg/project"
	"github.com/orbiqd/orbiqd-projectkit/internal/pkg/source"
	agentAPI "github.com/orbiqd/orbiqd-projectkit/pkg/agent"
	instructionAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/instruction"
	skillAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/skill"
	projectAPI "github.com/orbiqd/orbiqd-projectkit/pkg/project"
	sourceAPI "github.com/orbiqd/orbiqd-projectkit/pkg/source"
	"github.com/spf13/afero"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	var cmd projectkit.Cmd

	runtime := kong.Parse(&cmd,
		kong.Name("projectkit"),
		kong.Description("OrbiqD ProjectKit"),
		kong.UsageOnError(),
		kong.Vars{"version": version + " (" + commit + ", " + date + ")"},
	)

	projectRootPath, err := os.Getwd()
	if err != nil {
		runtime.Fatalf("Failed to get current working directory: %v", err)
	}

	projectRootFs := afero.NewBasePathFs(afero.NewOsFs(), projectRootPath)

	logger, err := log.CreateLoggerFromConfig(cmd.Log)
	if err != nil {
		runtime.Fatalf("create logger: %v", err)
	}
	slog.SetDefault(logger)

	ctx := context.Background()
	err = runtime.BindToProvider(func() (context.Context, error) {
		return ctx, nil
	})
	if err != nil {
		runtime.Fatalf("bind context to provider: %v", err)
	}

	configLoader, err := projectAPI.DefaultConfigLoader()
	if err != nil {
		runtime.Fatalf("create config loader: %v", err)
	}
	runtime.BindTo(configLoader, (*projectAPI.ConfigLoader)(nil))

	sourceDriverRepository := source.NewDriverRepository()
	err = sourceDriverRepository.RegisterDriver(source.NewLocalDriver())
	if err != nil {
		runtime.Fatalf("register local driver: %v", err)
	}

	sourceResolver := source.NewResolver(sourceDriverRepository)
	runtime.BindTo(sourceResolver, (*sourceAPI.Resolver)(nil))

	instructionRepository := instruction.NewMemoryRepository()
	runtime.BindTo(instructionRepository, (*instructionAPI.Repository)(nil))

	agentRegistry := agent.NewStaticRegistry()
	err = agentRegistry.Register(codex.NewProvider(projectRootFs))
	if err != nil {
		runtime.Fatalf("register agent provider: %v", err)
	}
	err = agentRegistry.Register(claude.NewProvider(projectRootFs))
	if err != nil {
		runtime.Fatalf("register agent provider: %v", err)
	}
	runtime.BindTo(agentRegistry, (*agentAPI.Registry)(nil))

	skillRepository := skill.NewMemoryRepository()
	runtime.BindTo(skillRepository, (*skillAPI.Repository)(nil))

	err = runtime.Run()
	if err != nil {
		if errors.Is(err, projectAPI.ErrNoConfigResolved) {
			logger.Error("No config was resolved. Create a config in the current working directory or in the home directory.")
			os.Exit(NoConfigExitCode)
		}

		logger.Error("Failed to run.", slog.String("error", err.Error()))
		os.Exit(ErrorExitCode)
	}

	os.Exit(NoErrorExitCode)
}

var (
	NoErrorExitCode  = 0
	NoConfigExitCode = 1
	ErrorExitCode    = 255
)
