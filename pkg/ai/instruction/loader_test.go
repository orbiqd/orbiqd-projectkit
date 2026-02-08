package instruction

import (
	"errors"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.nhat.io/aferomock"
)

func TestLoader_Load(t *testing.T) {
	tests := []struct {
		name          string
		setupFs       func(fs afero.Fs)
		wantLen       int
		wantErr       error
		checkCategory string
	}{
		{
			name: "single valid yaml file",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "test.yaml", []byte(`category: test-category
rules:
  - Rule one
  - Rule two
`), 0644)
			},
			wantLen:       1,
			wantErr:       nil,
			checkCategory: "test-category",
		},
		{
			name: "multiple valid files yaml and yml",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "first.yaml", []byte(`category: first
rules:
  - Rule A
`), 0644)
				_ = afero.WriteFile(fs, "second.yml", []byte(`category: second
rules:
  - Rule B
`), 0644)
			},
			wantLen: 2,
			wantErr: nil,
		},
		{
			name: "no yaml files",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "test.txt", []byte("not yaml"), 0644)
			},
			wantLen: 0,
			wantErr: ErrNoInstructionsFound,
		},
		{
			name: "invalid yaml syntax",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "invalid.yaml", []byte(`category: test
rules:
  - [invalid yaml structure
`), 0644)
			},
			wantLen: 0,
			wantErr: ErrParseFailed,
		},
		{
			name: "valid yaml but empty category",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "empty.yaml", []byte(`category: ""
rules:
  - Rule one
`), 0644)
			},
			wantLen: 0,
			wantErr: ErrValidationFailed,
		},
		{
			name: "first file error second file valid",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "bad.yaml", []byte(`category: test
rules: []
`), 0644)
				_ = afero.WriteFile(fs, "good.yaml", []byte(`category: good
rules:
  - Rule one
`), 0644)
			},
			wantLen: 0,
			wantErr: ErrValidationFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			tt.setupFs(fs)

			loader := NewLoader(fs)
			got, err := loader.Load()

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
			} else {
				require.NoError(t, err)
				assert.Len(t, got, tt.wantLen)
				if tt.checkCategory != "" && len(got) > 0 {
					assert.Equal(t, Category(tt.checkCategory), got[0].Category)
				}
			}
		})
	}
}

func TestLoader_resolveFiles(t *testing.T) {
	tests := []struct {
		name       string
		setupFs    func(fs afero.Fs)
		customFs   afero.Fs
		wantLen    int
		wantErr    error
		errContain string
	}{
		{
			name: "yaml and yml files",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "test1.yaml", []byte("content"), 0644)
				_ = afero.WriteFile(fs, "test2.yml", []byte("content"), 0644)
			},
			wantLen: 2,
			wantErr: nil,
		},
		{
			name: "mixed extensions",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "test.yaml", []byte("content"), 0644)
				_ = afero.WriteFile(fs, "test.txt", []byte("content"), 0644)
				_ = afero.WriteFile(fs, "test.json", []byte("content"), 0644)
			},
			wantLen: 1,
			wantErr: nil,
		},
		{
			name: "empty directory",
			setupFs: func(fs afero.Fs) {
			},
			wantLen: 0,
			wantErr: ErrNoInstructionsFound,
		},
		{
			name: "only directories",
			setupFs: func(fs afero.Fs) {
				_ = fs.Mkdir("subdir", 0755)
			},
			wantLen: 0,
			wantErr: ErrNoInstructionsFound,
		},
		{
			name: "directory named with yaml extension",
			setupFs: func(fs afero.Fs) {
				_ = fs.Mkdir("subdir.yaml", 0755)
				_ = afero.WriteFile(fs, "valid.yaml", []byte("content"), 0644)
			},
			wantLen: 1,
			wantErr: nil,
		},
		{
			name:       "read directory error",
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

func TestLoader_loadInstructions(t *testing.T) {
	tests := []struct {
		name         string
		setupFs      func(fs afero.Fs)
		filePath     string
		wantCategory Category
		wantRulesLen int
		wantErr      error
	}{
		{
			name: "valid yaml",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "test.yaml", []byte(`category: test-cat
rules:
  - Rule one
  - Rule two
`), 0644)
			},
			filePath:     "test.yaml",
			wantCategory: "test-cat",
			wantRulesLen: 2,
			wantErr:      nil,
		},
		{
			name: "file does not exist",
			setupFs: func(fs afero.Fs) {
			},
			filePath: "missing.yaml",
			wantErr:  ErrReadFailed,
		},
		{
			name: "invalid yaml",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "bad.yaml", []byte(`category: [invalid
`), 0644)
			},
			filePath: "bad.yaml",
			wantErr:  ErrParseFailed,
		},
		{
			name: "empty file",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "empty.yaml", []byte(``), 0644)
			},
			filePath:     "empty.yaml",
			wantCategory: "",
			wantRulesLen: 0,
			wantErr:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			tt.setupFs(fs)

			loader := NewLoader(fs)
			got, err := loader.loadInstructions(tt.filePath)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantCategory, got.Category)
				assert.Len(t, got.Rules, tt.wantRulesLen)
			}
		})
	}
}

func TestLoader_validate(t *testing.T) {
	tests := []struct {
		name         string
		instructions Instructions
		wantErr      error
	}{
		{
			name: "valid instructions",
			instructions: Instructions{
				Category: "test",
				Rules:    []Rule{"Rule 1"},
			},
			wantErr: nil,
		},
		{
			name: "empty category",
			instructions: Instructions{
				Category: "",
				Rules:    []Rule{"Rule 1"},
			},
			wantErr: ErrValidationFailed,
		},
		{
			name: "nil rules",
			instructions: Instructions{
				Category: "test",
				Rules:    nil,
			},
			wantErr: ErrValidationFailed,
		},
		{
			name: "empty rules slice",
			instructions: Instructions{
				Category: "test",
				Rules:    []Rule{},
			},
			wantErr: ErrValidationFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewLoader(afero.NewMemMapFs())
			err := loader.validate(tt.instructions)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
			} else {
				require.NoError(t, err)
			}
		})
	}
}
