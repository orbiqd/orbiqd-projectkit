package project

import (
	"errors"
	"path/filepath"

	"github.com/spf13/afero"
)

func CreateProjectFs(fs afero.Fs, currentDir string, homeDir string) (afero.Fs, error) {
	for {
		if currentDir == homeDir {
			return nil, ErrProjectRootNotFound
		}

		gitDir := filepath.Join(currentDir, ".git")
		configFile := filepath.Join(currentDir, ConfigFileName)

		gitExists, _ := afero.DirExists(fs, gitDir)
		if gitExists {
			return afero.NewBasePathFs(fs, currentDir), nil
		}

		configExists, _ := afero.Exists(fs, configFile)
		if configExists {
			return afero.NewBasePathFs(fs, currentDir), nil
		}

		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			return nil, ErrProjectRootNotFound
		}

		currentDir = parent
	}
}

var ErrProjectRootNotFound = errors.New("project root path not found")
