package rulebook

import (
	"errors"
	"testing"

	"github.com/orbiqd/orbiqd-projectkit/pkg/ai"
	"github.com/orbiqd/orbiqd-projectkit/pkg/ai/instruction"
	"github.com/orbiqd/orbiqd-projectkit/pkg/ai/mcp"
	"github.com/orbiqd/orbiqd-projectkit/pkg/ai/skill"
	"github.com/orbiqd/orbiqd-projectkit/pkg/ai/workflow"
	"github.com/orbiqd/orbiqd-projectkit/pkg/doc"
	"github.com/orbiqd/orbiqd-projectkit/pkg/doc/standard"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.nhat.io/aferomock"
)

func fsWithOpenError(t *testing.T, base afero.Fs, shouldError func(name string) bool) afero.Fs {
	t.Helper()
	return aferomock.OverrideFs(base, aferomock.FsCallbacks{
		OpenFunc: func(name string) (afero.File, error) {
			if shouldError(name) {
				return nil, errors.New("simulated fs error")
			}
			return base.Open(name)
		},
	})
}

func TestLoader_loadMetadata(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupFs    func(fs afero.Fs)
		wantErr    error
		errContain string
	}{
		{
			name: "valid metadata with full AI config",
			setupFs: func(fs afero.Fs) {
				content := `ai:
  instruction:
    sources:
      - uri: rulebook://ai/instructions
  skill:
    sources:
      - uri: rulebook://ai/skills
  workflow:
    sources:
      - uri: rulebook://ai/workflows`
				_ = afero.WriteFile(fs, rulebookFileName, []byte(content), 0644)
			},
			wantErr: nil,
		},
		{
			name: "valid metadata with partial AI config",
			setupFs: func(fs afero.Fs) {
				content := `ai:
  skill:
    sources:
      - uri: rulebook://ai/skills`
				_ = afero.WriteFile(fs, rulebookFileName, []byte(content), 0644)
			},
			wantErr: nil,
		},
		{
			name: "valid metadata with mcp sources",
			setupFs: func(fs afero.Fs) {
				content := `ai:
  mcp:
    sources:
      - uri: rulebook://ai/mcp`
				_ = afero.WriteFile(fs, rulebookFileName, []byte(content), 0644)
			},
			wantErr: nil,
		},
		{
			name: "missing file",
			setupFs: func(fs afero.Fs) {
			},
			wantErr: ErrMissingMetadataFile,
		},
		{
			name: "invalid YAML",
			setupFs: func(fs afero.Fs) {
				content := `ai:
  invalid: [unclosed array`
				_ = afero.WriteFile(fs, rulebookFileName, []byte(content), 0644)
			},
			wantErr:    ErrParseFailed,
			errContain: "parse failed",
		},
		{
			name: "read error",
			setupFs: func(fs afero.Fs) {
			},
			wantErr:    ErrReadFailed,
			errContain: "read failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fs := afero.NewMemMapFs()
			if tt.setupFs != nil {
				tt.setupFs(fs)
			}

			if tt.name == "read error" {
				_ = afero.WriteFile(fs, rulebookFileName, []byte("content"), 0644)
				fs = fsWithOpenError(t, fs, func(name string) bool {
					return name == rulebookFileName
				})
			}

			loader := NewLoader(fs)
			metadata, err := loader.loadMetadata()

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				if tt.errContain != "" {
					assert.Contains(t, err.Error(), tt.errContain)
				}
				assert.Nil(t, metadata)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, metadata)
			assert.NotNil(t, metadata.AI)
		})
	}
}

