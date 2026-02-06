package project

import (
	"errors"
	"path/filepath"
	"testing"

	agentAPI "github.com/orbiqd/orbiqd-projectkit/pkg/agent"
	"github.com/orbiqd/orbiqd-projectkit/pkg/ai"
	"github.com/orbiqd/orbiqd-projectkit/pkg/ai/instruction"
	"github.com/orbiqd/orbiqd-projectkit/pkg/ai/workflow"
	projectAPI "github.com/orbiqd/orbiqd-projectkit/pkg/project"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigLoader_ResolveConfigPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		homeDir          string
		homeDirErr       error
		cwd              string
		cwdErr           error
		homeConfigExists bool
		cwdConfigExists  bool
		wantPaths        []string
		wantErr          error
	}{
		{
			name:             "WhenBothConfigsExist_ThenReturnsBothPaths",
			homeDir:          "/home/user",
			cwd:              "/project",
			homeConfigExists: true,
			cwdConfigExists:  true,
			wantPaths:        []string{"/home/user/.projectkit.yaml", "/project/.projectkit.yaml"},
		},
		{
			name:             "WhenOnlyHomeConfigExists_ThenReturnsHomePath",
			homeDir:          "/home/user",
			cwd:              "/project",
			homeConfigExists: true,
			cwdConfigExists:  false,
			wantPaths:        []string{"/home/user/.projectkit.yaml"},
		},
		{
			name:             "WhenOnlyCwdConfigExists_ThenReturnsCwdPath",
			homeDir:          "/home/user",
			cwd:              "/project",
			homeConfigExists: false,
			cwdConfigExists:  true,
			wantPaths:        []string{"/project/.projectkit.yaml"},
		},
		{
			name:             "WhenNoConfigsExist_ThenReturnsError",
			homeDir:          "/home/user",
			cwd:              "/project",
			homeConfigExists: false,
			cwdConfigExists:  false,
			wantErr:          projectAPI.ErrNoConfigResolved,
		},
		{
			name:            "WhenHomeDirFails_ThenReturnsCwdPath",
			homeDirErr:      errors.New("home dir error"),
			cwd:             "/project",
			cwdConfigExists: true,
			wantPaths:       []string{"/project/.projectkit.yaml"},
		},
		{
			name:       "WhenBothResolversFail_ThenReturnsError",
			homeDirErr: errors.New("home dir error"),
			cwdErr:     errors.New("cwd error"),
			wantErr:    projectAPI.ErrNoConfigResolved,
		},
		{
			name:             "WhenHomeEqualsCwd_ThenDeduplicates",
			homeDir:          "/same/dir",
			cwd:              "/same/dir",
			homeConfigExists: true,
			cwdConfigExists:  true,
			wantPaths:        []string{"/same/dir/.projectkit.yaml"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			fs := afero.NewMemMapFs()

			if tt.homeConfigExists && tt.homeDir != "" {
				require.NoError(t, afero.WriteFile(fs, filepath.Join(tt.homeDir, ConfigFileName), []byte{}, 0644))
			}
			if tt.cwdConfigExists && tt.cwd != "" {
				require.NoError(t, afero.WriteFile(fs, filepath.Join(tt.cwd, ConfigFileName), []byte{}, 0644))
			}

			loader := NewConfigLoader(
				WithConfigLoaderFs(fs),
				WithConfigLoaderGetHomeDirFn(func() (string, error) { return tt.homeDir, tt.homeDirErr }),
				WithConfigLoaderGetWorkDirFn(func() (string, error) { return tt.cwd, tt.cwdErr }),
			)

			// Act
			paths, err := loader.resolvePaths()

			// Assert
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, paths)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantPaths, paths)
			}
		})
	}
}

