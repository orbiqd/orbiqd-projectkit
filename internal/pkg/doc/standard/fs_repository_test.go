package standard

import (
	"encoding/json"
	"errors"
	"io/fs"
	"path/filepath"
	"testing"

	standardAPI "github.com/orbiqd/orbiqd-projectkit/pkg/doc/standard"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.nhat.io/aferomock"
)

func TestFsRepository_GetAll_WhenEmpty_ThenReturnsEmptySlice(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)

	result, err := repo.GetAll()

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestFsRepository_GetAll_WhenSingleStandard_ThenReturnsIt(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)
	standard := standardAPI.Standard{
		Metadata: standardAPI.Metadata{
			Name:    "Test Standard",
			Version: "1.0.0",
			Tags:    []string{"test-tag"},
			Scope: standardAPI.ScopeMetadata{
				Languages: []string{"en"},
			},
			Related: standardAPI.RelationMetadata{},
		},
		Specification: standardAPI.Specification{
			Purpose: "Test purpose for standard",
			Goals:   []string{"Test goal for standard"},
		},
		Requirements: standardAPI.Requirements{
			Rules: []standardAPI.RequirementRule{
				{
					Level:     "must",
					Statement: "Test statement for requirement",
					Rationale: "Test rationale for requirement",
				},
			},
		},
		Examples: standardAPI.Examples{
			Good: []standardAPI.Example{
				{
					Title:    "Test Example",
					Language: "go",
					Snippet:  "code snippet",
					Reason:   "Test reason for example",
				},
			},
		},
	}

	err := repo.AddStandard(standard)
	require.NoError(t, err)

	result, err := repo.GetAll()

	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "Test Standard", result[0].Metadata.Name)
}

func TestFsRepository_GetAll_WhenMultipleStandards_ThenReturnsSortedByName(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)
	standard1 := standardAPI.Standard{
		Metadata: standardAPI.Metadata{
			Name:    "Zebra Standard",
			Version: "1.0.0",
			Tags:    []string{"test-tag"},
			Scope: standardAPI.ScopeMetadata{
				Languages: []string{"en"},
			},
			Related: standardAPI.RelationMetadata{},
		},
		Specification: standardAPI.Specification{
			Purpose: "Test purpose for standard",
			Goals:   []string{"Test goal for standard"},
		},
		Requirements: standardAPI.Requirements{
			Rules: []standardAPI.RequirementRule{
				{
					Level:     "must",
					Statement: "Test statement for requirement",
					Rationale: "Test rationale for requirement",
				},
			},
		},
		Examples: standardAPI.Examples{
			Good: []standardAPI.Example{
				{
					Title:    "Test Example",
					Language: "go",
					Snippet:  "code snippet",
					Reason:   "Test reason for example",
				},
			},
		},
	}
	standard2 := standardAPI.Standard{
		Metadata: standardAPI.Metadata{
			Name:    "Alpha Standard",
			Version: "1.0.0",
			Tags:    []string{"test-tag"},
			Scope: standardAPI.ScopeMetadata{
				Languages: []string{"en"},
			},
			Related: standardAPI.RelationMetadata{},
		},
		Specification: standardAPI.Specification{
			Purpose: "Test purpose for standard",
			Goals:   []string{"Test goal for standard"},
		},
		Requirements: standardAPI.Requirements{
			Rules: []standardAPI.RequirementRule{
				{
					Level:     "must",
					Statement: "Test statement for requirement",
					Rationale: "Test rationale for requirement",
				},
			},
		},
		Examples: standardAPI.Examples{
			Good: []standardAPI.Example{
				{
					Title:    "Test Example",
					Language: "go",
					Snippet:  "code snippet",
					Reason:   "Test reason for example",
				},
			},
		},
	}
	standard3 := standardAPI.Standard{
		Metadata: standardAPI.Metadata{
			Name:    "Middle Standard",
			Version: "1.0.0",
			Tags:    []string{"test-tag"},
			Scope: standardAPI.ScopeMetadata{
				Languages: []string{"en"},
			},
			Related: standardAPI.RelationMetadata{},
		},
		Specification: standardAPI.Specification{
			Purpose: "Test purpose for standard",
			Goals:   []string{"Test goal for standard"},
		},
		Requirements: standardAPI.Requirements{
			Rules: []standardAPI.RequirementRule{
				{
					Level:     "must",
					Statement: "Test statement for requirement",
					Rationale: "Test rationale for requirement",
				},
			},
		},
		Examples: standardAPI.Examples{
			Good: []standardAPI.Example{
				{
					Title:    "Test Example",
					Language: "go",
					Snippet:  "code snippet",
					Reason:   "Test reason for example",
				},
			},
		},
	}

	err := repo.AddStandard(standard1)
	require.NoError(t, err)
	err = repo.AddStandard(standard2)
	require.NoError(t, err)
	err = repo.AddStandard(standard3)
	require.NoError(t, err)

	result, err := repo.GetAll()

	require.NoError(t, err)
	require.Len(t, result, 3)
	assert.Equal(t, "Alpha Standard", result[0].Metadata.Name)
	assert.Equal(t, "Middle Standard", result[1].Metadata.Name)
	assert.Equal(t, "Zebra Standard", result[2].Metadata.Name)
}

