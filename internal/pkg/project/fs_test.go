package project

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateProjectFs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupFs    func(afero.Fs)
		currentDir string
		homeDir    string
		wantErr    error
		verifyRoot string
	}{
		{
			name: "WhenGitDirInCurrentDir_ThenReturnsFs",
			setupFs: func(fs afero.Fs) {
				require.NoError(t, fs.MkdirAll("/project/.git", 0755))
			},
			currentDir: "/project",
			homeDir:    "/home",
			wantErr:    nil,
			verifyRoot: "/project",
		},
		{
			name: "WhenConfigFileInCurrentDir_ThenReturnsFs",
			setupFs: func(fs afero.Fs) {
				require.NoError(t, fs.MkdirAll("/project", 0755))
				require.NoError(t, afero.WriteFile(fs, "/project/.projectkit.yaml", []byte("test"), 0644))
			},
			currentDir: "/project",
			homeDir:    "/home",
			wantErr:    nil,
			verifyRoot: "/project",
		},
		{
			name: "WhenGitDirInParentDir_ThenReturnsFs",
			setupFs: func(fs afero.Fs) {
				require.NoError(t, fs.MkdirAll("/project/.git", 0755))
				require.NoError(t, fs.MkdirAll("/project/sub", 0755))
			},
			currentDir: "/project/sub",
			homeDir:    "/home",
			wantErr:    nil,
			verifyRoot: "/project",
		},
		{
			name: "WhenConfigFileInParentDir_ThenReturnsFs",
			setupFs: func(fs afero.Fs) {
				require.NoError(t, fs.MkdirAll("/project/sub/deep", 0755))
				require.NoError(t, afero.WriteFile(fs, "/project/.projectkit.yaml", []byte("test"), 0644))
			},
			currentDir: "/project/sub/deep",
			homeDir:    "/home",
			wantErr:    nil,
			verifyRoot: "/project",
		},
		{
			name: "WhenBothMarkersExist_ThenReturnsFs",
			setupFs: func(fs afero.Fs) {
				require.NoError(t, fs.MkdirAll("/project/.git", 0755))
				require.NoError(t, afero.WriteFile(fs, "/project/.projectkit.yaml", []byte("test"), 0644))
			},
			currentDir: "/project",
			homeDir:    "/home",
			wantErr:    nil,
			verifyRoot: "/project",
		},
		{
			name: "WhenCurrentDirIsHomeDir_ThenReturnsError",
			setupFs: func(fs afero.Fs) {
				require.NoError(t, fs.MkdirAll("/home/.git", 0755))
			},
			currentDir: "/home",
			homeDir:    "/home",
			wantErr:    ErrProjectRootNotFound,
			verifyRoot: "",
		},
		{
			name: "WhenTraversingReachesHomeDir_ThenReturnsError",
			setupFs: func(fs afero.Fs) {
				require.NoError(t, fs.MkdirAll("/home/project/sub", 0755))
			},
			currentDir: "/home/project/sub",
			homeDir:    "/home",
			wantErr:    ErrProjectRootNotFound,
			verifyRoot: "",
		},
		{
			name: "WhenNoMarkersAndReachesRoot_ThenReturnsError",
			setupFs: func(fs afero.Fs) {
				require.NoError(t, fs.MkdirAll("/project/sub", 0755))
			},
			currentDir: "/project/sub",
			homeDir:    "/home",
			wantErr:    ErrProjectRootNotFound,
			verifyRoot: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fs := afero.NewMemMapFs()
			tt.setupFs(fs)

			result, err := CreateProjectFs(fs, tt.currentDir, tt.homeDir)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, result)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			testFileName := "test-file.txt"
			testContent := []byte("test content")
			testFilePath := filepath.Join(tt.verifyRoot, testFileName)

			require.NoError(t, afero.WriteFile(fs, testFilePath, testContent, 0644))

			exists, err := afero.Exists(result, testFileName)
			require.NoError(t, err)
			assert.True(t, exists)

			content, err := afero.ReadFile(result, testFileName)
			require.NoError(t, err)
			assert.Equal(t, testContent, content)
		})
	}
}
