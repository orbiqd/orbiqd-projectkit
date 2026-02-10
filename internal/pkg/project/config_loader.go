package project

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	projectAPI "github.com/orbiqd/orbiqd-projectkit/pkg/project"
	rulebookAPI "github.com/orbiqd/orbiqd-projectkit/pkg/rulebook"
	"github.com/spf13/afero"
	"sigs.k8s.io/yaml"

	"github.com/orbiqd/orbiqd-projectkit/pkg/ai"
	"github.com/orbiqd/orbiqd-projectkit/pkg/ai/instruction"
	"github.com/orbiqd/orbiqd-projectkit/pkg/ai/skill"
	"github.com/orbiqd/orbiqd-projectkit/pkg/ai/workflow"
)

const (
	// ConfigFileName is the default project configuration file name.
	ConfigFileName = ".projectkit.yaml"
)

// ConfigLoaderOpt configures a ConfigLoader.
type ConfigLoaderOpt func(*ConfigLoader)

// ConfigLoader loads and merges project configuration files.
type ConfigLoader struct {
	fs afero.Fs

	getWorkDirFn func() (string, error)
	getHomeDirFn func() (string, error)
}

var _ projectAPI.ConfigLoader = (*ConfigLoader)(nil)

// WithConfigLoaderFs sets the filesystem used to load configuration files.
func WithConfigLoaderFs(fs afero.Fs) ConfigLoaderOpt {
	return func(loader *ConfigLoader) {
		loader.fs = fs
	}
}

// WithConfigLoaderGetWorkDirFn sets the working directory resolver for config discovery.
func WithConfigLoaderGetWorkDirFn(fn func() (string, error)) ConfigLoaderOpt {
	return func(loader *ConfigLoader) {
		loader.getWorkDirFn = fn
	}
}

// WithConfigLoaderGetHomeDirFn sets the home directory resolver for config discovery.
func WithConfigLoaderGetHomeDirFn(fn func() (string, error)) ConfigLoaderOpt {
	return func(loader *ConfigLoader) {
		loader.getHomeDirFn = fn
	}
}

// NewConfigLoader builds a ConfigLoader with default dependencies.
func NewConfigLoader(opts ...ConfigLoaderOpt) *ConfigLoader {
	loader := &ConfigLoader{
		fs:           afero.NewOsFs(),
		getWorkDirFn: os.Getwd,
		getHomeDirFn: os.UserHomeDir,
	}

	for _, opt := range opts {
		opt(loader)
	}

	return loader
}

// Load resolves, validates, and merges configuration files.
func (loader *ConfigLoader) Load() (*projectAPI.Config, error) {
	paths, err := loader.resolvePaths()
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}

	var configs []projectAPI.Config

	for _, path := range paths {
		config, err := loader.load(path)
		if err != nil {
			return nil, fmt.Errorf("load config: %s: %w", path, err)
		}

		err = loader.validate(*config)
		if err != nil {
			return nil, fmt.Errorf("validate config: %s: %w", path, err)
		}

		configs = append(configs, *config)
	}

	result := loader.merge(configs...)

	return &result, nil
}

func (loader *ConfigLoader) resolvePaths() ([]string, error) {
	var paths []string

	if home, err := loader.getHomeDirFn(); err == nil {
		homePath, err := filepath.Abs(filepath.Join(home, ConfigFileName))
		if err == nil {
			if exists, _ := afero.Exists(loader.fs, homePath); exists {
				paths = append(paths, homePath)
			}
		}
	}

	if cwd, err := loader.getWorkDirFn(); err == nil {
		cwdPath, err := filepath.Abs(filepath.Join(cwd, ConfigFileName))
		if err == nil {
			if exists, _ := afero.Exists(loader.fs, cwdPath); exists {
				paths = append(paths, cwdPath)
			}
		}
	}

	if len(paths) == 2 && paths[0] == paths[1] {
		paths = paths[:1]
	}

	if len(paths) == 0 {
		return nil, projectAPI.ErrNoConfigResolved
	}

	return paths, nil
}

func (loader *ConfigLoader) load(configPath string) (*projectAPI.Config, error) {
	data, err := afero.ReadFile(loader.fs, configPath)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", configPath, projectAPI.ErrConfigLoadFailed)
	}

	var config projectAPI.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("unmarshal %s: %w", configPath, projectAPI.ErrConfigLoadFailed)
	}

	return &config, nil
}

func (loader *ConfigLoader) validate(config projectAPI.Config) error {
	validate := validator.New()

	if err := validate.Struct(config); err != nil {
		return fmt.Errorf("%w: %v", projectAPI.ErrConfigValidationFailed, err)
	}

	return nil
}

func (loader *ConfigLoader) merge(configs ...projectAPI.Config) projectAPI.Config {
	result := projectAPI.Config{
		Rulebook: &rulebookAPI.Config{},
		AI: &ai.Config{
			Instruction: &instruction.Config{},
			Skill:       &skill.Config{},
			Workflows:   &workflow.Config{},
		},
	}

	for _, cfg := range configs {
		result.Agents = append(result.Agents, cfg.Agents...)

		if cfg.Rulebook != nil {
			result.Rulebook.Sources = append(result.Rulebook.Sources, cfg.Rulebook.Sources...)
		}

		if cfg.AI != nil {
			if cfg.AI.Instruction != nil {
				result.AI.Instruction.Sources = append(result.AI.Instruction.Sources, cfg.AI.Instruction.Sources...)
			}
			if cfg.AI.Skill != nil {
				result.AI.Skill.Sources = append(result.AI.Skill.Sources, cfg.AI.Skill.Sources...)
			}
			if cfg.AI.Workflows != nil {
				result.AI.Workflows.Sources = append(result.AI.Workflows.Sources, cfg.AI.Workflows.Sources...)
			}
		}
	}

	return result
}

func init() {
	projectAPI.RegisterDefaultLoader(func() projectAPI.ConfigLoader {
		return NewConfigLoader()
	})
}
