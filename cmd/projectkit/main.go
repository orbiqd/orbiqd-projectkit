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
	mcpInternal "github.com/orbiqd/orbiqd-projectkit/internal/pkg/ai/mcp"
	"github.com/orbiqd/orbiqd-projectkit/internal/pkg/ai/skill"
	"github.com/orbiqd/orbiqd-projectkit/internal/pkg/ai/workflow"
	"github.com/orbiqd/orbiqd-projectkit/internal/pkg/doc/standard"
	"github.com/orbiqd/orbiqd-projectkit/internal/pkg/git"
	"github.com/orbiqd/orbiqd-projectkit/internal/pkg/log"
	projectInternal "github.com/orbiqd/orbiqd-projectkit/internal/pkg/project"
	"github.com/orbiqd/orbiqd-projectkit/internal/pkg/source"
	agentAPI "github.com/orbiqd/orbiqd-projectkit/pkg/agent"
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

	err = runtime.BindSingletonProvider(projectInternal.NewProjectFsProvider())
	if err != nil {
		runtime.Fatalf("Failed to bind project filesystem provider: %v", err)
	}

	err = runtime.BindSingletonProvider(git.NewGitFsProvider())
	if err != nil {
		runtime.Fatalf("Failed to bind git filesystem provider: %v", err)
	}

	configLoader, err := projectAPI.DefaultConfigLoader()
	if err != nil {
		runtime.Fatalf("create config loader: %v", err)
	}
	runtime.BindTo(configLoader, (*projectAPI.ConfigLoader)(nil))

	err = runtime.BindSingletonProvider(projectInternal.NewConfigProvider())
	if err != nil {
		runtime.Fatalf("bind config provider: %v", err)
	}

	sourceDriverRepository := source.NewDriverRepository()
	err = sourceDriverRepository.RegisterDriver(source.NewLocalDriver())
	if err != nil {
		runtime.Fatalf("register local driver: %v", err)
	}

	sourceResolver := source.NewResolver(sourceDriverRepository)
	runtime.BindTo(sourceResolver, (*sourceAPI.Resolver)(nil))

	err = runtime.BindSingletonProvider(skill.NewFsRepositoryProvider())
	if err != nil {
		runtime.Fatalf("bind skill repository provider: %v", err)
		return
	}

	err = runtime.BindSingletonProvider(instruction.NewFsRepositoryProvider())
	if err != nil {
		runtime.Fatalf("bind instruction repository provider: %v", err)
		return
	}

	err = runtime.BindSingletonProvider(standard.NewFsRepositoryProvider())
	if err != nil {
		runtime.Fatalf("bind standard repository provider: %v", err)
		return
	}

	err = runtime.BindSingletonProvider(workflow.NewFsRepositoryProvider())
	if err != nil {
		runtime.Fatalf("bind workflow repository provider: %v", err)
		return
	}

	err = runtime.BindSingletonProvider(mcpInternal.NewFsRepositoryProvider())
	if err != nil {
		runtime.Fatalf("bind mcp repository provider: %v", err)
		return
	}

	err = runtime.BindSingletonProvider(agent.NewRegistryProvider(
		func(rootFs afero.Fs) agentAPI.Provider { return claude.NewProvider(rootFs) },
		func(rootFs afero.Fs) agentAPI.Provider { return codex.NewProvider(rootFs) },
	))
	if err != nil {
		runtime.Fatalf("bind agent registry provider: %v", err)
	}

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