func TestConfigLoader_Load(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		fileContent string
		fileExists  bool
		wantConfig  *projectAPI.Config
		wantErr     error
	}{
		{
			name:       "WhenValidYAML_ThenReturnsConfig",
			fileExists: true,
			fileContent: `
rulebook:
  sources:
    - uri: "file://./rulebooks/general"
`,
			wantConfig: &projectAPI.Config{
				Rulebook: &projectAPI.RulebookConfig{
					Sources: []projectAPI.RulebookSourceConfig{
						{URI: "file://./rulebooks/general"},
					},
				},
			},
		},
		{
			name:       "WhenFileNotExists_ThenReturnsError",
			fileExists: false,
			wantErr:    projectAPI.ErrConfigLoadFailed,
		},
		{
			name:        "WhenInvalidYAML_ThenReturnsError",
			fileExists:  true,
			fileContent: "invalid: yaml: content: [",
			wantErr:     projectAPI.ErrConfigLoadFailed,
		},
		{
			name:        "WhenEmptyFile_ThenReturnsEmptyConfig",
			fileExists:  true,
			fileContent: "",
			wantConfig:  &projectAPI.Config{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			fs := afero.NewMemMapFs()
			configPath := "/test/.projectkit.yaml"

			if tt.fileExists {
				require.NoError(t, afero.WriteFile(fs, configPath, []byte(tt.fileContent), 0644))
			}

			loader := NewConfigLoader(WithConfigLoaderFs(fs))

			// Act
			config, err := loader.load(configPath)

			// Assert
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, config)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantConfig, config)
			}
		})
	}
}

func TestConfigLoader_validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  projectAPI.Config
		wantErr error
	}{
		{
			name:    "WhenEmptyConfig_ThenReturnsNil",
			config:  projectAPI.Config{},
			wantErr: nil,
		},
		{
			name: "WhenValidConfig_ThenReturnsNil",
			config: projectAPI.Config{
				Rulebook: &projectAPI.RulebookConfig{
					Sources: []projectAPI.RulebookSourceConfig{{URI: "file://x"}},
				},
			},
			wantErr: nil,
		},
		{
			name: "WhenInvalidConfig_ThenReturnsValidationError",
			config: projectAPI.Config{
				Rulebook: &projectAPI.RulebookConfig{
					Sources: []projectAPI.RulebookSourceConfig{},
				},
			},
			wantErr: projectAPI.ErrConfigValidationFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			loader := NewConfigLoader()
			err := loader.validate(tt.config)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestInit_WhenPackageImported_ThenDefaultLoaderIsRegistered(t *testing.T) {
	// Act
	loader, err := projectAPI.DefaultConfigLoader()

	// Assert
	require.NoError(t, err)
	require.NotNil(t, loader)

	_, ok := loader.(*ConfigLoader)
	assert.True(t, ok, "expected loader to be *ConfigLoader")
}

func TestConfigLoader_LoadIntegration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		homeConfig  string
		cwdConfig   string
		wantErr     error
		wantSources []string
	}{
		{
			name:    "WhenNoConfigFiles_ThenReturnsNoConfigError",
			wantErr: projectAPI.ErrNoConfigResolved,
		},
		{
			name: "WhenHomeConfigOnly_ThenReturnsConfig",
			homeConfig: `
rulebook:
  sources:
    - uri: "file://home-source"
`,
			wantSources: []string{"file://home-source"},
		},
		{
			name: "WhenCwdConfigOnly_ThenReturnsConfig",
			cwdConfig: `
rulebook:
  sources:
    - uri: "file://cwd-source"
`,
			wantSources: []string{"file://cwd-source"},
		},
		{
			name: "WhenBothConfigs_ThenMergesInOrder",
			homeConfig: `
rulebook:
  sources:
    - uri: "file://home-source"
`,
			cwdConfig: `
rulebook:
  sources:
    - uri: "file://cwd-source"
`,
			wantSources: []string{"file://home-source", "file://cwd-source"},
		},
		{
			name:       "WhenInvalidYAML_ThenReturnsLoadError",
			homeConfig: "invalid: yaml: [",
			wantErr:    projectAPI.ErrConfigLoadFailed,
		},
		{
			name: "WhenValidationFails_ThenReturnsValidationError",
			homeConfig: `
rulebook:
  sources: []
`,
			wantErr: projectAPI.ErrConfigValidationFailed,
		},
		{
			name: "WhenSecondConfigInvalid_ThenReturnsError",
			homeConfig: `
rulebook:
  sources:
    - uri: "file://home-source"
`,
			cwdConfig: "invalid: yaml: [",
			wantErr:   projectAPI.ErrConfigLoadFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			fs := afero.NewMemMapFs()
			homeDir := "/home/user"
			cwdDir := "/project"

			if tt.homeConfig != "" {
				require.NoError(t, afero.WriteFile(fs,
					filepath.Join(homeDir, ConfigFileName),
					[]byte(tt.homeConfig), 0644))
			}
			if tt.cwdConfig != "" {
				require.NoError(t, afero.WriteFile(fs,
					filepath.Join(cwdDir, ConfigFileName),
					[]byte(tt.cwdConfig), 0644))
			}

			loader := NewConfigLoader(
				WithConfigLoaderFs(fs),
				WithConfigLoaderGetHomeDirFn(func() (string, error) { return homeDir, nil }),
				WithConfigLoaderGetWorkDirFn(func() (string, error) { return cwdDir, nil }),
			)

			// Act
			config, err := loader.Load()

			// Assert
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, config)
			} else {
				require.NoError(t, err)
				require.NotNil(t, config)

				var gotSources []string
				for _, s := range config.Rulebook.Sources {
					gotSources = append(gotSources, s.URI)
				}
				assert.Equal(t, tt.wantSources, gotSources)
			}
		})
	}
}

