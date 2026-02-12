package loader

import (
	"errors"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	rulebookAPI "github.com/orbiqd/orbiqd-projectkit/pkg/rulebook"
	sourceAPI "github.com/orbiqd/orbiqd-projectkit/pkg/source"
)

func validRulebookFs(t *testing.T, category string, rules []string, skillName, skillDescription, skillInstructions string) afero.Fs {
	t.Helper()

	fs := afero.NewMemMapFs()

	rulebookContent := `ai:
  instruction:
    sources:
      - uri: rulebook://ai/instructions
  skill:
    sources:
      - uri: rulebook://ai/skills
doc:
  standard:
    sources:
      - uri: rulebook://docs/standards
`
	require.NoError(t, afero.WriteFile(fs, "rulebook.yaml", []byte(rulebookContent), 0644))

	require.NoError(t, fs.MkdirAll("/ai/instructions", 0755))
	instructionContent := "category: " + category + "\nrules:\n"
	for _, rule := range rules {
		instructionContent += "  - " + rule + "\n"
	}
	require.NoError(t, afero.WriteFile(fs, "/ai/instructions/"+category+".yaml", []byte(instructionContent), 0644))

	require.NoError(t, fs.MkdirAll("/ai/skills/"+skillName, 0755))
	skillMetadata := `name: ` + skillName + `
description: ` + skillDescription + `
`
	require.NoError(t, afero.WriteFile(fs, "/ai/skills/"+skillName+"/metadata.yaml", []byte(skillMetadata), 0644))
	require.NoError(t, afero.WriteFile(fs, "/ai/skills/"+skillName+"/instructions.md", []byte(skillInstructions), 0644))

	require.NoError(t, fs.MkdirAll("/docs/standards", 0755))
	standardContent := `metadata:
  id: test-standard
  name: test-standard
  version: 0.1.0
  tags:
    - test
  scope:
    languages:
      - go
  relations:
    standard: []
specification:
  purpose: A test documentation standard for testing purposes only
  goals:
    - follow the test standard
requirements:
  rules:
    - level: must
      statement: Follow the test standard rules
      rationale: Testing requires following standards
examples:
  good:
    - title: Test example
      language: go
      snippet: log.Info("test")
      reason: Uses test standard
`
	require.NoError(t, afero.WriteFile(fs, "/docs/standards/test-standard.yaml", []byte(standardContent), 0644))

	return fs
}

func TestLoadRulebooksFromConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		sources          []rulebookAPI.SourceConfig
		mockSetup        func(*sourceAPI.MockResolver)
		wantRulebooksLen int
	}{
		{
			name:             "WhenNoSources_ThenReturnsNilRulebooks",
			sources:          []rulebookAPI.SourceConfig{},
			mockSetup:        func(m *sourceAPI.MockResolver) {},
			wantRulebooksLen: 0,
		},
		{
			name: "WhenSingleSource_ThenReturnsRulebook",
			sources: []rulebookAPI.SourceConfig{
				{URI: "file://./rulebooks"},
			},
			mockSetup: func(m *sourceAPI.MockResolver) {
				fs := validRulebookFs(t, "test-category", []string{"Rule one"}, "test-skill", "Test skill", "Test instructions")
				m.EXPECT().Resolve("file://./rulebooks").Return(fs, nil)
			},
			wantRulebooksLen: 1,
		},
		{
			name: "WhenMultipleSources_ThenReturnsCombinedRulebooks",
			sources: []rulebookAPI.SourceConfig{
				{URI: "file://./rulebooks1"},
				{URI: "file://./rulebooks2"},
			},
			mockSetup: func(m *sourceAPI.MockResolver) {
				fs1 := validRulebookFs(t, "category-one", []string{"First rule"}, "skill-one", "First skill", "First instructions")
				fs2 := validRulebookFs(t, "category-two", []string{"Second rule"}, "skill-two", "Second skill", "Second instructions")
				m.EXPECT().Resolve("file://./rulebooks1").Return(fs1, nil)
				m.EXPECT().Resolve("file://./rulebooks2").Return(fs2, nil)
			},
			wantRulebooksLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockResolver := sourceAPI.NewMockResolver(t)
			tt.mockSetup(mockResolver)

			config := rulebookAPI.Config{
				Sources: tt.sources,
			}

			rulebooks, err := LoadRulebooksFromConfig(config, mockResolver)

			require.NoError(t, err)
			assert.Len(t, rulebooks, tt.wantRulebooksLen)

			for _, rb := range rulebooks {
				if len(rb.Doc.Standards) > 0 {
					assert.Equal(t, "test-standard", rb.Doc.Standards[0].Metadata.Name)
				}
			}
		})
	}
}

