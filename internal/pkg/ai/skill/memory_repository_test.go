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
