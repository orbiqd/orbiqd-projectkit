package skill

import (
	"encoding/json"
	"errors"
	"io/fs"
	"path/filepath"
	"testing"

	skillAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/skill"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.nhat.io/aferomock"
)

func TestFsRepository_AddSkill_WhenNewSkill_ThenStoresSuccessfully(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)
	skill := skillAPI.Skill{
		Metadata: skillAPI.Metadata{
			Name:        "test-skill",
			Description: "Test skill description",
		},
		Instructions: "Test instructions",
	}

	err := repo.AddSkill(skill)
	require.NoError(t, err)

	result, err := repo.GetSkillByName("test-skill")
	require.NoError(t, err)
	assert.Equal(t, skill.Metadata.Name, result.Metadata.Name)
	assert.Equal(t, skill.Metadata.Description, result.Metadata.Description)
	assert.Equal(t, skill.Instructions, result.Instructions)
}

func TestFsRepository_AddSkill_WhenDuplicateName_ThenReturnsAlreadyExistsError(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)
	skill := skillAPI.Skill{
		Metadata: skillAPI.Metadata{
			Name:        "duplicate-skill",
			Description: "First skill",
		},
		Instructions: "First instructions",
	}

	err := repo.AddSkill(skill)
	require.NoError(t, err)

	duplicateSkill := skillAPI.Skill{
		Metadata: skillAPI.Metadata{
			Name:        "duplicate-skill",
			Description: "Second skill",
		},
		Instructions: "Second instructions",
	}

	err = repo.AddSkill(duplicateSkill)
	require.ErrorIs(t, err, skillAPI.ErrSkillAlreadyExists)
}

func TestFsRepository_GetSkillByName_WhenSkillExists_ThenReturnsSkill(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)
	skill := skillAPI.Skill{
		Metadata: skillAPI.Metadata{
			Name:        "existing-skill",
			Description: "Existing skill description",
		},
		Instructions: "Existing instructions",
		Scripts: map[skillAPI.ScriptName]skillAPI.Script{
			"script1": {
				ContentType: "text/plain",
				Content:     []byte("script content"),
			},
		},
	}

	err := repo.AddSkill(skill)
	require.NoError(t, err)

	result, err := repo.GetSkillByName("existing-skill")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, skillAPI.Name("existing-skill"), result.Metadata.Name)
	assert.Equal(t, "Existing skill description", result.Metadata.Description)
	assert.Equal(t, "Existing instructions", result.Instructions)
	require.Len(t, result.Scripts, 1)
	script, exists := result.Scripts["script1"]
	require.True(t, exists)
	assert.Equal(t, "text/plain", script.ContentType)
	assert.Equal(t, []byte("script content"), script.Content)
}

func TestFsRepository_GetSkillByName_WhenSkillNotFound_ThenReturnsError(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)
	skill := skillAPI.Skill{
		Metadata: skillAPI.Metadata{
			Name:        "some-skill",
			Description: "Some skill",
		},
		Instructions: "Some instructions",
	}

	err := repo.AddSkill(skill)
	require.NoError(t, err)

	result, err := repo.GetSkillByName("non-existing-skill")
	require.ErrorIs(t, err, skillAPI.ErrSkillNotFound)
	assert.Nil(t, result)
}

func TestFsRepository_GetSkillByName_WhenEmptyRepository_ThenReturnsError(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)

	result, err := repo.GetSkillByName("any-skill")
	require.ErrorIs(t, err, skillAPI.ErrSkillNotFound)
	assert.Nil(t, result)
}

func TestFsRepository_GetAll_WhenEmpty_ThenReturnsEmptySlice(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)

	skills, err := repo.GetAll()
	require.NoError(t, err)
	require.NotNil(t, skills)
	assert.Empty(t, skills)
}

func TestFsRepository_GetAll_WhenSingleSkill_ThenReturnsSkill(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)
	skill := skillAPI.Skill{
		Metadata: skillAPI.Metadata{
			Name:        "test-skill",
			Description: "Test description",
		},
		Instructions: "Test instructions",
	}

	err := repo.AddSkill(skill)
	require.NoError(t, err)

	skills, err := repo.GetAll()
	require.NoError(t, err)
	require.Len(t, skills, 1)
	assert.Equal(t, skill.Metadata.Name, skills[0].Metadata.Name)
	assert.Equal(t, skill.Metadata.Description, skills[0].Metadata.Description)
}

