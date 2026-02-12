package instruction

import (
	"encoding/json"
	"errors"
	"io/fs"
	"path/filepath"
	"testing"

	instructionAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/instruction"
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

func TestFsRepository_GetAll_WhenHasInstructions_ThenReturnsAll(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)
	instructions1 := instructionAPI.Instructions{
		Category: "coding",
		Rules:    []instructionAPI.Rule{"rule1"},
	}
	instructions2 := instructionAPI.Instructions{
		Category: "testing",
		Rules:    []instructionAPI.Rule{"rule2"},
	}

	err := repo.AddInstructions(instructions1)
	require.NoError(t, err)
	err = repo.AddInstructions(instructions2)
	require.NoError(t, err)

	result, err := repo.GetAll()

	require.NoError(t, err)
	require.Len(t, result, 2)

	categoriesFound := make(map[instructionAPI.Category]bool)
	for _, inst := range result {
		categoriesFound[inst.Category] = true
	}

	assert.True(t, categoriesFound["coding"])
	assert.True(t, categoriesFound["testing"])
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

	instructions := instructionAPI.Instructions{
		Category: "coding",
		Rules:    []instructionAPI.Rule{"rule1"},
	}
	err = repo.AddInstructions(instructions)
	require.NoError(t, err)

	result, err := repo.GetAll()

	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, instructionAPI.Category("coding"), result[0].Category)
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

func TestFsRepository_AddInstructions_WhenSingleInstruction_ThenStoresRules(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)
	instructions := instructionAPI.Instructions{
		Category: "coding",
		Rules:    []instructionAPI.Rule{"rule1", "rule2"},
	}

	err := repo.AddInstructions(instructions)
	require.NoError(t, err)

	result, err := repo.GetAll()
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, instructionAPI.Category("coding"), result[0].Category)
	assert.Equal(t, []instructionAPI.Rule{"rule1", "rule2"}, result[0].Rules)
}

func TestFsRepository_AddInstructions_WhenMultipleCategories_ThenStoresEachSeparately(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)
	instructions1 := instructionAPI.Instructions{
		Category: "coding",
		Rules:    []instructionAPI.Rule{"rule1"},
	}
	instructions2 := instructionAPI.Instructions{
		Category: "testing",
		Rules:    []instructionAPI.Rule{"rule2"},
	}

	err := repo.AddInstructions(instructions1)
	require.NoError(t, err)
	err = repo.AddInstructions(instructions2)
	require.NoError(t, err)

	result, err := repo.GetAll()
	require.NoError(t, err)
	require.Len(t, result, 2)

	categoriesFound := make(map[instructionAPI.Category][]instructionAPI.Rule)
	for _, inst := range result {
		categoriesFound[inst.Category] = inst.Rules
	}

	assert.Equal(t, []instructionAPI.Rule{"rule1"}, categoriesFound["coding"])
	assert.Equal(t, []instructionAPI.Rule{"rule2"}, categoriesFound["testing"])
}

func TestFsRepository_AddInstructions_WhenSameCategoryAddedTwice_ThenAppendsRules(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)
	instructions1 := instructionAPI.Instructions{
		Category: "coding",
		Rules:    []instructionAPI.Rule{"rule1"},
	}
	instructions2 := instructionAPI.Instructions{
		Category: "coding",
		Rules:    []instructionAPI.Rule{"rule2", "rule3"},
	}

	err := repo.AddInstructions(instructions1)
	require.NoError(t, err)
	err = repo.AddInstructions(instructions2)
	require.NoError(t, err)

	result, err := repo.GetAll()
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, instructionAPI.Category("coding"), result[0].Category)
	assert.Equal(t, []instructionAPI.Rule{"rule1", "rule2", "rule3"}, result[0].Rules)
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

func TestFsRepository_RemoveAll_WhenHasInstructions_ThenRemovesAll(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)
	instructions1 := instructionAPI.Instructions{
		Category: "coding",
		Rules:    []instructionAPI.Rule{"rule1", "rule2"},
	}
	instructions2 := instructionAPI.Instructions{
		Category: "testing",
		Rules:    []instructionAPI.Rule{"rule3"},
	}

	err := repo.AddInstructions(instructions1)
	require.NoError(t, err)
	err = repo.AddInstructions(instructions2)
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

