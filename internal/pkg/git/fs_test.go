package git

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateGitFs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupFs    func(afero.Fs)
		currentDir string
		wantErr    error
		verifyRoot string
	}{
		{
			name: "WhenGitDirInCurrentDir_ThenReturnsFs",
			setupFs: func(fs afero.Fs) {
				require.NoError(t, fs.MkdirAll("/project/.git", 0755))
			},
			currentDir: "/project",
			wantErr:    nil,
			verifyRoot: "/project/.git",
		},
		{
			name: "WhenGitDirInParentDir_ThenReturnsFs",
			setupFs: func(fs afero.Fs) {
				require.NoError(t, fs.MkdirAll("/project/.git", 0755))
				require.NoError(t, fs.MkdirAll("/project/sub", 0755))
			},
			currentDir: "/project/sub",
			wantErr:    nil,
			verifyRoot: "/project/.git",
		},
		{
			name: "WhenGitDirInCurrentDirEvenIfHome_ThenReturnsFs",
			setupFs: func(fs afero.Fs) {
				require.NoError(t, fs.MkdirAll("/home/.git", 0755))
			},
			currentDir: "/home",
			wantErr:    nil,
			verifyRoot: "/home/.git",
		},
		{
			name: "WhenGitDirInAncestorDir_ThenReturnsFs",
			setupFs: func(fs afero.Fs) {
				require.NoError(t, fs.MkdirAll("/home/.git", 0755))
				require.NoError(t, fs.MkdirAll("/home/project/sub", 0755))
			},
			currentDir: "/home/project/sub",
			wantErr:    nil,
			verifyRoot: "/home/.git",
		},
		{
			name: "WhenNoGitDirAndReachesRoot_ThenReturnsError",
			setupFs: func(fs afero.Fs) {
				require.NoError(t, fs.MkdirAll("/project/sub", 0755))
			},
			currentDir: "/project/sub",
			wantErr:    ErrGitRepoNotFound,
			verifyRoot: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fs := afero.NewMemMapFs()
			tt.setupFs(fs)

			result, err := CreateGitFs(fs, tt.currentDir)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, result)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			testFileName := "config"
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