func TestFsRepository_GetAll_WhenNonJsonFilesExist_ThenIgnoresThem(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)

	err := afero.WriteFile(fs, "readme.txt", []byte("some text"), 0644)
	require.NoError(t, err)
	err = afero.WriteFile(fs, "config.yaml", []byte("key: value"), 0644)
	require.NoError(t, err)
	err = fs.Mkdir("subdir", 0755)
	require.NoError(t, err)

	standard := standardAPI.Standard{
		Metadata: standardAPI.Metadata{
			Name:    "Test Standard",
			Version: "1.0.0",
			Tags:    []string{"test-tag"},
			Scope: standardAPI.ScopeMetadata{
				Languages: []string{"en"},
			},
			Related: standardAPI.RelationMetadata{},
		},
		Specification: standardAPI.Specification{
			Purpose: "Test purpose for standard",
			Goals:   []string{"Test goal for standard"},
		},
		Requirements: standardAPI.Requirements{
			Rules: []standardAPI.RequirementRule{
				{
					Level:     "must",
					Statement: "Test statement for requirement",
					Rationale: "Test rationale for requirement",
				},
			},
		},
		Examples: standardAPI.Examples{
			Good: []standardAPI.Example{
				{
					Title:    "Test Example",
					Language: "go",
					Snippet:  "code snippet",
					Reason:   "Test reason for example",
				},
			},
		},
	}
	err = repo.AddStandard(standard)
	require.NoError(t, err)

	result, err := repo.GetAll()

	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "Test Standard", result[0].Metadata.Name)
}

func TestFsRepository_GetAll_WhenInvalidJson_ThenReturnsError(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)

	err := afero.WriteFile(fs, "invalid.json", []byte("not a valid json"), 0644)
	require.NoError(t, err)

	result, err := repo.GetAll()

	require.Error(t, err)
	assert.Nil(t, result)
}

func TestFsRepository_GetAll_WhenReadDirFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	readDirErr := errors.New("read dir failed")
	base := afero.NewMemMapFs()
	fs := aferomock.OverrideFs(base, aferomock.FsCallbacks{
		OpenFunc: func(name string) (afero.File, error) {
			if name == "." {
				return nil, readDirErr
			}
			return base.Open(name)
		},
	})
	repo := NewFsRepository(fs)

	result, err := repo.GetAll()

	require.ErrorIs(t, err, readDirErr)
	assert.Nil(t, result)
}

func TestFsRepository_GetAll_WhenReadFileFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	base := afero.NewMemMapFs()
	err := afero.WriteFile(base, "test.json", []byte(`{"metadata":{"name":"Test"}}`), 0644)
	require.NoError(t, err)

	readFileErr := errors.New("read file failed")
	fs := aferomock.OverrideFs(base, aferomock.FsCallbacks{
		OpenFunc: func(name string) (afero.File, error) {
			if name == "test.json" {
				return nil, readFileErr
			}
			return base.Open(name)
		},
	})
	repo := NewFsRepository(fs)

	result, err := repo.GetAll()

	require.ErrorIs(t, err, readFileErr)
	assert.Nil(t, result)
}