func TestFsRepository_RemoveAll_WhenCalledMultipleTimes_ThenSucceeds(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)
	instructions := instructionAPI.Instructions{
		Category: "coding",
		Rules:    []instructionAPI.Rule{"rule1"},
	}

	err := repo.AddInstructions(instructions)
	require.NoError(t, err)

	err = repo.RemoveAll()
	require.NoError(t, err)
	err = repo.RemoveAll()
	require.NoError(t, err)

	result, err := repo.GetAll()
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestFsRepository_RemoveAll_WhenFollowedByAdd_ThenAcceptsNewInstructions(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)
	oldInstructions := instructionAPI.Instructions{
		Category: "old-category",
		Rules:    []instructionAPI.Rule{"old-rule"},
	}

	err := repo.AddInstructions(oldInstructions)
	require.NoError(t, err)

	err = repo.RemoveAll()
	require.NoError(t, err)

	newInstructions := instructionAPI.Instructions{
		Category: "new-category",
		Rules:    []instructionAPI.Rule{"new-rule1", "new-rule2"},
	}

	err = repo.AddInstructions(newInstructions)
	require.NoError(t, err)

	result, err := repo.GetAll()
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, instructionAPI.Category("new-category"), result[0].Category)
	assert.Equal(t, []instructionAPI.Rule{"new-rule1", "new-rule2"}, result[0].Rules)
}

func TestFsRepository_AddInstructions_WhenCalled_ThenCreatesFileWithUuidName(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)
	instructions := instructionAPI.Instructions{
		Category: "coding",
		Rules:    []instructionAPI.Rule{"rule1"},
	}

	err := repo.AddInstructions(instructions)
	require.NoError(t, err)

	files, err := afero.ReadDir(fs, ".")
	require.NoError(t, err)
	require.Len(t, files, 1)

	filename := files[0].Name()
	assert.Equal(t, ".json", filepath.Ext(filename))
	assert.Equal(t, 41, len(filename))
}

func TestFsRepository_AddInstructions_WhenCalled_ThenStoresValidJson(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)
	instructions := instructionAPI.Instructions{
		Category: "coding",
		Rules:    []instructionAPI.Rule{"rule1", "rule2"},
	}

	err := repo.AddInstructions(instructions)
	require.NoError(t, err)

	files, err := afero.ReadDir(fs, ".")
	require.NoError(t, err)
	require.Len(t, files, 1)

	content, err := afero.ReadFile(fs, files[0].Name())
	require.NoError(t, err)

	var stored instructionAPI.Instructions
	err = json.Unmarshal(content, &stored)
	require.NoError(t, err)
	assert.Equal(t, instructionAPI.Category("coding"), stored.Category)
	assert.Equal(t, []instructionAPI.Rule{"rule1", "rule2"}, stored.Rules)
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
	err := afero.WriteFile(base, "test.json", []byte(`{"category":"coding","rules":["rule1"]}`), 0644)
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

func TestFsRepository_AddInstructions_WhenReadDirFails_ThenReturnsError(t *testing.T) {
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
	instructions := instructionAPI.Instructions{
		Category: "coding",
		Rules:    []instructionAPI.Rule{"rule1"},
	}

	err := repo.AddInstructions(instructions)

	require.ErrorIs(t, err, readDirErr)
}

func TestFsRepository_AddInstructions_WhenReadFileFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	base := afero.NewMemMapFs()
	err := afero.WriteFile(base, "existing.json", []byte(`{"category":"coding","rules":["rule1"]}`), 0644)
	require.NoError(t, err)

	readFileErr := errors.New("read file failed")
	fs := aferomock.OverrideFs(base, aferomock.FsCallbacks{
		OpenFunc: func(name string) (afero.File, error) {
			if name == "existing.json" {
				return nil, readFileErr
			}
			return base.Open(name)
		},
	})
	repo := NewFsRepository(fs)
	instructions := instructionAPI.Instructions{
		Category: "coding",
		Rules:    []instructionAPI.Rule{"rule2"},
	}

	err = repo.AddInstructions(instructions)

	require.ErrorIs(t, err, readFileErr)
}

func TestFsRepository_AddInstructions_WhenWriteFileFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	writeFileErr := errors.New("write file failed")
	base := afero.NewMemMapFs()
	fsCallbacks := aferomock.OverrideFs(base, aferomock.FsCallbacks{
		OpenFileFunc: func(name string, flag int, perm fs.FileMode) (afero.File, error) {
			return nil, writeFileErr
		},
	})
	repo := NewFsRepository(fsCallbacks)
	instructions := instructionAPI.Instructions{
		Category: "coding",
		Rules:    []instructionAPI.Rule{"rule1"},
	}

	err := repo.AddInstructions(instructions)

	require.ErrorIs(t, err, writeFileErr)
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
	err := afero.WriteFile(base, "test.json", []byte(`{"category":"coding","rules":["rule1"]}`), 0644)
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