func TestLoader_loadMetadata_ValidatesStructure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		yamlContent  string
		validateFunc func(t *testing.T, metadata *Metadata)
	}{
		{
			name: "metadata with instruction sources",
			yamlContent: `ai:
  instruction:
    sources:
      - uri: rulebook://ai/instructions`,
			validateFunc: func(t *testing.T, metadata *Metadata) {
				require.NotNil(t, metadata.AI)
				require.NotNil(t, metadata.AI.Instruction)
				require.Len(t, metadata.AI.Instruction.Sources, 1)
				assert.Equal(t, "rulebook://ai/instructions", metadata.AI.Instruction.Sources[0].URI)
			},
		},
		{
			name: "metadata with skill sources",
			yamlContent: `ai:
  skill:
    sources:
      - uri: rulebook://ai/skills`,
			validateFunc: func(t *testing.T, metadata *Metadata) {
				require.NotNil(t, metadata.AI)
				require.NotNil(t, metadata.AI.Skill)
				require.Len(t, metadata.AI.Skill.Sources, 1)
				assert.Equal(t, "rulebook://ai/skills", metadata.AI.Skill.Sources[0].URI)
			},
		},
		{
			name: "metadata with multiple sources",
			yamlContent: `ai:
  instruction:
    sources:
      - uri: rulebook://ai/instructions
      - uri: file://./custom-instructions
  skill:
    sources:
      - uri: rulebook://ai/skills
      - uri: file://./custom-skills`,
			validateFunc: func(t *testing.T, metadata *Metadata) {
				require.NotNil(t, metadata.AI)
				require.NotNil(t, metadata.AI.Instruction)
				require.Len(t, metadata.AI.Instruction.Sources, 2)
				assert.Equal(t, "rulebook://ai/instructions", metadata.AI.Instruction.Sources[0].URI)
				assert.Equal(t, "file://./custom-instructions", metadata.AI.Instruction.Sources[1].URI)

				require.NotNil(t, metadata.AI.Skill)
				require.Len(t, metadata.AI.Skill.Sources, 2)
				assert.Equal(t, "rulebook://ai/skills", metadata.AI.Skill.Sources[0].URI)
				assert.Equal(t, "file://./custom-skills", metadata.AI.Skill.Sources[1].URI)
			},
		},
		{
			name: "metadata with doc standard sources",
			yamlContent: `doc:
  standard:
    sources:
      - uri: rulebook://docs/standards/golang`,
			validateFunc: func(t *testing.T, metadata *Metadata) {
				require.NotNil(t, metadata.Doc)
				require.NotNil(t, metadata.Doc.Standard)
				require.Len(t, metadata.Doc.Standard.Sources, 1)
				assert.Equal(t, "rulebook://docs/standards/golang", metadata.Doc.Standard.Sources[0].URI)
			},
		},
		{
			name: "metadata with workflow sources",
			yamlContent: `ai:
  workflow:
    sources:
      - uri: rulebook://ai/workflows`,
			validateFunc: func(t *testing.T, metadata *Metadata) {
				require.NotNil(t, metadata.AI)
				require.NotNil(t, metadata.AI.Workflows)
				require.Len(t, metadata.AI.Workflows.Sources, 1)
				assert.Equal(t, "rulebook://ai/workflows", metadata.AI.Workflows.Sources[0].URI)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fs := afero.NewMemMapFs()
			_ = afero.WriteFile(fs, rulebookFileName, []byte(tt.yamlContent), 0644)

			loader := NewLoader(fs)
			metadata, err := loader.loadMetadata()

			require.NoError(t, err)
			require.NotNil(t, metadata)
			tt.validateFunc(t, metadata)
		})
	}
}

