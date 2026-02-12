package loader

import (
	"errors"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	standardAPI "github.com/orbiqd/orbiqd-projectkit/pkg/doc/standard"
	sourceAPI "github.com/orbiqd/orbiqd-projectkit/pkg/source"
)

func validStandardFs(t *testing.T, name, version string) afero.Fs {
	t.Helper()

	fs := afero.NewMemMapFs()
	content := `metadata:
  id: test-standard
  name: ` + name + `
  version: ` + version + `
  tags:
    - test-tag
  scope:
    languages:
      - en
  relations:
    standard: []
specification:
  purpose: Test purpose for the standard
  goals:
    - First goal of the standard
requirements:
  rules:
    - level: must
      statement: This is a test requirement
      rationale: This is the rationale for the requirement
examples:
  good:
    - title: Good Example
      language: go
      snippet: |
        package main
        func main() {}
      reason: This is a good example because it follows the standard
`
	require.NoError(t, afero.WriteFile(fs, "standard.yaml", []byte(content), 0644))

	return fs
}

func TestLoadDocStandardsFromConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		sources          []standardAPI.SourceConfig
		mockSetup        func(*sourceAPI.MockResolver)
		wantStandardsLen int
	}{
		{
			name:             "WhenNoSources_ThenReturnsNilStandards",
			sources:          []standardAPI.SourceConfig{},
			mockSetup:        func(m *sourceAPI.MockResolver) {},
			wantStandardsLen: 0,
		},
		{
			name: "WhenSingleSourceWithOneStandard_ThenReturnsStandard",
			sources: []standardAPI.SourceConfig{
				{URI: "file://./standards"},
			},
			mockSetup: func(m *sourceAPI.MockResolver) {
				fs := validStandardFs(t, "Test Standard", "1.0.0")
				m.EXPECT().Resolve("file://./standards").Return(fs, nil)
			},
			wantStandardsLen: 1,
		},
		{
			name: "WhenMultipleSources_ThenReturnsCombinedStandards",
			sources: []standardAPI.SourceConfig{
				{URI: "file://./standards1"},
				{URI: "file://./standards2"},
			},
			mockSetup: func(m *sourceAPI.MockResolver) {
				fs1 := validStandardFs(t, "Standard One", "1.0.0")
				fs2 := validStandardFs(t, "Standard Two", "2.0.0")
				m.EXPECT().Resolve("file://./standards1").Return(fs1, nil)
				m.EXPECT().Resolve("file://./standards2").Return(fs2, nil)
			},
			wantStandardsLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockResolver := sourceAPI.NewMockResolver(t)
			tt.mockSetup(mockResolver)

			config := standardAPI.Config{
				Sources: tt.sources,
			}

			standards, err := LoadDocStandardsFromConfig(config, mockResolver)

			require.NoError(t, err)
			assert.Len(t, standards, tt.wantStandardsLen)
		})
	}
}

func TestLoadDocStandardsFromConfig_WhenResolverFails_ThenReturnsResolveError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	resolveErr := errors.New("resolver failed")
	mockResolver.EXPECT().Resolve("file://./standards").Return(nil, resolveErr)

	config := standardAPI.Config{
		Sources: []standardAPI.SourceConfig{
			{URI: "file://./standards"},
		},
	}

	standards, err := LoadDocStandardsFromConfig(config, mockResolver)

	require.Error(t, err)
	assert.Empty(t, standards)
	assert.Contains(t, err.Error(), "resolve: file://./standards")
	assert.ErrorIs(t, err, resolveErr)
}

func TestLoadDocStandardsFromConfig_WhenLoaderFails_ThenReturnsLoadStandardError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	emptyFs := afero.NewMemMapFs()
	mockResolver.EXPECT().Resolve("file://./standards").Return(emptyFs, nil)

	config := standardAPI.Config{
		Sources: []standardAPI.SourceConfig{
			{URI: "file://./standards"},
		},
	}

	standards, err := LoadDocStandardsFromConfig(config, mockResolver)

	require.Error(t, err)
	assert.Empty(t, standards)
	assert.Contains(t, err.Error(), "load standard:")
	assert.ErrorIs(t, err, standardAPI.ErrNoStandardsFound)
}

func TestLoadDocStandardsFromConfig_WhenSecondSourceResolverFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	fs1 := validStandardFs(t, "Standard One", "1.0.0")
	resolveErr := errors.New("second resolver failed")

	mockResolver.EXPECT().Resolve("file://./standards1").Return(fs1, nil)
	mockResolver.EXPECT().Resolve("file://./standards2").Return(nil, resolveErr)

	config := standardAPI.Config{
		Sources: []standardAPI.SourceConfig{
			{URI: "file://./standards1"},
			{URI: "file://./standards2"},
		},
	}

	standards, err := LoadDocStandardsFromConfig(config, mockResolver)

	require.Error(t, err)
	assert.Empty(t, standards)
	assert.Contains(t, err.Error(), "resolve: file://./standards2")
	assert.ErrorIs(t, err, resolveErr)
}

func TestLoadDocStandardsFromConfig_WhenSecondSourceLoaderFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	fs1 := validStandardFs(t, "Standard One", "1.0.0")
	emptyFs := afero.NewMemMapFs()

	mockResolver.EXPECT().Resolve("file://./standards1").Return(fs1, nil)
	mockResolver.EXPECT().Resolve("file://./standards2").Return(emptyFs, nil)

	config := standardAPI.Config{
		Sources: []standardAPI.SourceConfig{
			{URI: "file://./standards1"},
			{URI: "file://./standards2"},
		},
	}

	standards, err := LoadDocStandardsFromConfig(config, mockResolver)

	require.Error(t, err)
	assert.Empty(t, standards)
	assert.Contains(t, err.Error(), "load standard:")
	assert.ErrorIs(t, err, standardAPI.ErrNoStandardsFound)
}