func TestFsRepository_GetAll_WhenMultipleSkills_ThenReturnsSortedByName(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)

	skill1 := skillAPI.Skill{
		Metadata: skillAPI.Metadata{
			Name:        "zebra-skill",
			Description: "Last alphabetically",
		},
		Instructions: "Instructions 1",
	}

	skill2 := skillAPI.Skill{
		Metadata: skillAPI.Metadata{
			Name:        "alpha-skill",
			Description: "First alphabetically",
		},
		Instructions: "Instructions 2",
	}

	skill3 := skillAPI.Skill{
		Metadata: skillAPI.Metadata{
			Name:        "middle-skill",
			Description: "Middle alphabetically",
		},
		Instructions: "Instructions 3",
	}

	err := repo.AddSkill(skill1)
	require.NoError(t, err)
	err = repo.AddSkill(skill2)
	require.NoError(t, err)
	err = repo.AddSkill(skill3)
	require.NoError(t, err)

	skills, err := repo.GetAll()
	require.NoError(t, err)
	require.Len(t, skills, 3)

	assert.Equal(t, skillAPI.Name("alpha-skill"), skills[0].Metadata.Name)
	assert.Equal(t, skillAPI.Name("middle-skill"), skills[1].Metadata.Name)
	assert.Equal(t, skillAPI.Name("zebra-skill"), skills[2].Metadata.Name)
}

func TestFsRepository_GetAll_WhenSkillsWithScripts_ThenReturnsAllData(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)
	skill := skillAPI.Skill{
		Metadata: skillAPI.Metadata{
			Name:        "skill-with-scripts",
			Description: "Skill description",
		},
		Instructions: "Skill instructions",
		Scripts: map[skillAPI.ScriptName]skillAPI.Script{
			"script1.sh": {
				ContentType: "application/x-sh",
				Content:     []byte("#!/bin/bash\necho test"),
			},
			"script2.py": {
				ContentType: "text/x-python",
				Content:     []byte("print('test')"),
			},
		},
	}

	err := repo.AddSkill(skill)
	require.NoError(t, err)

	skills, err := repo.GetAll()
	require.NoError(t, err)
	require.Len(t, skills, 1)

	result := skills[0]
	assert.Equal(t, skill.Metadata.Name, result.Metadata.Name)
	require.Len(t, result.Scripts, 2)
	assert.Contains(t, result.Scripts, skillAPI.ScriptName("script1.sh"))
	assert.Contains(t, result.Scripts, skillAPI.ScriptName("script2.py"))
}

func TestFsRepository_GetAll_WhenNonJsonFilesAndDirectoriesExist_ThenIgnoresThem(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)

	err := afero.WriteFile(fs, "readme.txt", []byte("some text"), 0644)
	require.NoError(t, err)
	err = afero.WriteFile(fs, "config.yaml", []byte("key: value"), 0644)
	require.NoError(t, err)
	err = fs.Mkdir("subdir", 0755)
	require.NoError(t, err)

	skill := skillAPI.Skill{
		Metadata: skillAPI.Metadata{
			Name:        "test-skill",
			Description: "Test skill",
		},
		Instructions: "Test instructions",
	}
	err = repo.AddSkill(skill)
	require.NoError(t, err)

	result, err := repo.GetAll()

	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, skillAPI.Name("test-skill"), result[0].Metadata.Name)
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

func TestFsRepository_GetSkillByName_WhenInvalidJson_ThenReturnsError(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)

	err := afero.WriteFile(fs, "invalid.json", []byte("not a valid json"), 0644)
	require.NoError(t, err)

	result, err := repo.GetSkillByName("any-skill")

	require.Error(t, err)
	assert.Nil(t, result)
}

func TestFsRepository_RemoveAll_WhenEmpty_ThenSucceeds(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)

	err := repo.RemoveAll()
	require.NoError(t, err)

	skills, err := repo.GetAll()
	require.NoError(t, err)
	assert.Empty(t, skills)
}

