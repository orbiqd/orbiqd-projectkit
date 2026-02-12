package loader

import (
	"errors"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sourceAPI "github.com/orbiqd/orbiqd-projectkit/pkg/source"
	workflowAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/workflow"
)

func validWorkflowFs(t *testing.T, id, name, description string) afero.Fs {
	t.Helper()

	fs := afero.NewMemMapFs()

	workflow := `metadata:
  id: ` + id + `
  name: ` + name + `
  description: ` + description + `
  version: 1.0.0
steps:
  - id: step1
    name: Step 1
    description: First step
    instructions:
      - Do something
`
	require.NoError(t, afero.WriteFile(fs, id+".yaml", []byte(workflow), 0644))

	return fs
}

func TestLoadWorkflowsFromConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		sources          []workflowAPI.SourceConfig
		mockSetup        func(*sourceAPI.MockResolver)
		wantWorkflowsLen int
	}{
		{
			name:             "WhenNoSources_ThenReturnsNilWorkflows",
			sources:          []workflowAPI.SourceConfig{},
			mockSetup:        func(m *sourceAPI.MockResolver) {},
			wantWorkflowsLen: 0,
		},
		{
			name: "WhenSingleSourceWithOneWorkflow_ThenReturnsWorkflow",
			sources: []workflowAPI.SourceConfig{
				{URI: "file://./workflows"},
			},
			mockSetup: func(m *sourceAPI.MockResolver) {
				fs := validWorkflowFs(t, "test-workflow", "Test Workflow", "Test workflow description")
				m.EXPECT().Resolve("file://./workflows").Return(fs, nil)
			},
			wantWorkflowsLen: 1,
		},
		{
			name: "WhenMultipleSources_ThenReturnsCombinedWorkflows",
			sources: []workflowAPI.SourceConfig{
				{URI: "file://./workflows1"},
				{URI: "file://./workflows2"},
			},
			mockSetup: func(m *sourceAPI.MockResolver) {
				fs1 := validWorkflowFs(t, "workflow-one", "Workflow One", "First workflow")
				fs2 := validWorkflowFs(t, "workflow-two", "Workflow Two", "Second workflow")
				m.EXPECT().Resolve("file://./workflows1").Return(fs1, nil)
				m.EXPECT().Resolve("file://./workflows2").Return(fs2, nil)
			},
			wantWorkflowsLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockResolver := sourceAPI.NewMockResolver(t)
			tt.mockSetup(mockResolver)

			config := workflowAPI.Config{
				Sources: tt.sources,
			}

			workflows, err := LoadWorkflowsFromConfig(config, mockResolver)

			require.NoError(t, err)
			assert.Len(t, workflows, tt.wantWorkflowsLen)
		})
	}
}

func TestLoadWorkflowsFromConfig_WhenResolverFails_ThenReturnsResolveError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	resolveErr := errors.New("resolver failed")
	mockResolver.EXPECT().Resolve("file://./workflows").Return(nil, resolveErr)

	config := workflowAPI.Config{
		Sources: []workflowAPI.SourceConfig{
			{URI: "file://./workflows"},
		},
	}

	workflows, err := LoadWorkflowsFromConfig(config, mockResolver)

	require.Error(t, err)
	assert.Empty(t, workflows)
	assert.Contains(t, err.Error(), "resolve: file://./workflows")
	assert.ErrorIs(t, err, resolveErr)
}

func TestLoadWorkflowsFromConfig_WhenLoaderFails_ThenReturnsLoadWorkflowsError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	emptyFs := afero.NewMemMapFs()
	mockResolver.EXPECT().Resolve("file://./workflows").Return(emptyFs, nil)

	config := workflowAPI.Config{
		Sources: []workflowAPI.SourceConfig{
			{URI: "file://./workflows"},
		},
	}

	workflows, err := LoadWorkflowsFromConfig(config, mockResolver)

	require.Error(t, err)
	assert.Empty(t, workflows)
	assert.Contains(t, err.Error(), "load workflows:")
	assert.ErrorIs(t, err, workflowAPI.ErrNoWorkflowsFound)
}

func TestLoadWorkflowsFromConfig_WhenSecondSourceResolverFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	fs1 := validWorkflowFs(t, "workflow-one", "Workflow One", "First workflow")
	resolveErr := errors.New("second resolver failed")

	mockResolver.EXPECT().Resolve("file://./workflows1").Return(fs1, nil)
	mockResolver.EXPECT().Resolve("file://./workflows2").Return(nil, resolveErr)

	config := workflowAPI.Config{
		Sources: []workflowAPI.SourceConfig{
			{URI: "file://./workflows1"},
			{URI: "file://./workflows2"},
		},
	}

	workflows, err := LoadWorkflowsFromConfig(config, mockResolver)

	require.Error(t, err)
	assert.Empty(t, workflows)
	assert.Contains(t, err.Error(), "resolve: file://./workflows2")
	assert.ErrorIs(t, err, resolveErr)
}

func TestLoadWorkflowsFromConfig_WhenSecondSourceLoaderFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	fs1 := validWorkflowFs(t, "workflow-one", "Workflow One", "First workflow")
	emptyFs := afero.NewMemMapFs()

	mockResolver.EXPECT().Resolve("file://./workflows1").Return(fs1, nil)
	mockResolver.EXPECT().Resolve("file://./workflows2").Return(emptyFs, nil)

	config := workflowAPI.Config{
		Sources: []workflowAPI.SourceConfig{
			{URI: "file://./workflows1"},
			{URI: "file://./workflows2"},
		},
	}

	workflows, err := LoadWorkflowsFromConfig(config, mockResolver)

	require.Error(t, err)
	assert.Empty(t, workflows)
	assert.Contains(t, err.Error(), "load workflows:")
	assert.ErrorIs(t, err, workflowAPI.ErrNoWorkflowsFound)
}
