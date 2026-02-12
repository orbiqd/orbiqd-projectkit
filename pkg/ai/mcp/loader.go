package mcp

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/go-playground/validator/v10"
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

func (loader *Loader) Load() ([]MCPServer, error) {
	filePaths, err := loader.resolveFiles()
	if err != nil {
		return nil, err
	}

	var result []MCPServer
	for _, filePath := range filePaths {
		server, err := loader.loadMCPServer(filePath)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", filePath, err)
		}

		if err := loader.validate(*server); err != nil {
			return nil, fmt.Errorf("%s: %w", filePath, err)
		}

		result = append(result, *server)
	}

	return result, nil
}

func (loader *Loader) validate(server MCPServer) error {
	validate := validator.New()

	if err := validate.Struct(server); err != nil {
		return fmt.Errorf("%w: %v", ErrValidationFailed, err)
	}

	return nil
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
		if ext == ".yaml" || ext == ".yml" {
			filePaths = append(filePaths, entry.Name())
		}
	}

	if len(filePaths) == 0 {
		return nil, ErrNoMCPServersFound
	}

	return filePaths, nil
}

func (loader *Loader) loadMCPServer(filePath string) (*MCPServer, error) {
	data, err := afero.ReadFile(loader.fs, filePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrReadFailed, err)
	}

	var server MCPServer
	if err := yaml.Unmarshal(data, &server); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParseFailed, err)
	}

	return &server, nil
}

var ErrNoMCPServersFound = errors.New("no mcp servers found")
var ErrParseFailed = errors.New("parse failed")
var ErrReadFailed = errors.New("read failed")
var ErrValidationFailed = errors.New("validation failed")
