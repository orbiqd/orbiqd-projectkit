package standard

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/spf13/afero"
	"sigs.k8s.io/yaml"
)

type Loader struct {
	fs afero.Fs
}

func NewLoader(fs afero.Fs) *Loader {
	return &Loader{
		fs: fs,
	}
}

func (loader *Loader) Load() ([]Standard, error) {
	files, err := loader.resolveFiles()
	if err != nil {
		return nil, fmt.Errorf("resolve files: %w", err)
	}

	var standards []Standard

	for _, filePath := range files {
		standard, err := loader.loadStandard(filePath)
		if err != nil {
			return nil, fmt.Errorf("load standard: %s: %w", filePath, err)
		}

		err = Validate(*standard)
		if err != nil {
			return nil, fmt.Errorf("validate standard: %s: %w", filePath, err)
		}

		standards = append(standards, *standard)
	}

	if len(standards) == 0 {
		return nil, ErrNoStandardsFound
	}

	return standards, nil
}

func (loader *Loader) resolveFiles() ([]string, error) {
	entries, err := afero.ReadDir(loader.fs, ".")
	if err != nil {
		return nil, fmt.Errorf("read directory: %w", err)
	}

	var filePaths []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := filepath.Ext(entry.Name())
		if ext == ".yaml" {
			filePaths = append(filePaths, entry.Name())
		}
	}

	return filePaths, nil
}

func (loader *Loader) loadStandard(path string) (*Standard, error) {
	data, err := afero.ReadFile(loader.fs, path)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrReadFailed, err)
	}

	var standard Standard
	if err := yaml.Unmarshal(data, &standard); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrParseFailed, err)
	}

	return &standard, nil
}

var ErrNoStandardsFound = errors.New("no standards found")
var ErrReadFailed = errors.New("read failed")
var ErrParseFailed = errors.New("parse failed")
