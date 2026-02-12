package standard

import (
	"errors"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.nhat.io/aferomock"
)

func TestLoader_resolveFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupFs    func(fs afero.Fs)
		customFs   afero.Fs
		wantLen    int
		wantErr    error
		errContain string
	}{
		{
			name: "single yaml file",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "test.yaml", []byte("content"), 0644)
			},
			wantLen: 1,
			wantErr: nil,
		},
		{
			name: "multiple yaml files",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "first.yaml", []byte("content"), 0644)
				_ = afero.WriteFile(fs, "second.yaml", []byte("content"), 0644)
				_ = afero.WriteFile(fs, "third.yaml", []byte("content"), 0644)
			},
			wantLen: 3,
			wantErr: nil,
		},
		{
			name: "skips directories",
			setupFs: func(fs afero.Fs) {
				_ = fs.Mkdir("subdir", 0755)
				_ = afero.WriteFile(fs, "valid.yaml", []byte("content"), 0644)
			},
			wantLen: 1,
			wantErr: nil,
		},
		{
			name: "skips non-yaml files",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "test.yaml", []byte("content"), 0644)
				_ = afero.WriteFile(fs, "test.json", []byte("content"), 0644)
				_ = afero.WriteFile(fs, "test.txt", []byte("content"), 0644)
				_ = afero.WriteFile(fs, "test.md", []byte("content"), 0644)
			},
			wantLen: 1,
			wantErr: nil,
		},
		{
			name: "empty directory",
			setupFs: func(fs afero.Fs) {
			},
			wantLen: 0,
			wantErr: nil,
		},
		{
			name:     "read directory error",
			customFs: aferomock.OverrideFs(afero.NewMemMapFs(), aferomock.FsCallbacks{
				OpenFunc: func(name string) (afero.File, error) {
					return nil, errors.New("simulated fs error")
				},
			}),
			wantLen:    0,
			errContain: "read directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var fs afero.Fs
			if tt.customFs != nil {
				fs = tt.customFs
			} else {
				fs = afero.NewMemMapFs()
				if tt.setupFs != nil {
					tt.setupFs(fs)
				}
			}

			loader := NewLoader(fs)
			got, err := loader.resolveFiles()

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
			} else if tt.errContain != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContain)
			} else {
				require.NoError(t, err)
				assert.Len(t, got, tt.wantLen)
			}
		})
	}
}

func TestLoader_loadStandard(t *testing.T) {
	t.Parallel()

	validYaml := `metadata:
  name: Test Standard
  version: 1.0.0
  tags:
    - test-tag
  scope:
    languages:
      - en
  relations:
    standard: []
specification:
  purpose: This is a test standard for unit testing purposes
  goals:
    - Verify yaml parsing works correctly
requirements:
  rules:
    - level: must
      statement: The implementation must parse valid yaml correctly
      rationale: Correct parsing is essential for the system to function
examples:
  good:
    - title: Example title
      language: go
      snippet: "fmt.Println(\"hello\")"
      reason: This demonstrates proper usage of the standard
`

	tests := []struct {
		name     string
		setupFs  func(fs afero.Fs)
		filePath string
		wantName string
		wantErr  error
	}{
		{
			name: "valid yaml",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "test.yaml", []byte(validYaml), 0644)
			},
			filePath: "test.yaml",
			wantName: "Test Standard",
			wantErr:  nil,
		},
		{
			name: "invalid yaml syntax",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "bad.yaml", []byte(`metadata: [invalid`), 0644)
			},
			filePath: "bad.yaml",
			wantErr:  ErrParseFailed,
		},
		{
			name: "missing file",
			setupFs: func(fs afero.Fs) {
			},
			filePath: "missing.yaml",
			wantErr:  ErrReadFailed,
		},
		{
			name: "read error",
			setupFs: func(fs afero.Fs) {
			},
			filePath: "error.yaml",
			wantErr:  ErrReadFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fs := afero.NewMemMapFs()
			tt.setupFs(fs)

			loader := NewLoader(fs)
			got, err := loader.loadStandard(tt.filePath)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tt.wantName, got.Metadata.Name)
			}
		})
	}
}

