package action

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validStandardYAML(t *testing.T) string {
	t.Helper()

	return `metadata:
  id: test-standard
  name: Test Standard
  version: 1.0.0
  tags:
    - test-tag
  scope:
    languages:
      - en
  relations:
    standard: []
specification:
  purpose: Test purpose for the standard
  goals:
    - First goal of the standard
requirements:
  rules:
    - level: must
      statement: This is a test requirement
      rationale: This is the rationale for the requirement
examples:
  good:
    - title: Good Example
      language: go
      snippet: |
        package main
        func main() {}
      reason: This is a good example because it follows the standard
`
}

func TestValidateDocStandardActionRun_WhenValidYAML_ThenReturnsNil(t *testing.T) {
	t.Parallel()

	readFile := func(path string) ([]byte, error) {
		return []byte(validStandardYAML(t)), nil
	}

	action := NewValidateDocStandardAction("test.yaml", readFile)

	err := action.Run()

	require.NoError(t, err)
}

func TestValidateDocStandardActionRun_WhenReadFileFails_ThenReturnsReadFileError(t *testing.T) {
	t.Parallel()

	readFileErr := errors.New("read file error")
	readFile := func(path string) ([]byte, error) {
		return nil, readFileErr
	}

	action := NewValidateDocStandardAction("test.yaml", readFile)

	err := action.Run()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read file")
	assert.ErrorIs(t, err, readFileErr)
}

func TestValidateDocStandardActionRun_WhenYAMLInvalid_ThenReturnsParseError(t *testing.T) {
	t.Parallel()

	readFile := func(path string) ([]byte, error) {
		return []byte("invalid: yaml: content: [[["), nil
	}

	action := NewValidateDocStandardAction("test.yaml", readFile)

	err := action.Run()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse YAML")
}

func TestValidateDocStandardActionRun_WhenValidationFails_ThenReturnsValidationError(t *testing.T) {
	t.Parallel()

	invalidYAML := `metadata:
  version: 1.0.0
specification:
  purpose: Test purpose
requirements:
  rules: []
examples:
  good: []
`

	readFile := func(path string) ([]byte, error) {
		return []byte(invalidYAML), nil
	}

	action := NewValidateDocStandardAction("test.yaml", readFile)

	err := action.Run()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}
