package git

import (
	"fmt"
	"os"

	projectAPI "github.com/orbiqd/orbiqd-projectkit/pkg/project"
)

// NewGitFsProvider builds a provider function for the git filesystem.
func NewGitFsProvider() func(projectAPI.Fs) (Fs, error) {
	return func(projectFs projectAPI.Fs) (Fs, error) {
		currentDir, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("working directory lookup: %w", err)
		}

		gitFs, err := CreateGitFs(projectFs, currentDir)
		if err != nil {
			return nil, fmt.Errorf("git filesystem creation: %w", err)
		}

		return gitFs, nil
	}
}
