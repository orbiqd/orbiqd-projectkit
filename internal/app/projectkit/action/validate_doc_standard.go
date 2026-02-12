package action

import (
	"errors"
	"fmt"

	standardAPI "github.com/orbiqd/orbiqd-projectkit/pkg/doc/standard"
	"sigs.k8s.io/yaml"
)

type ValidateDocStandardAction struct {
	path     string
	readFile func(string) ([]byte, error)
}

func NewValidateDocStandardAction(path string, readFile func(string) ([]byte, error)) *ValidateDocStandardAction {
	return &ValidateDocStandardAction{
		path:     path,
		readFile: readFile,
	}
}

func (action *ValidateDocStandardAction) Run() error {
	data, err := action.readFile(action.path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var standard standardAPI.Standard
	if err := yaml.Unmarshal(data, &standard); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := standardAPI.Validate(standard); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}

var ErrValidationFailed = errors.New("validation failed")
