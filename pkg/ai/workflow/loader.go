package workflow

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	"sigs.k8s.io/yaml"
)

type Loader struct {
	rootFs afero.Fs
}

func NewLoader(rootFs afero.Fs) *Loader {
	return &Loader{
		rootFs: rootFs,
	}
}

func (loader *Loader) Load() ([]Workflow, error) {
	workflowPaths, err := loader.resolvePaths()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve paths: %w", err)
	}

	var workflows []Workflow
	for _, workflowPath := range workflowPaths {
		workflow, err := loader.loadWorkflow(workflowPath)
		if err != nil {
			return []Workflow{}, fmt.Errorf("failed to load workflow: %s: %w", workflowPath, err)
		}

		workflows = append(workflows, *workflow)
	}

	if len(workflows) == 0 {
		return []Workflow{}, ErrNoWorkflowsFound
	}

	return workflows, nil
}

func (loader *Loader) resolvePaths() ([]string, error) {
	entries, err := afero.ReadDir(loader.rootFs, ".")
	if err != nil {
		return nil, fmt.Errorf("read directory: %w", err)
	}

	var filePaths []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext == ".yaml" || ext == ".yml" {
			filePaths = append(filePaths, entry.Name())
		}
	}

	return filePaths, nil
}

func (loader *Loader) loadWorkflow(path string) (*Workflow, error) {
	data, err := afero.ReadFile(loader.rootFs, path)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrReadFailed, err)
	}

	var workflow Workflow
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParseFailed, err)
	}

	err = Validate(workflow)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrValidationFailed, err)
	}

	return &workflow, nil
}

var ErrNoWorkflowsFound = errors.New("no workflows found")
var ErrReadFailed = errors.New("read failed")
var ErrParseFailed = errors.New("parse failed")
var ErrValidationFailed = errors.New("validation failed")
