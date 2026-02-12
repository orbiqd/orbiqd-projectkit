package loader

import (
	"errors"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	skillAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/skill"
	sourceAPI "github.com/orbiqd/orbiqd-projectkit/pkg/source"
)

func validSkillFs(t *testing.T, name, description, instructions string) afero.Fs {
	t.Helper()

	fs := afero.NewMemMapFs()

	require.NoError(t, fs.Mkdir(name, 0755))

	metadata := `name: ` + name + `
description: ` + description + `
`
	require.NoError(t, afero.WriteFile(fs, name+"/metadata.yaml", []byte(metadata), 0644))
	require.NoError(t, afero.WriteFile(fs, name+"/instructions.md", []byte(instructions), 0644))

	return fs
}

func TestLoadAiSkillsFromConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		sources       []skillAPI.SourceConfig
		mockSetup     func(*sourceAPI.MockResolver)
		wantSkillsLen int
	}{
		{
			name:          "WhenNoSources_ThenReturnsNilSkills",
			sources:       []skillAPI.SourceConfig{},
			mockSetup:     func(m *sourceAPI.MockResolver) {},
			wantSkillsLen: 0,
		},
		{
			name: "WhenSingleSourceWithOneSkill_ThenReturnsSkill",
			sources: []skillAPI.SourceConfig{
				{URI: "file://./skills"},
			},
			mockSetup: func(m *sourceAPI.MockResolver) {
				fs := validSkillFs(t, "test-skill", "Test skill description", "Test instructions")
				m.EXPECT().Resolve("file://./skills").Return(fs, nil)
			},
			wantSkillsLen: 1,
		},
		{
			name: "WhenMultipleSources_ThenReturnsCombinedSkills",
			sources: []skillAPI.SourceConfig{
				{URI: "file://./skills1"},
				{URI: "file://./skills2"},
			},
			mockSetup: func(m *sourceAPI.MockResolver) {
				fs1 := validSkillFs(t, "skill-one", "First skill", "First instructions")
				fs2 := validSkillFs(t, "skill-two", "Second skill", "Second instructions")
				m.EXPECT().Resolve("file://./skills1").Return(fs1, nil)
				m.EXPECT().Resolve("file://./skills2").Return(fs2, nil)
			},
			wantSkillsLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockResolver := sourceAPI.NewMockResolver(t)
			tt.mockSetup(mockResolver)

			config := skillAPI.Config{
				Sources: tt.sources,
			}

			skills, err := LoadAiSkillsFromConfig(config, mockResolver)

			require.NoError(t, err)
			assert.Len(t, skills, tt.wantSkillsLen)
		})
	}
}

func TestLoadAiSkillsFromConfig_WhenResolverFails_ThenReturnsResolveError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	resolveErr := errors.New("resolver failed")
	mockResolver.EXPECT().Resolve("file://./skills").Return(nil, resolveErr)

	config := skillAPI.Config{
		Sources: []skillAPI.SourceConfig{
			{URI: "file://./skills"},
		},
	}

	skills, err := LoadAiSkillsFromConfig(config, mockResolver)

	require.Error(t, err)
	assert.Empty(t, skills)
	assert.Contains(t, err.Error(), "resolve: file://./skills")
	assert.ErrorIs(t, err, resolveErr)
}

func TestLoadAiSkillsFromConfig_WhenLoaderFails_ThenReturnsLoadSkillsError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	emptyFs := afero.NewMemMapFs()
	mockResolver.EXPECT().Resolve("file://./skills").Return(emptyFs, nil)

	config := skillAPI.Config{
		Sources: []skillAPI.SourceConfig{
			{URI: "file://./skills"},
		},
	}

	skills, err := LoadAiSkillsFromConfig(config, mockResolver)

	require.Error(t, err)
	assert.Empty(t, skills)
	assert.Contains(t, err.Error(), "load skills:")
	assert.ErrorIs(t, err, skillAPI.ErrNoSkillsFound)
}

func TestLoadAiSkillsFromConfig_WhenSecondSourceResolverFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	fs1 := validSkillFs(t, "skill-one", "First skill", "First instructions")
	resolveErr := errors.New("second resolver failed")

	mockResolver.EXPECT().Resolve("file://./skills1").Return(fs1, nil)
	mockResolver.EXPECT().Resolve("file://./skills2").Return(nil, resolveErr)

	config := skillAPI.Config{
		Sources: []skillAPI.SourceConfig{
			{URI: "file://./skills1"},
			{URI: "file://./skills2"},
		},
	}

	skills, err := LoadAiSkillsFromConfig(config, mockResolver)

	require.Error(t, err)
	assert.Empty(t, skills)
	assert.Contains(t, err.Error(), "resolve: file://./skills2")
	assert.ErrorIs(t, err, resolveErr)
}

func TestLoadAiSkillsFromConfig_WhenSecondSourceLoaderFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	fs1 := validSkillFs(t, "skill-one", "First skill", "First instructions")
	emptyFs := afero.NewMemMapFs()

	mockResolver.EXPECT().Resolve("file://./skills1").Return(fs1, nil)
	mockResolver.EXPECT().Resolve("file://./skills2").Return(emptyFs, nil)

	config := skillAPI.Config{
		Sources: []skillAPI.SourceConfig{
			{URI: "file://./skills1"},
			{URI: "file://./skills2"},
		},
	}

	skills, err := LoadAiSkillsFromConfig(config, mockResolver)

	require.Error(t, err)
	assert.Empty(t, skills)
	assert.Contains(t, err.Error(), "load skills:")
	assert.ErrorIs(t, err, skillAPI.ErrNoSkillsFound)
}
