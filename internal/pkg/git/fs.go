package git

import (
	"errors"
	"path/filepath"

	"github.com/spf13/afero"
)

// Fs represents the git filesystem abstraction.
type Fs interface {
	afero.Fs
}

// CreateGitFs returns a filesystem scoped to the closest .git directory.
func CreateGitFs(projectFs afero.Fs, currentDir string) (afero.Fs, error) {
	for {
		gitDir := filepath.Join(currentDir, ".git")

		gitExists, _ := afero.DirExists(projectFs, gitDir)
		if gitExists {
			return afero.NewBasePathFs(projectFs, gitDir), nil
		}

		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			return nil, ErrGitRepoNotFound
		}

		currentDir = parent
	}
}

// ErrGitRepoNotFound indicates that no git repository was found in any parent directory.
var ErrGitRepoNotFound = errors.New("git repository not found")