func TestLoadRulebooksFromConfig_WhenResolverFails_ThenReturnsResolveError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	resolveErr := errors.New("resolver failed")
	mockResolver.EXPECT().Resolve("file://./rulebooks").Return(nil, resolveErr)

	config := rulebookAPI.Config{
		Sources: []rulebookAPI.SourceConfig{
			{URI: "file://./rulebooks"},
		},
	}

	rulebooks, err := LoadRulebooksFromConfig(config, mockResolver)

	require.Error(t, err)
	assert.Empty(t, rulebooks)
	assert.Contains(t, err.Error(), "resolve rulebook uri")
	assert.ErrorIs(t, err, resolveErr)
}

func TestLoadRulebooksFromConfig_WhenLoaderFails_ThenReturnsLoadRulebookError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	emptyFs := afero.NewMemMapFs()
	mockResolver.EXPECT().Resolve("file://./rulebooks").Return(emptyFs, nil)

	config := rulebookAPI.Config{
		Sources: []rulebookAPI.SourceConfig{
			{URI: "file://./rulebooks"},
		},
	}

	rulebooks, err := LoadRulebooksFromConfig(config, mockResolver)

	require.Error(t, err)
	assert.Empty(t, rulebooks)
	assert.Contains(t, err.Error(), "load rulebook")
	assert.ErrorIs(t, err, rulebookAPI.ErrMissingMetadataFile)
}

func TestLoadRulebooksFromConfig_WhenSecondSourceResolverFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	fs1 := validRulebookFs(t, "category-one", []string{"First rule"}, "skill-one", "First skill", "First instructions")
	resolveErr := errors.New("second resolver failed")

	mockResolver.EXPECT().Resolve("file://./rulebooks1").Return(fs1, nil)
	mockResolver.EXPECT().Resolve("file://./rulebooks2").Return(nil, resolveErr)

	config := rulebookAPI.Config{
		Sources: []rulebookAPI.SourceConfig{
			{URI: "file://./rulebooks1"},
			{URI: "file://./rulebooks2"},
		},
	}

	rulebooks, err := LoadRulebooksFromConfig(config, mockResolver)

	require.Error(t, err)
	assert.Empty(t, rulebooks)
	assert.Contains(t, err.Error(), "resolve rulebook uri")
	assert.ErrorIs(t, err, resolveErr)
}

func TestLoadRulebooksFromConfig_WhenSecondSourceLoaderFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	fs1 := validRulebookFs(t, "category-one", []string{"First rule"}, "skill-one", "First skill", "First instructions")
	emptyFs := afero.NewMemMapFs()

	mockResolver.EXPECT().Resolve("file://./rulebooks1").Return(fs1, nil)
	mockResolver.EXPECT().Resolve("file://./rulebooks2").Return(emptyFs, nil)

	config := rulebookAPI.Config{
		Sources: []rulebookAPI.SourceConfig{
			{URI: "file://./rulebooks1"},
			{URI: "file://./rulebooks2"},
		},
	}

	rulebooks, err := LoadRulebooksFromConfig(config, mockResolver)

	require.Error(t, err)
	assert.Empty(t, rulebooks)
	assert.Contains(t, err.Error(), "load rulebook")
	assert.ErrorIs(t, err, rulebookAPI.ErrMissingMetadataFile)
}