func TestFsRepository_RemoveAll_WhenHasSkills_ThenRemovesAll(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)
	skill1 := skillAPI.Skill{
		Metadata: skillAPI.Metadata{
			Name:        "skill1",
			Description: "First skill",
		},
		Instructions: "Instructions 1",
	}
	skill2 := skillAPI.Skill{
		Metadata: skillAPI.Metadata{
			Name:        "skill2",
			Description: "Second skill",
		},
		Instructions: "Instructions 2",
	}

	err := repo.AddSkill(skill1)
	require.NoError(t, err)
	err = repo.AddSkill(skill2)
	require.NoError(t, err)

	skills, err := repo.GetAll()
	require.NoError(t, err)
	require.Len(t, skills, 2)

	err = repo.RemoveAll()
	require.NoError(t, err)

	skills, err = repo.GetAll()
	require.NoError(t, err)
	assert.Empty(t, skills)
}

func TestFsRepository_RemoveAll_WhenCalledMultipleTimes_ThenSucceeds(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)
	skill := skillAPI.Skill{
		Metadata: skillAPI.Metadata{
			Name:        "test-skill",
			Description: "Test skill",
		},
		Instructions: "Test instructions",
	}

	err := repo.AddSkill(skill)
	require.NoError(t, err)

	err = repo.RemoveAll()
	require.NoError(t, err)

	err = repo.RemoveAll()
	require.NoError(t, err)

	skills, err := repo.GetAll()
	require.NoError(t, err)
	assert.Empty(t, skills)
}

func TestFsRepository_RemoveAll_WhenFollowedByAdd_ThenAcceptsNewSkill(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)
	oldSkill := skillAPI.Skill{
		Metadata: skillAPI.Metadata{
			Name:        "old-skill",
			Description: "Old skill",
		},
		Instructions: "Old instructions",
	}

	err := repo.AddSkill(oldSkill)
	require.NoError(t, err)

	err = repo.RemoveAll()
	require.NoError(t, err)

	newSkill := skillAPI.Skill{
		Metadata: skillAPI.Metadata{
			Name:        "new-skill",
			Description: "New skill",
		},
		Instructions: "New instructions",
	}

	err = repo.AddSkill(newSkill)
	require.NoError(t, err)

	skills, err := repo.GetAll()
	require.NoError(t, err)
	require.Len(t, skills, 1)
	assert.Equal(t, skillAPI.Name("new-skill"), skills[0].Metadata.Name)
	assert.Equal(t, "New skill", skills[0].Metadata.Description)
}

func TestFsRepository_RemoveAll_WhenFollowedByGetByName_ThenReturnsNotFound(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)
	skill := skillAPI.Skill{
		Metadata: skillAPI.Metadata{
			Name:        "removed-skill",
			Description: "Will be removed",
		},
		Instructions: "Instructions",
	}

	err := repo.AddSkill(skill)
	require.NoError(t, err)

	result, err := repo.GetSkillByName("removed-skill")
	require.NoError(t, err)
	require.NotNil(t, result)

	err = repo.RemoveAll()
	require.NoError(t, err)

	result, err = repo.GetSkillByName("removed-skill")
	require.ErrorIs(t, err, skillAPI.ErrSkillNotFound)
	assert.Nil(t, result)
}

func TestFsRepository_AddSkill_WhenCalled_ThenCreatesFileWithUuidName(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)
	skill := skillAPI.Skill{
		Metadata: skillAPI.Metadata{
			Name:        "test-skill",
			Description: "Test skill",
		},
		Instructions: "Test instructions",
	}

	err := repo.AddSkill(skill)
	require.NoError(t, err)

	files, err := afero.ReadDir(fs, ".")
	require.NoError(t, err)
	require.Len(t, files, 1)

	filename := files[0].Name()
	assert.Equal(t, ".json", filepath.Ext(filename))
	assert.Equal(t, 41, len(filename))
}

