package git

import (
	"errors"
	"io/fs"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.nhat.io/aferomock"
)

func writeExcludeFile(t *testing.T, fs afero.Fs, content string) {
	t.Helper()
	require.NoError(t, fs.MkdirAll("info", 0755))
	require.NoError(t, afero.WriteFile(fs, "info/exclude", []byte(content), 0644))
}

func TestMatchesPattern(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		line     string
		pattern  string
		expected bool
	}{
		{
			name:     "WhenExactMatch_ThenReturnsTrue",
			line:     ".ai",
			pattern:  ".ai",
			expected: true,
		},
		{
			name:     "WhenLineHasTrailingSpaces_ThenReturnsTrue",
			line:     ".ai   ",
			pattern:  ".ai",
			expected: true,
		},
		{
			name:     "WhenLineHasTrailingTabs_ThenReturnsTrue",
			line:     ".ai\t\t",
			pattern:  ".ai",
			expected: true,
		},
		{
			name:     "WhenLineHasTrailingMixedWhitespace_ThenReturnsTrue",
			line:     ".ai \t ",
			pattern:  ".ai",
			expected: true,
		},
		{
			name:     "WhenLineHasLeadingSpaces_ThenReturnsFalse",
			line:     "  .ai",
			pattern:  ".ai",
			expected: false,
		},
		{
			name:     "WhenLineDiffers_ThenReturnsFalse",
			line:     ".env",
			pattern:  ".ai",
			expected: false,
		},
		{
			name:     "WhenBothEmpty_ThenReturnsTrue",
			line:     "",
			pattern:  "",
			expected: true,
		},
		{
			name:     "WhenLineEmptyPatternNot_ThenReturnsFalse",
			line:     "",
			pattern:  ".ai",
			expected: false,
		},
		{
			name:     "WhenLineNotEmptyPatternEmpty_ThenReturnsFalse",
			line:     ".ai",
			pattern:  "",
			expected: false,
		},
		{
			name:     "WhenLineIsOnlyWhitespace_ThenMatchesEmpty",
			line:     "   ",
			pattern:  "",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := matchesPattern(tt.line, tt.pattern)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestReadExcludeLines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupFile     bool
		fileContent   string
		expectedLines []string
	}{
		{
			name:          "WhenFileNotExists_ThenReturnsNil",
			setupFile:     false,
			fileContent:   "",
			expectedLines: nil,
		},
		{
			name:          "WhenFileEmpty_ThenReturnsSingleEmptyString",
			setupFile:     true,
			fileContent:   "",
			expectedLines: []string{""},
		},
		{
			name:          "WhenSingleLine_ThenReturnsLines",
			setupFile:     true,
			fileContent:   ".ai\n",
			expectedLines: []string{".ai", ""},
		},
		{
			name:          "WhenMultipleLines_ThenReturnsAll",
			setupFile:     true,
			fileContent:   ".ai\n.env\n",
			expectedLines: []string{".ai", ".env", ""},
		},
		{
			name:          "WhenNoTrailingNewline_ThenReturnsLines",
			setupFile:     true,
			fileContent:   ".ai",
			expectedLines: []string{".ai"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fs := afero.NewMemMapFs()
			if tt.setupFile {
				writeExcludeFile(t, fs, tt.fileContent)
			}

			result, err := readExcludeLines(fs)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedLines, result)
		})
	}
}

func TestReadExcludeLines_WhenStatFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	statErr := errors.New("stat failed")
	base := afero.NewMemMapFs()
	fs := aferomock.OverrideFs(base, aferomock.FsCallbacks{
		StatFunc: func(name string) (fs.FileInfo, error) {
			if name == "info/exclude" {
				return nil, statErr
			}
			return base.Stat(name)
		},
	})

	result, err := readExcludeLines(fs)

	require.ErrorIs(t, err, statErr)
	assert.Nil(t, result)
}

