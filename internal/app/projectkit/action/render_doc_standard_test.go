package action

import (
	"errors"
	"os"
	"testing"

	standardAPI "github.com/orbiqd/orbiqd-projectkit/pkg/doc/standard"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.nhat.io/aferomock"
)

func sampleStandard(name string) standardAPI.Standard {
	return standardAPI.Standard{
		Metadata: standardAPI.Metadata{
			Name:    name,
			Version: "1.0.0",
			Tags:    []string{"test"},
			Scope: standardAPI.ScopeMetadata{
				Languages: []string{"en"},
			},
		},
		Specification: standardAPI.Specification{
			Purpose: "Test purpose",
			Goals:   []string{"Test goal"},
		},
		Requirements: standardAPI.Requirements{
			Rules: []standardAPI.RequirementRule{
				{
					Level:     "must",
					Statement: "Test requirement",
					Rationale: "Test rationale",
				},
			},
		},
		Examples: standardAPI.Examples{
			Good: []standardAPI.Example{
				{
					Title:    "Good Example",
					Language: "go",
					Snippet:  "package main",
					Reason:   "Test reason",
				},
			},
		},
	}
}

func TestRenderDocStandardActionRun_WhenNoRenderConfigs_ThenReturnsNil(t *testing.T) {
	t.Parallel()

	mockRepo := standardAPI.NewMockRepository(t)
	mockRepo.EXPECT().GetAll().Return([]standardAPI.Standard{sampleStandard("Test Standard")}, nil)

	projectFs := afero.NewMemMapFs()
	action := NewRenderDocStandardAction(mockRepo, projectFs, []standardAPI.RenderConfig{})

	err := action.Run()

	require.NoError(t, err)
}

func TestRenderDocStandardActionRun_WhenNoStandards_ThenReturnsNil(t *testing.T) {
	t.Parallel()

	mockRepo := standardAPI.NewMockRepository(t)
	mockRepo.EXPECT().GetAll().Return([]standardAPI.Standard{}, nil)

	projectFs := afero.NewMemMapFs()
	renderConfigs := []standardAPI.RenderConfig{
		{Format: "markdown", Destination: "/docs"},
	}
	action := NewRenderDocStandardAction(mockRepo, projectFs, renderConfigs)

	err := action.Run()

	require.NoError(t, err)

	exists, _ := afero.Exists(projectFs, "/docs")
	assert.True(t, exists)
}

func TestRenderDocStandardActionRun_WhenSingleStandardMarkdown_ThenWritesFile(t *testing.T) {
	t.Parallel()

	standard := sampleStandard("Test Standard")
	mockRepo := standardAPI.NewMockRepository(t)
	mockRepo.EXPECT().GetAll().Return([]standardAPI.Standard{standard}, nil)

	projectFs := afero.NewMemMapFs()
	renderConfigs := []standardAPI.RenderConfig{
		{Format: "markdown", Destination: "/docs"},
	}
	action := NewRenderDocStandardAction(mockRepo, projectFs, renderConfigs)

	err := action.Run()

	require.NoError(t, err)

	exists, _ := afero.Exists(projectFs, "/docs/test-standard.md")
	assert.True(t, exists)

	content, _ := afero.ReadFile(projectFs, "/docs/test-standard.md")
	assert.NotEmpty(t, content)
}

func TestRenderDocStandardActionRun_WhenMultipleStandards_ThenWritesAllFiles(t *testing.T) {
	t.Parallel()

	standards := []standardAPI.Standard{
		sampleStandard("First Standard"),
		sampleStandard("Second Standard"),
	}
	mockRepo := standardAPI.NewMockRepository(t)
	mockRepo.EXPECT().GetAll().Return(standards, nil)

	projectFs := afero.NewMemMapFs()
	renderConfigs := []standardAPI.RenderConfig{
		{Format: "markdown", Destination: "/docs"},
	}
	action := NewRenderDocStandardAction(mockRepo, projectFs, renderConfigs)

	err := action.Run()

	require.NoError(t, err)

	exists1, _ := afero.Exists(projectFs, "/docs/first-standard.md")
	exists2, _ := afero.Exists(projectFs, "/docs/second-standard.md")
	assert.True(t, exists1)
	assert.True(t, exists2)
}

func TestRenderDocStandardActionRun_WhenDestinationNotExists_ThenCreatesDirectory(t *testing.T) {
	t.Parallel()

	standard := sampleStandard("Test Standard")
	mockRepo := standardAPI.NewMockRepository(t)
	mockRepo.EXPECT().GetAll().Return([]standardAPI.Standard{standard}, nil)

	projectFs := afero.NewMemMapFs()
	renderConfigs := []standardAPI.RenderConfig{
		{Format: "markdown", Destination: "/new/docs"},
	}
	action := NewRenderDocStandardAction(mockRepo, projectFs, renderConfigs)

	err := action.Run()

	require.NoError(t, err)

	exists, _ := afero.DirExists(projectFs, "/new/docs")
	assert.True(t, exists)

	fileExists, _ := afero.Exists(projectFs, "/new/docs/test-standard.md")
	assert.True(t, fileExists)
}