func TestFsRepository_AddStandard_WhenValid_ThenStoresAndCanBeRetrieved(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)
	standard := standardAPI.Standard{
		Metadata: standardAPI.Metadata{
			Name:    "Test Standard",
			Version: "1.0.0",
			Tags:    []string{"test-tag"},
			Scope: standardAPI.ScopeMetadata{
				Languages: []string{"en"},
			},
			Related: standardAPI.RelationMetadata{},
		},
		Specification: standardAPI.Specification{
			Purpose: "Test purpose for standard",
			Goals:   []string{"Test goal for standard"},
		},
		Requirements: standardAPI.Requirements{
			Rules: []standardAPI.RequirementRule{
				{
					Level:     "must",
					Statement: "Test statement for requirement",
					Rationale: "Test rationale for requirement",
				},
			},
		},
		Examples: standardAPI.Examples{
			Good: []standardAPI.Example{
				{
					Title:    "Test Example",
					Language: "go",
					Snippet:  "code snippet",
					Reason:   "Test reason for example",
				},
			},
		},
	}

	err := repo.AddStandard(standard)
	require.NoError(t, err)

	result, err := repo.GetAll()
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "Test Standard", result[0].Metadata.Name)
	assert.Equal(t, "1.0.0", result[0].Metadata.Version)
}

func TestFsRepository_AddStandard_WhenCalled_ThenCreatesJsonFile(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)
	standard := standardAPI.Standard{
		Metadata: standardAPI.Metadata{
			Name:    "Test Standard",
			Version: "1.0.0",
			Tags:    []string{"test-tag"},
			Scope: standardAPI.ScopeMetadata{
				Languages: []string{"en"},
			},
			Related: standardAPI.RelationMetadata{},
		},
		Specification: standardAPI.Specification{
			Purpose: "Test purpose for standard",
			Goals:   []string{"Test goal for standard"},
		},
		Requirements: standardAPI.Requirements{
			Rules: []standardAPI.RequirementRule{
				{
					Level:     "must",
					Statement: "Test statement for requirement",
					Rationale: "Test rationale for requirement",
				},
			},
		},
		Examples: standardAPI.Examples{
			Good: []standardAPI.Example{
				{
					Title:    "Test Example",
					Language: "go",
					Snippet:  "code snippet",
					Reason:   "Test reason for example",
				},
			},
		},
	}

	err := repo.AddStandard(standard)
	require.NoError(t, err)

	files, err := afero.ReadDir(fs, ".")
	require.NoError(t, err)
	require.Len(t, files, 1)

	filename := files[0].Name()
	assert.Equal(t, ".json", filepath.Ext(filename))
	assert.Equal(t, 41, len(filename))

	content, err := afero.ReadFile(fs, filename)
	require.NoError(t, err)

	var stored standardAPI.Standard
	err = json.Unmarshal(content, &stored)
	require.NoError(t, err)
	assert.Equal(t, "Test Standard", stored.Metadata.Name)
	assert.Equal(t, "1.0.0", stored.Metadata.Version)
}

func TestFsRepository_AddStandard_WhenWriteFileFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	writeFileErr := errors.New("write file failed")
	base := afero.NewMemMapFs()
	fs := aferomock.OverrideFs(base, aferomock.FsCallbacks{
		OpenFileFunc: func(name string, flag int, perm fs.FileMode) (afero.File, error) {
			return nil, writeFileErr
		},
	})
	repo := NewFsRepository(fs)
	standard := standardAPI.Standard{
		Metadata: standardAPI.Metadata{
			Name:    "Test Standard",
			Version: "1.0.0",
			Tags:    []string{"test-tag"},
			Scope: standardAPI.ScopeMetadata{
				Languages: []string{"en"},
			},
			Related: standardAPI.RelationMetadata{},
		},
		Specification: standardAPI.Specification{
			Purpose: "Test purpose for standard",
			Goals:   []string{"Test goal for standard"},
		},
		Requirements: standardAPI.Requirements{
			Rules: []standardAPI.RequirementRule{
				{
					Level:     "must",
					Statement: "Test statement for requirement",
					Rationale: "Test rationale for requirement",
				},
			},
		},
		Examples: standardAPI.Examples{
			Good: []standardAPI.Example{
				{
					Title:    "Test Example",
					Language: "go",
					Snippet:  "code snippet",
					Reason:   "Test reason for example",
				},
			},
		},
	}

	err := repo.AddStandard(standard)

	require.ErrorIs(t, err, writeFileErr)
}

