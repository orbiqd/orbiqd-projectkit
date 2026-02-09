package skill

import (
	"testing"

	skillAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/skill"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryRepository_AddSkill_WhenNewSkill_ThenStoresSuccessfully(t *testing.T) {
	t.Parallel()

	repo := NewMemoryRepository()
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

func TestMemoryRepository_AddSkill_WhenDuplicateName_ThenReturnsAlreadyExistsError(t *testing.T) {
	t.Parallel()

	repo := NewMemoryRepository()
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

func TestMemoryRepository_GetSkillByName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupRepo   func(repo *MemoryRepository)
		skillName   skillAPI.Name
		wantErr     error
		wantSkill   *skillAPI.Skill
		checkResult func(t *testing.T, result *skillAPI.Skill)
	}{
		{
			name: "skill exists",
			setupRepo: func(repo *MemoryRepository) {
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
				_ = repo.AddSkill(skill)
			},
			skillName: "existing-skill",
			wantErr:   nil,
			checkResult: func(t *testing.T, result *skillAPI.Skill) {
				t.Helper()
				require.NotNil(t, result)
				assert.Equal(t, skillAPI.Name("existing-skill"), result.Metadata.Name)
				assert.Equal(t, "Existing skill description", result.Metadata.Description)
				assert.Equal(t, "Existing instructions", result.Instructions)
				require.Len(t, result.Scripts, 1)
				script, exists := result.Scripts["script1"]
				require.True(t, exists)
				assert.Equal(t, "text/plain", script.ContentType)
				assert.Equal(t, []byte("script content"), script.Content)
			},
		},
		{
			name: "skill not found",
			setupRepo: func(repo *MemoryRepository) {
				skill := skillAPI.Skill{
					Metadata: skillAPI.Metadata{
						Name:        "some-skill",
						Description: "Some skill",
					},
					Instructions: "Some instructions",
				}
				_ = repo.AddSkill(skill)
			},
			skillName: "non-existing-skill",
			wantErr:   skillAPI.ErrSkillNotFound,
		},
		{
			name:      "empty repository",
			setupRepo: func(repo *MemoryRepository) {},
			skillName: "any-skill",
			wantErr:   skillAPI.ErrSkillNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := NewMemoryRepository()
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}

			result, err := repo.GetSkillByName(tt.skillName)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				if tt.checkResult != nil {
					tt.checkResult(t, result)
				}
			}
		})
	}
}

func TestMemoryRepository_GetAll_WhenEmpty_ThenReturnsEmptySlice(t *testing.T) {
	t.Parallel()

	repo := NewMemoryRepository()

	skills, err := repo.GetAll()
	require.NoError(t, err)
	require.NotNil(t, skills)
	assert.Empty(t, skills)
}

func TestMemoryRepository_GetAll_WhenSingleSkill_ThenReturnsSkill(t *testing.T) {
	t.Parallel()

	repo := NewMemoryRepository()
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

func TestMemoryRepository_GetAll_WhenMultipleSkills_ThenReturnsSortedByName(t *testing.T) {
	t.Parallel()

	repo := NewMemoryRepository()

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

func TestMemoryRepository_GetAll_WhenSkillsWithScripts_ThenReturnsAllData(t *testing.T) {
	t.Parallel()

	repo := NewMemoryRepository()
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