func TestLoader_validateMetadata(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		metadata Metadata
		wantErr  error
	}{
		{
			name: "valid full config",
			metadata: Metadata{
				AI: &ai.Config{
					Instruction: &instruction.Config{
						Sources: []instruction.SourceConfig{
							{URI: "rulebook://ai/instructions"},
						},
					},
					Skill: &skill.Config{
						Sources: []skill.SourceConfig{
							{URI: "rulebook://ai/skills"},
						},
					},
					Workflows: &workflow.Config{
						Sources: []workflow.SourceConfig{
							{URI: "rulebook://ai/workflows"},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "valid partial config",
			metadata: Metadata{
				AI: &ai.Config{
					Skill: &skill.Config{
						Sources: []skill.SourceConfig{
							{URI: "rulebook://ai/skills"},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "nil AI field",
			metadata: Metadata{
				AI: nil,
			},
			wantErr: nil,
		},
		{
			name:     "empty Metadata",
			metadata: Metadata{},
			wantErr:  nil,
		},
		{
			name: "invalid instruction URI",
			metadata: Metadata{
				AI: &ai.Config{
					Instruction: &instruction.Config{
						Sources: []instruction.SourceConfig{
							{URI: ""},
						},
					},
				},
			},
			wantErr: ErrValidationFailed,
		},
		{
			name: "empty instruction sources",
			metadata: Metadata{
				AI: &ai.Config{
					Instruction: &instruction.Config{
						Sources: []instruction.SourceConfig{},
					},
				},
			},
			wantErr: ErrValidationFailed,
		},
		{
			name: "valid mcp config",
			metadata: Metadata{
				AI: &ai.Config{
					MCP: &mcp.Config{
						Sources: []mcp.SourceConfig{
							{URI: "rulebook://ai/mcp"},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "invalid mcp - empty sources",
			metadata: Metadata{
				AI: &ai.Config{
					MCP: &mcp.Config{
						Sources: []mcp.SourceConfig{},
					},
				},
			},
			wantErr: ErrValidationFailed,
		},
		{
			name: "valid doc config",
			metadata: Metadata{
				Doc: &doc.Config{
					Standard: &standard.Config{
						Sources: []standard.SourceConfig{
							{URI: "rulebook://docs/standards/golang"},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "invalid doc config - empty sources",
			metadata: Metadata{
				Doc: &doc.Config{
					Standard: &standard.Config{
						Sources: []standard.SourceConfig{},
					},
				},
			},
			wantErr: ErrValidationFailed,
		},
		{
			name: "valid workflow config",
			metadata: Metadata{
				AI: &ai.Config{
					Workflows: &workflow.Config{
						Sources: []workflow.SourceConfig{
							{URI: "rulebook://ai/workflows"},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "invalid workflow config - empty sources",
			metadata: Metadata{
				AI: &ai.Config{
					Workflows: &workflow.Config{
						Sources: []workflow.SourceConfig{},
					},
				},
			},
			wantErr: ErrValidationFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			loader := NewLoader(afero.NewMemMapFs())
			err := loader.validateMetadata(tt.metadata)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestLoader_resolveSourceUri(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		uri         string
		wantPath    string
		wantErr     error
		errContains string
	}{
		{
			name:     "valid simple path",
			uri:      "rulebook://ai/skills",
			wantPath: "/ai/skills",
			wantErr:  nil,
		},
		{
			name:     "valid nested path",
			uri:      "rulebook://ai/skills/git-commit",
			wantPath: "/ai/skills/git-commit",
			wantErr:  nil,
		},
		{
			name:        "empty path after scheme",
			uri:         "rulebook://",
			wantPath:    "",
			wantErr:     ErrEmptyPath,
			errContains: "empty path",
		},
		{
			name:        "different scheme",
			uri:         "file://./custom",
			wantPath:    "",
			wantErr:     ErrUnsupportedScheme,
			errContains: "unsupported scheme",
		},
		{
			name:        "no scheme separator",
			uri:         "foobar",
			wantPath:    "",
			wantErr:     ErrUnsupportedScheme,
			errContains: "unsupported scheme",
		},
		{
			name:        "empty string",
			uri:         "",
			wantPath:    "",
			wantErr:     ErrUnsupportedScheme,
			errContains: "unsupported scheme",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			loader := NewLoader(afero.NewMemMapFs())
			gotPath, err := loader.resolveSourceUri(tt.uri)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Equal(t, tt.wantPath, gotPath)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantPath, gotPath)
		})
	}
}

func TestLoader_Load(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupFs    func(fs afero.Fs)
		wantErr    bool
		errContain string
		validate   func(t *testing.T, rb *Rulebook)
	}{
		{
			name: "instructions and skills loaded",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, rulebookFileName, []byte(`ai:
  instruction:
    sources:
      - uri: rulebook://ai/instructions
  skill:
    sources:
      - uri: rulebook://ai/skills`), 0644)

				_ = fs.MkdirAll("/ai/instructions", 0755)
				_ = afero.WriteFile(fs, "/ai/instructions/01-coding.yaml", []byte(`category: "coding"
rules:
  - "write clean code"`), 0644)

				_ = fs.MkdirAll("/ai/skills/my-skill", 0755)
				_ = afero.WriteFile(fs, "/ai/skills/my-skill/metadata.yaml", []byte(`name: "my-skill"
description: "A test skill"`), 0644)
				_ = afero.WriteFile(fs, "/ai/skills/my-skill/instructions.md", []byte("Do something useful."), 0644)
			},
			wantErr: false,
			validate: func(t *testing.T, rb *Rulebook) {
				require.NotNil(t, rb)
				assert.Len(t, rb.AI.Instructions, 1)
				assert.Equal(t, "coding", string(rb.AI.Instructions[0].Category))
				assert.Len(t, rb.AI.Instructions[0].Rules, 1)
				assert.Len(t, rb.AI.Skills, 1)
				assert.Equal(t, "my-skill", string(rb.AI.Skills[0].Metadata.Name))
			},
		},
		{
			name: "only instructions",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, rulebookFileName, []byte(`ai:
  instruction:
    sources:
      - uri: rulebook://ai/instructions`), 0644)

				_ = fs.MkdirAll("/ai/instructions", 0755)
				_ = afero.WriteFile(fs, "/ai/instructions/01-coding.yaml", []byte(`category: "coding"
rules:
  - "write clean code"`), 0644)
			},
			wantErr: false,
			validate: func(t *testing.T, rb *Rulebook) {
				require.NotNil(t, rb)
				assert.Len(t, rb.AI.Instructions, 1)
				assert.Empty(t, rb.AI.Skills)
			},
		},
		{
			name: "only skills",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, rulebookFileName, []byte(`ai:
  skill:
    sources:
      - uri: rulebook://ai/skills`), 0644)

				_ = fs.MkdirAll("/ai/skills/my-skill", 0755)
				_ = afero.WriteFile(fs, "/ai/skills/my-skill/metadata.yaml", []byte(`name: "my-skill"
description: "A test skill"`), 0644)
				_ = afero.WriteFile(fs, "/ai/skills/my-skill/instructions.md", []byte("Do something useful."), 0644)
			},
			wantErr: false,
			validate: func(t *testing.T, rb *Rulebook) {
				require.NotNil(t, rb)
				assert.Empty(t, rb.AI.Instructions)
				assert.Len(t, rb.AI.Skills, 1)
			},
		},
		{
			name: "multiple instruction sources",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, rulebookFileName, []byte(`ai:
  instruction:
    sources:
      - uri: rulebook://ai/instructions1
      - uri: rulebook://ai/instructions2`), 0644)

				_ = fs.MkdirAll("/ai/instructions1", 0755)
				_ = afero.WriteFile(fs, "/ai/instructions1/01-coding.yaml", []byte(`category: "coding"
rules:
  - "write clean code"`), 0644)

				_ = fs.MkdirAll("/ai/instructions2", 0755)
				_ = afero.WriteFile(fs, "/ai/instructions2/02-testing.yaml", []byte(`category: "testing"
rules:
  - "write tests"`), 0644)
			},
			wantErr: false,
			validate: func(t *testing.T, rb *Rulebook) {
				require.NotNil(t, rb)
				assert.Len(t, rb.AI.Instructions, 2)
				assert.Equal(t, "coding", string(rb.AI.Instructions[0].Category))
				assert.Equal(t, "testing", string(rb.AI.Instructions[1].Category))
			},
		},
		{
			name: "missing metadata file",
			setupFs: func(fs afero.Fs) {
			},
			wantErr:    true,
			errContain: "load metadata",
			validate: func(t *testing.T, rb *Rulebook) {
				assert.Nil(t, rb)
			},
		},
		{
			name: "metadata validation fails",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, rulebookFileName, []byte(`ai:
  instruction:
    sources: []`), 0644)
			},
			wantErr:    true,
			errContain: "validate metadata",
			validate: func(t *testing.T, rb *Rulebook) {
				assert.Nil(t, rb)
			},
		},
		{
			name: "instruction URI resolution fails",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, rulebookFileName, []byte(`ai:
  instruction:
    sources:
      - uri: file://invalid/path`), 0644)
			},
			wantErr:    true,
			errContain: "ai instructions: resolve source path",
			validate: func(t *testing.T, rb *Rulebook) {
				assert.Nil(t, rb)
			},
		},
		{
			name: "skill URI resolution fails",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, rulebookFileName, []byte(`ai:
  skill:
    sources:
      - uri: http://invalid/scheme`), 0644)
			},
			wantErr:    true,
			errContain: "ai skills: resolve source path",
			validate: func(t *testing.T, rb *Rulebook) {
				assert.Nil(t, rb)
			},
		},
		{
			name: "instruction loading fails",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, rulebookFileName, []byte(`ai:
  instruction:
    sources:
      - uri: rulebook://ai/instructions
  skill:
    sources:
      - uri: rulebook://ai/skills`), 0644)

				_ = fs.MkdirAll("/ai/instructions", 0755)

				_ = fs.MkdirAll("/ai/skills/my-skill", 0755)
				_ = afero.WriteFile(fs, "/ai/skills/my-skill/metadata.yaml", []byte(`name: "my-skill"
description: "A test skill"`), 0644)
				_ = afero.WriteFile(fs, "/ai/skills/my-skill/instructions.md", []byte("Do something useful."), 0644)
			},
			wantErr:    true,
			errContain: "ai instructions: load ai instructions",
			validate: func(t *testing.T, rb *Rulebook) {
				assert.Nil(t, rb)
			},
		},
		{
			name: "skill loading fails",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, rulebookFileName, []byte(`ai:
  instruction:
    sources:
      - uri: rulebook://ai/instructions
  skill:
    sources:
      - uri: rulebook://ai/skills`), 0644)

				_ = fs.MkdirAll("/ai/instructions", 0755)
				_ = afero.WriteFile(fs, "/ai/instructions/01-coding.yaml", []byte(`category: "coding"
rules:
  - "write clean code"`), 0644)

				_ = fs.MkdirAll("/ai/skills", 0755)
			},
			wantErr:    true,
			errContain: "ai skills: load ai skills",
			validate: func(t *testing.T, rb *Rulebook) {
				assert.Nil(t, rb)
			},
		},
		{
			name: "only standards",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, rulebookFileName, []byte(`doc:
  standard:
    sources:
      - uri: rulebook://docs/standards`), 0644)

				_ = fs.MkdirAll("/docs/standards", 0755)
				_ = afero.WriteFile(fs, "/docs/standards/logging.yaml", []byte(`metadata:
  id: logging
  name: logging
  version: 0.1.0
  tags:
    - logging
  scope:
    languages:
      - go
  relations:
    standard: []
specification:
  purpose: Logging standard for testing purposes only
  goals:
    - use structured logging
requirements:
  rules:
    - level: must
      statement: Use structured logging for all log messages
      rationale: Structured logs are easier to parse and analyze
examples:
  good:
    - title: Structured logging example
      language: go
      snippet: log.Info("message")
      reason: Uses structured logging`), 0644)
			},
			wantErr: false,
			validate: func(t *testing.T, rb *Rulebook) {
				require.NotNil(t, rb)
				assert.Empty(t, rb.AI.Instructions)
				assert.Empty(t, rb.AI.Skills)
				assert.Len(t, rb.Doc.Standards, 1)
				assert.Equal(t, "logging", rb.Doc.Standards[0].Metadata.Name)
			},
		},
		{
			name: "instructions, skills and standards",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, rulebookFileName, []byte(`ai:
  instruction:
    sources:
      - uri: rulebook://ai/instructions
  skill:
    sources:
      - uri: rulebook://ai/skills
doc:
  standard:
    sources:
      - uri: rulebook://docs/standards`), 0644)

				_ = fs.MkdirAll("/ai/instructions", 0755)
				_ = afero.WriteFile(fs, "/ai/instructions/01-coding.yaml", []byte(`category: "coding"
rules:
  - "write clean code"`), 0644)

				_ = fs.MkdirAll("/ai/skills/my-skill", 0755)
				_ = afero.WriteFile(fs, "/ai/skills/my-skill/metadata.yaml", []byte(`name: "my-skill"
description: "A test skill"`), 0644)
				_ = afero.WriteFile(fs, "/ai/skills/my-skill/instructions.md", []byte("Do something useful."), 0644)

				_ = fs.MkdirAll("/docs/standards", 0755)
				_ = afero.WriteFile(fs, "/docs/standards/logging.yaml", []byte(`metadata:
  id: logging
  name: logging
  version: 0.1.0
  tags:
    - logging
  scope:
    languages:
      - go
  relations:
    standard: []
specification:
  purpose: Logging standard for testing purposes only
  goals:
    - use structured logging
requirements:
  rules:
    - level: must
      statement: Use structured logging for all log messages
      rationale: Structured logs are easier to parse and analyze
examples:
  good:
    - title: Structured logging example
      language: go
      snippet: log.Info("message")
      reason: Uses structured logging`), 0644)
			},
			wantErr: false,
			validate: func(t *testing.T, rb *Rulebook) {
				require.NotNil(t, rb)
				assert.Len(t, rb.AI.Instructions, 1)
				assert.Len(t, rb.AI.Skills, 1)
				assert.Len(t, rb.Doc.Standards, 1)
				assert.Equal(t, "logging", rb.Doc.Standards[0].Metadata.Name)
			},
		},
		{
			name: "standard URI resolution fails",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, rulebookFileName, []byte(`doc:
  standard:
    sources:
      - uri: http://invalid/scheme`), 0644)
			},
			wantErr:    true,
			errContain: "doc standards: resolve source path",
			validate: func(t *testing.T, rb *Rulebook) {
				assert.Nil(t, rb)
			},
		},
		{
			name: "standard loading fails",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, rulebookFileName, []byte(`doc:
  standard:
    sources:
      - uri: rulebook://docs/standards`), 0644)

				_ = fs.MkdirAll("/docs/standards", 0755)
			},
			wantErr:    true,
			errContain: "doc standards: load doc standards",
			validate: func(t *testing.T, rb *Rulebook) {
				assert.Nil(t, rb)
			},
		},
		{
			name: "only workflows",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, rulebookFileName, []byte(`ai:
  workflow:
    sources:
      - uri: rulebook://ai/workflows`), 0644)

				_ = fs.MkdirAll("/ai/workflows", 0755)
				_ = afero.WriteFile(fs, "/ai/workflows/test-workflow.yaml", []byte(`metadata:
  id: test-workflow
  name: Test Workflow
  description: A test workflow
  version: 1.0.0
steps:
  - id: step1
    name: Step 1
    description: First step
    instructions:
      - Do something`), 0644)
			},
			wantErr: false,
			validate: func(t *testing.T, rb *Rulebook) {
				require.NotNil(t, rb)
				assert.Empty(t, rb.AI.Instructions)
				assert.Empty(t, rb.AI.Skills)
				assert.Len(t, rb.AI.Workflows, 1)
				assert.Equal(t, workflow.WorkflowId("test-workflow"), rb.AI.Workflows[0].Metadata.ID)
			},
		},
		{
			name: "only mcp servers",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, rulebookFileName, []byte(`ai:
  mcp:
    sources:
      - uri: rulebook://ai/mcp`), 0644)

				_ = fs.MkdirAll("/ai/mcp", 0755)
				_ = afero.WriteFile(fs, "/ai/mcp/test-server.yaml", []byte(`name: "test-server"
stdio:
  executablePath: "/usr/local/bin/test"`), 0644)
			},
			wantErr: false,
			validate: func(t *testing.T, rb *Rulebook) {
				require.NotNil(t, rb)
				assert.Empty(t, rb.AI.Instructions)
				assert.Empty(t, rb.AI.Skills)
				assert.Empty(t, rb.AI.Workflows)
				assert.Len(t, rb.AI.MCPServers, 1)
				assert.Equal(t, "test-server", rb.AI.MCPServers[0].Name)
			},
		},
		{
			name: "mcp URI resolution fails",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, rulebookFileName, []byte(`ai:
  mcp:
    sources:
      - uri: http://invalid/scheme`), 0644)
			},
			wantErr:    true,
			errContain: "ai mcp: resolve source path",
			validate: func(t *testing.T, rb *Rulebook) {
				assert.Nil(t, rb)
			},
		},
		{
			name: "mcp loading fails",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, rulebookFileName, []byte(`ai:
  mcp:
    sources:
      - uri: rulebook://ai/mcp`), 0644)

				_ = fs.MkdirAll("/ai/mcp", 0755)
			},
			wantErr:    true,
			errContain: "ai mcp: load ai mcp servers",
			validate: func(t *testing.T, rb *Rulebook) {
				assert.Nil(t, rb)
			},
		},
		{
			name: "workflow URI resolution fails",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, rulebookFileName, []byte(`ai:
  workflow:
    sources:
      - uri: http://invalid/scheme`), 0644)
			},
			wantErr:    true,
			errContain: "ai workflows: resolve source path",
			validate: func(t *testing.T, rb *Rulebook) {
				assert.Nil(t, rb)
			},
		},
		{
			name: "workflow loading fails",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, rulebookFileName, []byte(`ai:
  workflow:
    sources:
      - uri: rulebook://ai/workflows`), 0644)

				_ = fs.MkdirAll("/ai/workflows", 0755)
			},
			wantErr:    true,
			errContain: "ai workflows: load ai workflows",
			validate: func(t *testing.T, rb *Rulebook) {
				assert.Nil(t, rb)
			},
		},
		{
			name: "all components loaded",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, rulebookFileName, []byte(`ai:
  instruction:
    sources:
      - uri: rulebook://ai/instructions
  skill:
    sources:
      - uri: rulebook://ai/skills
  workflow:
    sources:
      - uri: rulebook://ai/workflows
  mcp:
    sources:
      - uri: rulebook://ai/mcp
doc:
  standard:
    sources:
      - uri: rulebook://docs/standards`), 0644)

				_ = fs.MkdirAll("/ai/instructions", 0755)
				_ = afero.WriteFile(fs, "/ai/instructions/01-coding.yaml", []byte(`category: "coding"
rules:
  - "write clean code"`), 0644)

				_ = fs.MkdirAll("/ai/skills/my-skill", 0755)
				_ = afero.WriteFile(fs, "/ai/skills/my-skill/metadata.yaml", []byte(`name: "my-skill"
description: "A test skill"`), 0644)
				_ = afero.WriteFile(fs, "/ai/skills/my-skill/instructions.md", []byte("Do something useful."), 0644)

				_ = fs.MkdirAll("/ai/workflows", 0755)
				_ = afero.WriteFile(fs, "/ai/workflows/test-workflow.yaml", []byte(`metadata:
  id: test-workflow
  name: Test Workflow
  description: A test workflow
  version: 1.0.0
steps:
  - id: step1
    name: Step 1
    description: First step
    instructions:
      - Do something`), 0644)

				_ = fs.MkdirAll("/ai/mcp", 0755)
				_ = afero.WriteFile(fs, "/ai/mcp/test-server.yaml", []byte(`name: "test-server"
stdio:
  executablePath: "/usr/local/bin/test"`), 0644)

				_ = fs.MkdirAll("/docs/standards", 0755)
				_ = afero.WriteFile(fs, "/docs/standards/logging.yaml", []byte(`metadata:
  id: logging
  name: logging
  version: 0.1.0
  tags:
    - logging
  scope:
    languages:
      - go
  relations:
    standard: []
specification:
  purpose: Logging standard for testing purposes only
  goals:
    - use structured logging
requirements:
  rules:
    - level: must
      statement: Use structured logging for all log messages
      rationale: Structured logs are easier to parse and analyze
examples:
  good:
    - title: Structured logging example
      language: go
      snippet: log.Info("message")
      reason: Uses structured logging`), 0644)
			},
			wantErr: false,
			validate: func(t *testing.T, rb *Rulebook) {
				require.NotNil(t, rb)
				assert.Len(t, rb.AI.Instructions, 1)
				assert.Len(t, rb.AI.Skills, 1)
				assert.Len(t, rb.AI.Workflows, 1)
				assert.Len(t, rb.AI.MCPServers, 1)
				assert.Len(t, rb.Doc.Standards, 1)
				assert.Equal(t, workflow.WorkflowId("test-workflow"), rb.AI.Workflows[0].Metadata.ID)
				assert.Equal(t, "logging", rb.Doc.Standards[0].Metadata.Name)
			},
		},
		{
			name: "multiple workflow sources",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, rulebookFileName, []byte(`ai:
  workflow:
    sources:
      - uri: rulebook://ai/workflows1
      - uri: rulebook://ai/workflows2`), 0644)

				_ = fs.MkdirAll("/ai/workflows1", 0755)
				_ = afero.WriteFile(fs, "/ai/workflows1/workflow1.yaml", []byte(`metadata:
  id: workflow1
  name: Workflow 1
  description: First workflow
  version: 1.0.0
steps:
  - id: step1
    name: Step 1
    description: First step
    instructions:
      - Do something`), 0644)

				_ = fs.MkdirAll("/ai/workflows2", 0755)
				_ = afero.WriteFile(fs, "/ai/workflows2/workflow2.yaml", []byte(`metadata:
  id: workflow2
  name: Workflow 2
  description: Second workflow
  version: 1.0.0
steps:
  - id: step1
    name: Step 1
    description: First step
    instructions:
      - Do something else`), 0644)
			},
			wantErr: false,
			validate: func(t *testing.T, rb *Rulebook) {
				require.NotNil(t, rb)
				assert.Len(t, rb.AI.Workflows, 2)
				assert.Equal(t, workflow.WorkflowId("workflow1"), rb.AI.Workflows[0].Metadata.ID)
				assert.Equal(t, workflow.WorkflowId("workflow2"), rb.AI.Workflows[1].Metadata.ID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fs := afero.NewMemMapFs()
			if tt.setupFs != nil {
				tt.setupFs(fs)
			}

			loader := NewLoader(fs)
			rb, err := loader.Load()

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContain != "" {
					assert.Contains(t, err.Error(), tt.errContain)
				}
			} else {
				require.NoError(t, err)
			}

			if tt.validate != nil {
				tt.validate(t, rb)
			}
		})
	}
}