func TestFsRepository_RemoveAll_WhenEmpty_ThenSucceeds(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)

	err := repo.RemoveAll()
	require.NoError(t, err)

	result, err := repo.GetAll()
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestFsRepository_RemoveAll_WhenHasStandards_ThenRemovesAll(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)
	standard1 := standardAPI.Standard{
		Metadata: standardAPI.Metadata{
			Name:    "Test Standard 1",
			Version: "1.0.0",
			Tags:    []string{"test-tag"},
			Scope: standardAPI.ScopeMetadata{
				Languages: []string{"en"},
			},
			Related: standardAPI.RelationMetadata{},
		},
		Specification: standardAPI.Specification{
			Purpose: "Test purpose for standard",
			Goals:   []string{"Test goal for standard"},
		},
		Requirements: standardAPI.Requirements{
			Rules: []standardAPI.RequirementRule{
				{
					Level:     "must",
					Statement: "Test statement for requirement",
					Rationale: "Test rationale for requirement",
				},
			},
		},
		Examples: standardAPI.Examples{
			Good: []standardAPI.Example{
				{
					Title:    "Test Example",
					Language: "go",
					Snippet:  "code snippet",
					Reason:   "Test reason for example",
				},
			},
		},
	}
	standard2 := standardAPI.Standard{
		Metadata: standardAPI.Metadata{
			Name:    "Test Standard 2",
			Version: "1.0.0",
			Tags:    []string{"test-tag"},
			Scope: standardAPI.ScopeMetadata{
				Languages: []string{"en"},
			},
			Related: standardAPI.RelationMetadata{},
		},
		Specification: standardAPI.Specification{
			Purpose: "Test purpose for standard",
			Goals:   []string{"Test goal for standard"},
		},
		Requirements: standardAPI.Requirements{
			Rules: []standardAPI.RequirementRule{
				{
					Level:     "must",
					Statement: "Test statement for requirement",
					Rationale: "Test rationale for requirement",
				},
			},
		},
		Examples: standardAPI.Examples{
			Good: []standardAPI.Example{
				{
					Title:    "Test Example",
					Language: "go",
					Snippet:  "code snippet",
					Reason:   "Test reason for example",
				},
			},
		},
	}

	err := repo.AddStandard(standard1)
	require.NoError(t, err)
	err = repo.AddStandard(standard2)
	require.NoError(t, err)

	result, err := repo.GetAll()
	require.NoError(t, err)
	require.Len(t, result, 2)

	err = repo.RemoveAll()
	require.NoError(t, err)

	result, err = repo.GetAll()
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestFsRepository_RemoveAll_WhenReadDirFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	readDirErr := errors.New("read dir failed")
	base := afero.NewMemMapFs()
	fs := aferomock.OverrideFs(base, aferomock.FsCallbacks{
		OpenFunc: func(name string) (afero.File, error) {
			if name == "." {
				return nil, readDirErr
			}
			return base.Open(name)
		},
	})
	repo := NewFsRepository(fs)

	err := repo.RemoveAll()

	require.ErrorIs(t, err, readDirErr)
}

func TestFsRepository_RemoveAll_WhenRemoveFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	base := afero.NewMemMapFs()
	err := afero.WriteFile(base, "test.json", []byte(`{"metadata":{"name":"Test"}}`), 0644)
	require.NoError(t, err)

	removeErr := errors.New("remove failed")
	fs := aferomock.OverrideFs(base, aferomock.FsCallbacks{
		RemoveFunc: func(name string) error {
			return removeErr
		},
	})
	repo := NewFsRepository(fs)

	err = repo.RemoveAll()

	require.ErrorIs(t, err, removeErr)
}