func TestReadExcludeLines_WhenReadFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	readErr := errors.New("read failed")
	base := afero.NewMemMapFs()
	writeExcludeFile(t, base, ".ai\n")

	fs := aferomock.OverrideFs(base, aferomock.FsCallbacks{
		OpenFunc: func(name string) (afero.File, error) {
			if name == "info/exclude" {
				return nil, readErr
			}
			return base.Open(name)
		},
	})

	result, err := readExcludeLines(fs)

	require.ErrorIs(t, err, readErr)
	assert.Nil(t, result)
}

func TestIsExcluded(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupFile   bool
		fileContent string
		pattern     string
		expected    bool
	}{
		{
			name:        "WhenPatternFound_ThenReturnsTrue",
			setupFile:   true,
			fileContent: ".ai\n.env\n",
			pattern:     ".ai",
			expected:    true,
		},
		{
			name:        "WhenPatternNotFound_ThenReturnsFalse",
			setupFile:   true,
			fileContent: ".ai\n.env\n",
			pattern:     ".tmp",
			expected:    false,
		},
		{
			name:        "WhenFileNotExists_ThenReturnsFalse",
			setupFile:   false,
			fileContent: "",
			pattern:     ".ai",
			expected:    false,
		},
		{
			name:        "WhenPatternWithTrailingSpacesInFile_ThenReturnsTrue",
			setupFile:   true,
			fileContent: ".ai   \n",
			pattern:     ".ai",
			expected:    true,
		},
		{
			name:        "WhenFileEmpty_ThenReturnsFalse",
			setupFile:   true,
			fileContent: "",
			pattern:     ".ai",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fs := afero.NewMemMapFs()
			if tt.setupFile {
				writeExcludeFile(t, fs, tt.fileContent)
			}

			result, err := IsExcluded(fs, tt.pattern)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsExcluded_WhenReadError_ThenReturnsError(t *testing.T) {
	t.Parallel()

	statErr := errors.New("stat failed")
	base := afero.NewMemMapFs()
	fs := aferomock.OverrideFs(base, aferomock.FsCallbacks{
		StatFunc: func(name string) (fs.FileInfo, error) {
			if name == "info/exclude" {
				return nil, statErr
			}
			return base.Stat(name)
		},
	})

	result, err := IsExcluded(fs, ".ai")

	require.ErrorIs(t, err, statErr)
	assert.False(t, result)
}

func TestExclude(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		setupFile       bool
		initialContent  string
		pattern         string
		expectedContent string
	}{
		{
			name:            "WhenFileNotExists_ThenCreatesFileWithPattern",
			setupFile:       false,
			initialContent:  "",
			pattern:         ".ai",
			expectedContent: ".ai\n",
		},
		{
			name:            "WhenFileEmpty_ThenAddsPattern",
			setupFile:       true,
			initialContent:  "",
			pattern:         ".ai",
			expectedContent: "\n.ai\n",
		},
		{
			name:            "WhenPatternAlreadyExists_ThenNoOp",
			setupFile:       true,
			initialContent:  ".ai\n",
			pattern:         ".ai",
			expectedContent: ".ai\n",
		},
		{
			name:            "WhenOtherPatternsExist_ThenAppendsPattern",
			setupFile:       true,
			initialContent:  ".env\n",
			pattern:         ".ai",
			expectedContent: ".env\n.ai\n",
		},
		{
			name:            "WhenFileHasNoTrailingNewline_ThenAddsNewlineBeforePattern",
			setupFile:       true,
			initialContent:  ".env",
			pattern:         ".ai",
			expectedContent: ".env\n.ai\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fs := afero.NewMemMapFs()
			if tt.setupFile {
				writeExcludeFile(t, fs, tt.initialContent)
			}

			err := Exclude(fs, tt.pattern)

			require.NoError(t, err)

			content, err := afero.ReadFile(fs, "info/exclude")
			require.NoError(t, err)
			assert.Equal(t, tt.expectedContent, string(content))
		})
	}
}

func TestExclude_WhenMkdirAllFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mkdirErr := errors.New("mkdir failed")
	base := afero.NewMemMapFs()
	fs := aferomock.OverrideFs(base, aferomock.FsCallbacks{
		MkdirAllFunc: func(path string, perm fs.FileMode) error {
			return mkdirErr
		},
	})

	err := Exclude(fs, ".ai")

	require.ErrorIs(t, err, mkdirErr)
}