func TestRenderDocStandardActionRun_WhenCleanDestination_ThenRemovesOnlyMatchingExtensions(t *testing.T) {
	t.Parallel()

	standard := sampleStandard("Test Standard")
	mockRepo := standardAPI.NewMockRepository(t)
	mockRepo.EXPECT().GetAll().Return([]standardAPI.Standard{standard}, nil)

	projectFs := afero.NewMemMapFs()
	require.NoError(t, projectFs.MkdirAll("/docs", 0755))
	require.NoError(t, afero.WriteFile(projectFs, "/docs/old.md", []byte("old content"), 0644))
	require.NoError(t, afero.WriteFile(projectFs, "/docs/keep.txt", []byte("keep this"), 0644))

	renderConfigs := []standardAPI.RenderConfig{
		{Format: "markdown", Destination: "/docs"},
	}
	action := NewRenderDocStandardAction(mockRepo, projectFs, renderConfigs)

	err := action.Run()

	require.NoError(t, err)

	oldExists, _ := afero.Exists(projectFs, "/docs/old.md")
	keepExists, _ := afero.Exists(projectFs, "/docs/keep.txt")
	assert.False(t, oldExists)
	assert.True(t, keepExists)
}

func TestRenderDocStandardActionRun_WhenRepositoryGetAllFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	repoErr := errors.New("repository error")
	mockRepo := standardAPI.NewMockRepository(t)
	mockRepo.EXPECT().GetAll().Return(nil, repoErr)

	projectFs := afero.NewMemMapFs()
	renderConfigs := []standardAPI.RenderConfig{
		{Format: "markdown", Destination: "/docs"},
	}
	action := NewRenderDocStandardAction(mockRepo, projectFs, renderConfigs)

	err := action.Run()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get standards")
	assert.ErrorIs(t, err, repoErr)
}

func TestRenderDocStandardActionRun_WhenUnsupportedFormat_ThenReturnsUnsupportedFormatError(t *testing.T) {
	t.Parallel()

	standard := sampleStandard("Test Standard")
	mockRepo := standardAPI.NewMockRepository(t)
	mockRepo.EXPECT().GetAll().Return([]standardAPI.Standard{standard}, nil)

	projectFs := afero.NewMemMapFs()
	renderConfigs := []standardAPI.RenderConfig{
		{Format: "html", Destination: "/docs"},
	}
	action := NewRenderDocStandardAction(mockRepo, projectFs, renderConfigs)

	err := action.Run()

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnsupportedFormat)
}

func TestRenderDocStandardActionRun_WhenWriteFileFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	standard := sampleStandard("Test Standard")
	mockRepo := standardAPI.NewMockRepository(t)
	mockRepo.EXPECT().GetAll().Return([]standardAPI.Standard{standard}, nil)

	writeErr := errors.New("write file error")
	baseFs := afero.NewMemMapFs()
	projectFs := aferomock.OverrideFs(baseFs, aferomock.FsCallbacks{
		OpenFileFunc: func(name string, flag int, perm os.FileMode) (afero.File, error) {
			return nil, writeErr
		},
	})

	renderConfigs := []standardAPI.RenderConfig{
		{Format: "markdown", Destination: "/docs"},
	}
	action := NewRenderDocStandardAction(mockRepo, projectFs, renderConfigs)

	err := action.Run()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to write file")
}

func TestRenderDocStandardActionRun_WhenCleanDestinationRemoveFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	standard := sampleStandard("Test Standard")
	mockRepo := standardAPI.NewMockRepository(t)
	mockRepo.EXPECT().GetAll().Return([]standardAPI.Standard{standard}, nil)

	baseFs := afero.NewMemMapFs()
	require.NoError(t, baseFs.MkdirAll("/docs", 0755))
	require.NoError(t, afero.WriteFile(baseFs, "/docs/old.md", []byte("old content"), 0644))

	removeErr := errors.New("remove file error")
	projectFs := aferomock.OverrideFs(baseFs, aferomock.FsCallbacks{
		RemoveFunc: func(name string) error {
			return removeErr
		},
	})

	renderConfigs := []standardAPI.RenderConfig{
		{Format: "markdown", Destination: "/docs"},
	}
	action := NewRenderDocStandardAction(mockRepo, projectFs, renderConfigs)

	err := action.Run()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to clean destination")
}
