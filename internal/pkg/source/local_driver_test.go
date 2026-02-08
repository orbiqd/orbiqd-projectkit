package source

import (
	"errors"
	"os"
	"testing"

	sourceAPI "github.com/orbiqd/orbiqd-projectkit/pkg/source"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.nhat.io/aferomock"
)

func TestLocalDriver_Resolve(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupFs     func(afero.Fs)
		uri         string
		expectError bool
		errorMsg    string
		errorIs     error
	}{
		{
			name: "WhenRelativePath_ThenReturnsReadOnlyFs",
			setupFs: func(fs afero.Fs) {
				_ = fs.MkdirAll("data/config", 0o755)
			},
			uri:         "local://data/config",
			expectError: false,
		},
		{
			name: "WhenAbsolutePath_ThenReturnsReadOnlyFs",
			setupFs: func(fs afero.Fs) {
				_ = fs.MkdirAll("/tmp/data", 0o755)
			},
			uri:         "local:///tmp/data",
			expectError: false,
		},
		{
			name: "WhenPathNotExists_ThenReturnsError",
			setupFs: func(fs afero.Fs) {
			},
			uri:         "local://nonexistent",
			expectError: true,
			errorMsg:    "does not exist",
		},
		{
			name: "WhenEmptyPath_ThenReturnsError",
			setupFs: func(fs afero.Fs) {
			},
			uri:         "local://",
			expectError: true,
			errorMsg:    "empty path",
		},
		{
			name: "WhenInvalidScheme_ThenReturnsError",
			setupFs: func(fs afero.Fs) {
			},
			uri:         "http://data",
			expectError: true,
			errorIs:     sourceAPI.ErrUnsupportedScheme,
		},
		{
			name: "WhenPathWithDots_ThenNormalizesPath",
			setupFs: func(fs afero.Fs) {
				_ = fs.MkdirAll("data/sub", 0o755)
			},
			uri:         "local://data/sub/../",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			memFs := afero.NewMemMapFs()
			tt.setupFs(memFs)

			driver := NewLocalDriver(WithRootFs(memFs))

			result, err := driver.Resolve(tt.uri)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				if tt.errorIs != nil {
					assert.ErrorIs(t, err, tt.errorIs)
				}
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)

				_, ok := result.(*afero.ReadOnlyFs)
				assert.True(t, ok, "result should be ReadOnlyFs")

				writeErr := afero.WriteFile(result, "test.txt", []byte("content"), 0o644)
				assert.Error(t, writeErr, "write should fail on read-only filesystem")
			}
		})
	}
}

func TestLocalDriver_Resolve_WhenStatError_ThenReturnsError(t *testing.T) {
	t.Parallel()

	baseFs := afero.NewMemMapFs()
	errFs := aferomock.OverrideFs(baseFs, aferomock.FsCallbacks{
		StatFunc: func(name string) (os.FileInfo, error) {
			return nil, errors.New("permission denied")
		},
	})
	driver := NewLocalDriver(WithRootFs(errFs))

	result, err := driver.Resolve("local://somepath")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "checking path")
	assert.Contains(t, err.Error(), "permission denied")
	assert.Nil(t, result)
}

func TestLocalDriver_GetSupportedSchemes(t *testing.T) {
	t.Parallel()

	driver := NewLocalDriver()

	schemes := driver.GetSupportedSchemes()

	assert.Equal(t, []string{"local"}, schemes)
}
