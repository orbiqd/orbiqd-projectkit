package workflow

import (
	"errors"
	"testing"

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

func TestLoader_resolvePaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupFs    func(fs afero.Fs)
		wantLen    int
		wantErr    error
		errContain string
	}{
		{
			name: "single yaml file",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "workflow1.yaml", []byte("content"), 0644)
			},
			wantLen: 1,
			wantErr: nil,
		},
		{
			name: "single yml file",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "workflow1.yml", []byte("content"), 0644)
			},
			wantLen: 1,
			wantErr: nil,
		},
		{
			name: "multiple yaml files",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "workflow1.yaml", []byte("content"), 0644)
				_ = afero.WriteFile(fs, "workflow2.yaml", []byte("content"), 0644)
				_ = afero.WriteFile(fs, "workflow3.yml", []byte("content"), 0644)
			},
			wantLen: 3,
			wantErr: nil,
		},
		{
			name: "empty filesystem",
			setupFs: func(fs afero.Fs) {
			},
			wantLen: 0,
			wantErr: nil,
		},
		{
			name: "only non-yaml files",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "file1.txt", []byte("content"), 0644)
				_ = afero.WriteFile(fs, "file2.json", []byte("content"), 0644)
			},
			wantLen: 0,
			wantErr: nil,
		},
		{
			name: "mix of yaml and non-yaml files",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "workflow1.yaml", []byte("content"), 0644)
				_ = afero.WriteFile(fs, "workflow2.yml", []byte("content"), 0644)
				_ = afero.WriteFile(fs, "file1.txt", []byte("content"), 0644)
				_ = afero.WriteFile(fs, "file2.json", []byte("content"), 0644)
			},
			wantLen: 2,
			wantErr: nil,
		},
		{
			name: "directories are ignored",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "workflow1.yaml", []byte("content"), 0644)
				_ = fs.Mkdir("subdir", 0755)
			},
			wantLen: 1,
			wantErr: nil,
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
			got, err := loader.resolvePaths()

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else if tt.errContain != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContain)
			} else {
				require.NoError(t, err)
				assert.Len(t, got, tt.wantLen)
			}
		})
	}
}

func TestLoader_resolvePaths_WhenReadDirectoryError_ThenReturnsError(t *testing.T) {
	t.Parallel()

	fs := fsWithOpenError(t, afero.NewMemMapFs(), func(string) bool { return true })
	loader := NewLoader(fs)

	got, err := loader.resolvePaths()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "read directory")
	assert.Len(t, got, 0)
}

func TestLoader_loadWorkflow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupFs    func(fs afero.Fs)
		path       string
		wantErr    error
		wantID     WorkflowId
		wantName   string
		errContain string
	}{
		{
			name: "valid workflow",
			setupFs: func(fs afero.Fs) {
				content := `metadata:
  id: test-workflow
  name: Test Workflow
  description: A test workflow
  version: 1.0.0
steps:
  - id: step1
    name: Step 1
    description: First step
    instructions:
      - Do something`
				_ = afero.WriteFile(fs, "workflow.yaml", []byte(content), 0644)
			},
			path:     "workflow.yaml",
			wantErr:  nil,
			wantID:   WorkflowId("test-workflow"),
			wantName: "Test Workflow",
		},
		{
			name: "file not found",
			setupFs: func(fs afero.Fs) {
			},
			path:       "nonexistent.yaml",
			errContain: "read failed",
		},
		{
			name: "invalid yaml",
			setupFs: func(fs afero.Fs) {
				content := `invalid: yaml: content: {{{`
				_ = afero.WriteFile(fs, "workflow.yaml", []byte(content), 0644)
			},
			path:       "workflow.yaml",
			errContain: "parse failed",
		},
		{
			name: "missing required fields",
			setupFs: func(fs afero.Fs) {
				content := `metadata:
  id: test-workflow`
				_ = afero.WriteFile(fs, "workflow.yaml", []byte(content), 0644)
			},
			path:       "workflow.yaml",
			errContain: "validation failed",
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
			got, err := loader.loadWorkflow(tt.path)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else if tt.errContain != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContain)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, tt.wantID, got.Metadata.ID)
				assert.Equal(t, tt.wantName, got.Metadata.Name)
			}
		})
	}
}

func TestLoader_Load(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupFs    func(fs afero.Fs)
		wantLen    int
		wantErr    error
		errContain string
	}{
		{
			name: "single workflow",
			setupFs: func(fs afero.Fs) {
				content := `metadata:
  id: workflow1
  name: Workflow 1
  description: First workflow
  version: 1.0.0
steps:
  - id: step1
    name: Step 1
    description: First step
    instructions:
      - Do something`
				_ = afero.WriteFile(fs, "workflow1.yaml", []byte(content), 0644)
			},
			wantLen: 1,
			wantErr: nil,
		},
		{
			name: "multiple workflows",
			setupFs: func(fs afero.Fs) {
				content1 := `metadata:
  id: workflow1
  name: Workflow 1
  description: First workflow
  version: 1.0.0
steps:
  - id: step1
    name: Step 1
    description: First step
    instructions:
      - Do something`
				content2 := `metadata:
  id: workflow2
  name: Workflow 2
  description: Second workflow
  version: 1.0.0
steps:
  - id: step1
    name: Step 1
    description: First step
    instructions:
      - Do something else`
				_ = afero.WriteFile(fs, "workflow1.yaml", []byte(content1), 0644)
				_ = afero.WriteFile(fs, "workflow2.yaml", []byte(content2), 0644)
			},
			wantLen: 2,
			wantErr: nil,
		},
		{
			name: "empty filesystem",
			setupFs: func(fs afero.Fs) {
			},
			wantErr: ErrNoWorkflowsFound,
		},
		{
			name: "no yaml files",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "file1.txt", []byte("content"), 0644)
			},
			wantErr: ErrNoWorkflowsFound,
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
			got, err := loader.Load()

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else if tt.errContain != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContain)
			} else {
				require.NoError(t, err)
				assert.Len(t, got, tt.wantLen)
			}
		})
	}
}

func TestLoader_Load_WhenInvalidWorkflow_ThenReturnsError(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	content := `invalid: yaml: content: {{{`
	_ = afero.WriteFile(fs, "workflow.yaml", []byte(content), 0644)

	loader := NewLoader(fs)
	got, err := loader.Load()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load workflow")
	assert.Empty(t, got)
}

