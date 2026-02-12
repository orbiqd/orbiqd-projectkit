package mcp

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
		name      string
		setupFs   func(fs afero.Fs)
		wantLen   int
		wantErr   error
		checkName string
	}{
		{
			name: "single valid yaml file",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "test.yaml", []byte(`name: "test-server"
stdio:
  executablePath: "/usr/local/bin/test"
  arguments: ["--port", "8080"]
`), 0644)
			},
			wantLen:   1,
			wantErr:   nil,
			checkName: "test-server",
		},
		{
			name: "multiple valid files yaml and yml",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "first.yaml", []byte(`name: "first-server"
stdio:
  executablePath: "/usr/local/bin/first"
`), 0644)
				_ = afero.WriteFile(fs, "second.yml", []byte(`name: "second-server"
stdio:
  executablePath: "/usr/local/bin/second"
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
			wantErr: ErrNoMCPServersFound,
		},
		{
			name: "invalid yaml syntax",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "invalid.yaml", []byte(`name: "test"
stdio:
  executablePath: [invalid yaml structure
`), 0644)
			},
			wantLen: 0,
			wantErr: ErrParseFailed,
		},
		{
			name: "missing required field name",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "empty.yaml", []byte(`name: ""
stdio:
  executablePath: "/usr/local/bin/test"
`), 0644)
			},
			wantLen: 0,
			wantErr: ErrValidationFailed,
		},
		{
			name: "missing required field stdio",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "nostdio.yaml", []byte(`name: "test-server"
`), 0644)
			},
			wantLen: 0,
			wantErr: ErrValidationFailed,
		},
		{
			name: "missing required field executablePath",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "noexe.yaml", []byte(`name: "test-server"
stdio:
  executablePath: ""
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
				if tt.checkName != "" && len(got) > 0 {
					assert.Equal(t, tt.checkName, got[0].Name)
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
			wantErr: ErrNoMCPServersFound,
		},
		{
			name: "only directories",
			setupFs: func(fs afero.Fs) {
				_ = fs.Mkdir("subdir", 0755)
			},
			wantLen: 0,
			wantErr: ErrNoMCPServersFound,
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
			name: "read directory error",
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

func TestLoader_loadMCPServer(t *testing.T) {
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
				_ = afero.WriteFile(fs, "test.yaml", []byte(`name: "test-server"
stdio:
  executablePath: "/usr/local/bin/test"
  arguments: ["--port", "8080"]
`), 0644)
			},
			filePath: "test.yaml",
			wantName: "test-server",
			wantErr:  nil,
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
				_ = afero.WriteFile(fs, "bad.yaml", []byte(`name: [invalid
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
			filePath: "empty.yaml",
			wantName: "",
			wantErr:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			tt.setupFs(fs)

			loader := NewLoader(fs)
			got, err := loader.loadMCPServer(tt.filePath)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantName, got.Name)
			}
		})
	}
}

func TestLoader_validate(t *testing.T) {
	tests := []struct {
		name    string
		server  MCPServer
		wantErr error
	}{
		{
			name: "valid server",
			server: MCPServer{
				Name: "test-server",
				STDIO: &STDIOMCPServer{
					ExecutablePath: "/usr/local/bin/test",
				},
			},
			wantErr: nil,
		},
		{
			name: "empty name",
			server: MCPServer{
				Name: "",
				STDIO: &STDIOMCPServer{
					ExecutablePath: "/usr/local/bin/test",
				},
			},
			wantErr: ErrValidationFailed,
		},
		{
			name: "nil stdio",
			server: MCPServer{
				Name:  "test-server",
				STDIO: nil,
			},
			wantErr: ErrValidationFailed,
		},
		{
			name: "empty executablePath",
			server: MCPServer{
				Name: "test-server",
				STDIO: &STDIOMCPServer{
					ExecutablePath: "",
				},
			},
			wantErr: ErrValidationFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewLoader(afero.NewMemMapFs())
			err := loader.validate(tt.server)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
			} else {
				require.NoError(t, err)
			}
		})
	}
}