func TestFsRepository_AddSkill_WhenCalled_ThenStoresValidJson(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)
	skill := skillAPI.Skill{
		Metadata: skillAPI.Metadata{
			Name:        "test-skill",
			Description: "Test skill",
		},
		Instructions: "Test instructions",
	}

	err := repo.AddSkill(skill)
	require.NoError(t, err)

	files, err := afero.ReadDir(fs, ".")
	require.NoError(t, err)
	require.Len(t, files, 1)

	content, err := afero.ReadFile(fs, files[0].Name())
	require.NoError(t, err)

	var stored skillAPI.Skill
	err = json.Unmarshal(content, &stored)
	require.NoError(t, err)
	assert.Equal(t, skillAPI.Name("test-skill"), stored.Metadata.Name)
	assert.Equal(t, "Test skill", stored.Metadata.Description)
	assert.Equal(t, "Test instructions", stored.Instructions)
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
	skill := skillAPI.Skill{
		Metadata: skillAPI.Metadata{
			Name:        "test-skill",
			Description: "Test skill",
		},
		Instructions: "Test instructions",
	}
	data, err := json.Marshal(skill)
	require.NoError(t, err)
	err = afero.WriteFile(base, "test.json", data, 0644)
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

func TestFsRepository_GetSkillByName_WhenReadDirFails_ThenReturnsError(t *testing.T) {
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

	result, err := repo.GetSkillByName("any-skill")

	require.ErrorIs(t, err, readDirErr)
	assert.Nil(t, result)
}

func TestFsRepository_GetSkillByName_WhenReadFileFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	base := afero.NewMemMapFs()
	skill := skillAPI.Skill{
		Metadata: skillAPI.Metadata{
			Name:        "test-skill",
			Description: "Test skill",
		},
		Instructions: "Test instructions",
	}
	data, err := json.Marshal(skill)
	require.NoError(t, err)
	err = afero.WriteFile(base, "test.json", data, 0644)
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

	result, err := repo.GetSkillByName("test-skill")

	require.ErrorIs(t, err, readFileErr)
	assert.Nil(t, result)
}

func TestFsRepository_AddSkill_WhenReadDirFails_ThenReturnsError(t *testing.T) {
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
	skill := skillAPI.Skill{
		Metadata: skillAPI.Metadata{
			Name:        "test-skill",
			Description: "Test skill",
		},
		Instructions: "Test instructions",
	}

	err := repo.AddSkill(skill)

	require.ErrorIs(t, err, readDirErr)
}

func TestFsRepository_AddSkill_WhenReadFileFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	base := afero.NewMemMapFs()
	skill := skillAPI.Skill{
		Metadata: skillAPI.Metadata{
			Name:        "existing-skill",
			Description: "Existing skill",
		},
		Instructions: "Existing instructions",
	}
	data, err := json.Marshal(skill)
	require.NoError(t, err)
	err = afero.WriteFile(base, "existing.json", data, 0644)
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
	newSkill := skillAPI.Skill{
		Metadata: skillAPI.Metadata{
			Name:        "new-skill",
			Description: "New skill",
		},
		Instructions: "New instructions",
	}

	err = repo.AddSkill(newSkill)

	require.ErrorIs(t, err, readFileErr)
}

func TestFsRepository_AddSkill_WhenWriteFileFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	writeFileErr := errors.New("write file failed")
	base := afero.NewMemMapFs()
	fs := aferomock.OverrideFs(base, aferomock.FsCallbacks{
		OpenFileFunc: func(name string, flag int, perm fs.FileMode) (afero.File, error) {
			return nil, writeFileErr
		},
	})
	repo := NewFsRepository(fs)
	skill := skillAPI.Skill{
		Metadata: skillAPI.Metadata{
			Name:        "test-skill",
			Description: "Test skill",
		},
		Instructions: "Test instructions",
	}

	err := repo.AddSkill(skill)

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
	skill := skillAPI.Skill{
		Metadata: skillAPI.Metadata{
			Name:        "test-skill",
			Description: "Test skill",
		},
		Instructions: "Test instructions",
	}
	data, err := json.Marshal(skill)
	require.NoError(t, err)
	err = afero.WriteFile(base, "test.json", data, 0644)
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
