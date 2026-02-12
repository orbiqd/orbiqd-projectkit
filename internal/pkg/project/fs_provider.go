package project

import (
	"fmt"
	"os"

	projectAPI "github.com/orbiqd/orbiqd-projectkit/pkg/project"
	"github.com/spf13/afero"
)

func NewProjectFsProvider() func() (projectAPI.Fs, error) {
	return func() (projectAPI.Fs, error) {
		currentDir, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("working directory lookup: %w", err)
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("home directory lookup: %w", err)
		}

		projectFs, err := CreateProjectFs(afero.NewOsFs(), currentDir, homeDir)
		if err != nil {
			return nil, fmt.Errorf("project filesystem creation: %w", err)
		}

		return projectAPI.Fs(projectFs), nil
	}
}