func TestConfigLoader_merge(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		configs         []projectAPI.Config
		wantAgents      []agentAPI.Config
		wantRulebook    []projectAPI.RulebookSourceConfig
		wantInstruction []instruction.SourceConfig
		wantWorkflows   []workflow.SourceConfig
	}{
		{
			name:            "WhenEmptySlice_ThenReturnsInitializedEmptyStructures",
			configs:         []projectAPI.Config{},
			wantRulebook:    nil,
			wantInstruction: nil,
			wantWorkflows:   nil,
		},
		{
			name: "WhenOneConfigWithOnlyRulebook_ThenReturnsRulebookSources",
			configs: []projectAPI.Config{
				{
					Rulebook: &projectAPI.RulebookConfig{
						Sources: []projectAPI.RulebookSourceConfig{
							{URI: "file://A"},
							{URI: "file://B"},
						},
					},
				},
			},
			wantRulebook:    []projectAPI.RulebookSourceConfig{{URI: "file://A"}, {URI: "file://B"}},
			wantInstruction: nil,
			wantWorkflows:   nil,
		},
		{
			name: "WhenOneConfigWithOnlyAIInstruction_ThenReturnsInstructionSources",
			configs: []projectAPI.Config{
				{
					AI: &ai.Config{
						Instruction: &instruction.Config{
							Sources: []instruction.SourceConfig{{URI: "file://A"}},
						},
					},
				},
			},
			wantRulebook:    nil,
			wantInstruction: []instruction.SourceConfig{{URI: "file://A"}},
			wantWorkflows:   nil,
		},
		{
			name: "WhenOneConfigWithOnlyAIWorkflows_ThenReturnsWorkflowsSources",
			configs: []projectAPI.Config{
				{
					AI: &ai.Config{
						Workflows: &workflow.Config{
							Sources: []workflow.SourceConfig{{URI: "file://A"}},
						},
					},
				},
			},
			wantRulebook:    nil,
			wantInstruction: nil,
			wantWorkflows:   []workflow.SourceConfig{{URI: "file://A"}},
		},
		{
			name: "WhenTwoConfigsWithRulebook_ThenMergesSourcesInOrder",
			configs: []projectAPI.Config{
				{
					Rulebook: &projectAPI.RulebookConfig{
						Sources: []projectAPI.RulebookSourceConfig{{URI: "file://A"}, {URI: "file://B"}},
					},
				},
				{
					Rulebook: &projectAPI.RulebookConfig{
						Sources: []projectAPI.RulebookSourceConfig{{URI: "file://C"}, {URI: "file://D"}},
					},
				},
			},
			wantRulebook:    []projectAPI.RulebookSourceConfig{{URI: "file://A"}, {URI: "file://B"}, {URI: "file://C"}, {URI: "file://D"}},
			wantInstruction: nil,
			wantWorkflows:   nil,
		},
		{
			name: "WhenTwoConfigsWithAIInstruction_ThenMergesSourcesInOrder",
			configs: []projectAPI.Config{
				{
					AI: &ai.Config{
						Instruction: &instruction.Config{
							Sources: []instruction.SourceConfig{{URI: "file://A"}},
						},
					},
				},
				{
					AI: &ai.Config{
						Instruction: &instruction.Config{
							Sources: []instruction.SourceConfig{{URI: "file://B"}},
						},
					},
				},
			},
			wantRulebook:    nil,
			wantInstruction: []instruction.SourceConfig{{URI: "file://A"}, {URI: "file://B"}},
			wantWorkflows:   nil,
		},
		{
			name: "WhenTwoConfigsWithAIWorkflows_ThenMergesSourcesInOrder",
			configs: []projectAPI.Config{
				{
					AI: &ai.Config{
						Workflows: &workflow.Config{
							Sources: []workflow.SourceConfig{{URI: "file://A"}},
						},
					},
				},
				{
					AI: &ai.Config{
						Workflows: &workflow.Config{
							Sources: []workflow.SourceConfig{{URI: "file://B"}},
						},
					},
				},
			},
			wantRulebook:    nil,
			wantInstruction: nil,
			wantWorkflows:   []workflow.SourceConfig{{URI: "file://A"}, {URI: "file://B"}},
		},
		{
			name: "WhenConfigWithNilRulebook_ThenDoesNotCrash",
			configs: []projectAPI.Config{
				{Rulebook: nil},
			},
			wantRulebook:    nil,
			wantInstruction: nil,
			wantWorkflows:   nil,
		},
		{
			name: "WhenConfigWithNilAI_ThenDoesNotCrash",
			configs: []projectAPI.Config{
				{AI: nil},
			},
			wantRulebook:    nil,
			wantInstruction: nil,
			wantWorkflows:   nil,
		},
		{
			name: "WhenConfigWithAIButNilInstruction_ThenDoesNotCrash",
			configs: []projectAPI.Config{
				{AI: &ai.Config{Instruction: nil}},
			},
			wantRulebook:    nil,
			wantInstruction: nil,
			wantWorkflows:   nil,
		},
		{
			name: "WhenConfigWithAIButNilWorkflows_ThenDoesNotCrash",
			configs: []projectAPI.Config{
				{AI: &ai.Config{Workflows: nil}},
			},
			wantRulebook:    nil,
			wantInstruction: nil,
			wantWorkflows:   nil,
		},
		{
			name: "WhenThreeConfigs_ThenMergesAllInOrder",
			configs: []projectAPI.Config{
				{
					Rulebook: &projectAPI.RulebookConfig{
						Sources: []projectAPI.RulebookSourceConfig{{URI: "file://A"}},
					},
				},
				{
					Rulebook: &projectAPI.RulebookConfig{
						Sources: []projectAPI.RulebookSourceConfig{{URI: "file://B"}},
					},
				},
				{
					Rulebook: &projectAPI.RulebookConfig{
						Sources: []projectAPI.RulebookSourceConfig{{URI: "file://C"}},
					},
				},
			},
			wantRulebook:    []projectAPI.RulebookSourceConfig{{URI: "file://A"}, {URI: "file://B"}, {URI: "file://C"}},
			wantInstruction: nil,
			wantWorkflows:   nil,
		},
		{
			name: "WhenMixedFields_ThenMergesEachFieldIndependently",
			configs: []projectAPI.Config{
				{
					Rulebook: &projectAPI.RulebookConfig{
						Sources: []projectAPI.RulebookSourceConfig{{URI: "file://A"}},
					},
					AI: &ai.Config{
						Instruction: &instruction.Config{
							Sources: []instruction.SourceConfig{{URI: "file://B"}},
						},
					},
				},
				{
					Rulebook: &projectAPI.RulebookConfig{
						Sources: []projectAPI.RulebookSourceConfig{{URI: "file://C"}},
					},
				},
			},
			wantRulebook:    []projectAPI.RulebookSourceConfig{{URI: "file://A"}, {URI: "file://C"}},
			wantInstruction: []instruction.SourceConfig{{URI: "file://B"}},
			wantWorkflows:   nil,
		},
		{
			name: "WhenOneConfigWithOnlyAgents_ThenReturnsAgents",
			configs: []projectAPI.Config{
				{
					Agents: []agentAPI.Config{
						{Kind: "claude"},
						{Kind: "cursor"},
					},
				},
			},
			wantAgents:      []agentAPI.Config{{Kind: "claude"}, {Kind: "cursor"}},
			wantRulebook:    nil,
			wantInstruction: nil,
			wantWorkflows:   nil,
		},
		{
			name: "WhenTwoConfigsWithAgents_ThenMergesAgentsInOrder",
			configs: []projectAPI.Config{
				{
					Agents: []agentAPI.Config{
						{Kind: "claude"},
					},
				},
				{
					Agents: []agentAPI.Config{
						{Kind: "cursor"},
					},
				},
			},
			wantAgents:      []agentAPI.Config{{Kind: "claude"}, {Kind: "cursor"}},
			wantRulebook:    nil,
			wantInstruction: nil,
			wantWorkflows:   nil,
		},
		{
			name: "WhenConfigWithNilAgents_ThenDoesNotCrash",
			configs: []projectAPI.Config{
				{Agents: nil},
			},
			wantAgents:      nil,
			wantRulebook:    nil,
			wantInstruction: nil,
			wantWorkflows:   nil,
		},
		{
			name: "WhenAgentsHaveOptions_ThenOptionsArePreserved",
			configs: []projectAPI.Config{
				{
					Agents: []agentAPI.Config{
						{Kind: "claude", Options: map[string]any{"model": "sonnet"}},
					},
				},
			},
			wantAgents:      []agentAPI.Config{{Kind: "claude", Options: map[string]any{"model": "sonnet"}}},
			wantRulebook:    nil,
			wantInstruction: nil,
			wantWorkflows:   nil,
		},
		{
			name: "WhenAgentsMixedWithOtherFields_ThenMergesEachFieldIndependently",
			configs: []projectAPI.Config{
				{
					Agents: []agentAPI.Config{{Kind: "claude"}},
					Rulebook: &projectAPI.RulebookConfig{
						Sources: []projectAPI.RulebookSourceConfig{{URI: "file://A"}},
					},
					AI: &ai.Config{
						Instruction: &instruction.Config{
							Sources: []instruction.SourceConfig{{URI: "file://B"}},
						},
					},
				},
				{
					Agents: []agentAPI.Config{{Kind: "cursor"}},
					Rulebook: &projectAPI.RulebookConfig{
						Sources: []projectAPI.RulebookSourceConfig{{URI: "file://C"}},
					},
				},
			},
			wantAgents:      []agentAPI.Config{{Kind: "claude"}, {Kind: "cursor"}},
			wantRulebook:    []projectAPI.RulebookSourceConfig{{URI: "file://A"}, {URI: "file://C"}},
			wantInstruction: []instruction.SourceConfig{{URI: "file://B"}},
			wantWorkflows:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			loader := NewConfigLoader()

			// Act
			result := loader.merge(tt.configs...)

			// Assert
			require.NotNil(t, result.Rulebook)
			require.NotNil(t, result.AI)
			require.NotNil(t, result.AI.Instruction)
			require.NotNil(t, result.AI.Skill)
			require.NotNil(t, result.AI.Workflows)

			assert.Equal(t, tt.wantAgents, result.Agents)
			assert.Equal(t, tt.wantRulebook, result.Rulebook.Sources)
			assert.Equal(t, tt.wantInstruction, result.AI.Instruction.Sources)
			assert.Equal(t, tt.wantWorkflows, result.AI.Workflows.Sources)
		})
	}
}