func TestExclude_WhenWriteFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	writeErr := errors.New("write failed")
	base := afero.NewMemMapFs()
	fs := aferomock.OverrideFs(base, aferomock.FsCallbacks{
		OpenFileFunc: func(name string, flag int, perm fs.FileMode) (afero.File, error) {
			return nil, writeErr
		},
	})

	err := Exclude(fs, ".ai")

	require.ErrorIs(t, err, writeErr)
}

func TestExclude_WhenReadError_ThenReturnsError(t *testing.T) {
	t.Parallel()

	statErr := errors.New("stat failed")
	base := afero.NewMemMapFs()
	fs := aferomock.OverrideFs(base, aferomock.FsCallbacks{
		StatFunc: func(name string) (fs.FileInfo, error) {
			if name == "info/exclude" {
				return nil, statErr
			}
			return base.Stat(name)
		},
	})

	err := Exclude(fs, ".ai")

	require.ErrorIs(t, err, statErr)
}

func TestUnexclude(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		setupFile       bool
		initialContent  string
		pattern         string
		expectedContent string
		fileExists      bool
	}{
		{
			name:            "WhenPatternExists_ThenRemovesIt",
			setupFile:       true,
			initialContent:  ".ai\n.env\n",
			pattern:         ".ai",
			expectedContent: ".env\n",
			fileExists:      true,
		},
		{
			name:            "WhenPatternNotFound_ThenNoOp",
			setupFile:       true,
			initialContent:  ".env\n",
			pattern:         ".ai",
			expectedContent: ".env\n",
			fileExists:      true,
		},
		{
			name:            "WhenFileNotExists_ThenNoOp",
			setupFile:       false,
			initialContent:  "",
			pattern:         ".ai",
			expectedContent: "",
			fileExists:      false,
		},
		{
			name:            "WhenMultipleOccurrences_ThenRemovesAll",
			setupFile:       true,
			initialContent:  ".ai\n.env\n.ai\n",
			pattern:         ".ai",
			expectedContent: ".env\n",
			fileExists:      true,
		},
		{
			name:            "WhenPatternWithTrailingSpaces_ThenRemovesMatching",
			setupFile:       true,
			initialContent:  ".ai   \n.env\n",
			pattern:         ".ai",
			expectedContent: ".env\n",
			fileExists:      true,
		},
		{
			name:            "WhenOnlyPattern_ThenFileBecomesEmpty",
			setupFile:       true,
			initialContent:  ".ai\n",
			pattern:         ".ai",
			expectedContent: "",
			fileExists:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fs := afero.NewMemMapFs()
			if tt.setupFile {
				writeExcludeFile(t, fs, tt.initialContent)
			}

			err := Unexclude(fs, tt.pattern)

			require.NoError(t, err)

			if tt.fileExists {
				content, err := afero.ReadFile(fs, "info/exclude")
				require.NoError(t, err)
				assert.Equal(t, tt.expectedContent, string(content))
			} else {
				exists, err := afero.Exists(fs, "info/exclude")
				require.NoError(t, err)
				assert.False(t, exists)
			}
		})
	}
}

func TestUnexclude_WhenReadError_ThenReturnsError(t *testing.T) {
	t.Parallel()

	statErr := errors.New("stat failed")
	base := afero.NewMemMapFs()
	fs := aferomock.OverrideFs(base, aferomock.FsCallbacks{
		StatFunc: func(name string) (fs.FileInfo, error) {
			if name == "info/exclude" {
				return nil, statErr
			}
			return base.Stat(name)
		},
	})

	err := Unexclude(fs, ".ai")

	require.ErrorIs(t, err, statErr)
}

func TestUnexclude_WhenWriteFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	writeErr := errors.New("write failed")
	base := afero.NewMemMapFs()
	writeExcludeFile(t, base, ".ai\n.env\n")

	fs := aferomock.OverrideFs(base, aferomock.FsCallbacks{
		OpenFileFunc: func(name string, flag int, perm fs.FileMode) (afero.File, error) {
			return nil, writeErr
		},
	})

	err := Unexclude(fs, ".ai")

	require.ErrorIs(t, err, writeErr)
}