func TestLoader_Load(t *testing.T) {
	t.Parallel()

	validStandardYaml := `metadata:
  name: Integration Test Standard
  version: 1.0.0
  tags:
    - integration-test
  scope:
    languages:
      - en
  relations:
    standard: []
specification:
  purpose: This standard validates the complete loading pipeline
  goals:
    - Demonstrate end-to-end functionality
requirements:
  rules:
    - level: must
      statement: The system must load and validate standards correctly
      rationale: This ensures data integrity throughout the loading process
examples:
  good:
    - title: Valid example
      language: go
      snippet: "package main"
      reason: Demonstrates correct usage
`

	validStandardYaml2 := `metadata:
  name: Second Standard
  version: 2.0.0
  tags:
    - second-test
  scope:
    languages:
      - en
  relations:
    standard: []
specification:
  purpose: This is the second standard for testing multiple files
  goals:
    - Test multiple standard loading
requirements:
  rules:
    - level: should
      statement: The system should handle multiple standards
      rationale: Multiple standards are a common use case
examples:
  good:
    - title: Another example
      language: python
      snippet: "print('hello')"
      reason: Shows proper implementation
`

	invalidStandardYaml := `metadata:
  name: ""
  version: 1.0.0
  tags:
    - test
  scope:
    languages:
      - en
  relations:
    standard: []
specification:
  purpose: This standard has validation errors
  goals:
    - Test validation failure
requirements:
  rules:
    - level: must
      statement: This will fail validation due to empty name
      rationale: Empty names are not allowed
examples:
  good:
    - title: Example
      language: go
      snippet: "code"
      reason: Valid example
`

	tests := []struct {
		name       string
		setupFs    func(fs afero.Fs)
		customFs   afero.Fs
		wantLen    int
		wantErr    error
		wantAnyErr bool
		errContain string
		checkName  string
	}{
		{
			name: "single valid standard yaml",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "standard.yaml", []byte(validStandardYaml), 0644)
			},
			wantLen:   1,
			wantErr:   nil,
			checkName: "Integration Test Standard",
		},
		{
			name: "multiple valid standards",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "first.yaml", []byte(validStandardYaml), 0644)
				_ = afero.WriteFile(fs, "second.yaml", []byte(validStandardYaml2), 0644)
			},
			wantLen: 2,
			wantErr: nil,
		},
		{
			name: "no yaml files",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "readme.txt", []byte("not yaml"), 0644)
			},
			wantLen: 0,
			wantErr: ErrNoStandardsFound,
		},
		{
			name: "invalid yaml syntax",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "broken.yaml", []byte(`metadata: [broken`), 0644)
			},
			wantLen: 0,
			wantErr: ErrParseFailed,
		},
		{
			name: "validation failure",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "invalid.yaml", []byte(invalidStandardYaml), 0644)
			},
			wantLen:    0,
			wantAnyErr: true,
		},
		{
			name:       "resolve files error",
			customFs: aferomock.OverrideFs(afero.NewMemMapFs(), aferomock.FsCallbacks{
				OpenFunc: func(name string) (afero.File, error) {
					return nil, errors.New("simulated fs error")
				},
			}),
			wantLen:    0,
			errContain: "resolve files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var fs afero.Fs
			if tt.customFs != nil {
				fs = tt.customFs
			} else {
				fs = afero.NewMemMapFs()
				if tt.setupFs != nil {
					tt.setupFs(fs)
				}
			}

			loader := NewLoader(fs)
			got, err := loader.Load()

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
			} else if tt.wantAnyErr {
				require.Error(t, err)
			} else if tt.errContain != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContain)
			} else {
				require.NoError(t, err)
				assert.Len(t, got, tt.wantLen)
				if tt.checkName != "" && len(got) > 0 {
					assert.Equal(t, tt.checkName, got[0].Metadata.Name)
				}
			}
		})
	}
}
