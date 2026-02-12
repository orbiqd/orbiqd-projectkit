package action

import (
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"

	standardAPI "github.com/orbiqd/orbiqd-projectkit/pkg/doc/standard"
	projectAPI "github.com/orbiqd/orbiqd-projectkit/pkg/project"
	"github.com/spf13/afero"
)

var ErrUnsupportedFormat = errors.New("unsupported render format")

type RenderDocStandardAction struct {
	standardRepository standardAPI.Repository
	projectFs          projectAPI.Fs
	renderConfigs      []standardAPI.RenderConfig
	renderers          map[string]standardAPI.Renderer
}

func NewRenderDocStandardAction(
	standardRepository standardAPI.Repository,
	projectFs projectAPI.Fs,
	renderConfigs []standardAPI.RenderConfig,
	renderers map[string]standardAPI.Renderer,
) *RenderDocStandardAction {
	return &RenderDocStandardAction{
		standardRepository: standardRepository,
		projectFs:          projectFs,
		renderConfigs:      renderConfigs,
		renderers:          renderers,
	}
}

func (action *RenderDocStandardAction) Run() error {
	standards, err := action.standardRepository.GetAll()
	if err != nil {
		return fmt.Errorf("failed to get standards: %w", err)
	}
	slog.Info("Loaded standards from repository.", slog.Int("count", len(standards)))

	for _, renderConfig := range action.renderConfigs {
		if err := action.renderToDestination(renderConfig, standards); err != nil {
			return err
		}
	}

	return nil
}

func (action *RenderDocStandardAction) renderToDestination(
	renderConfig standardAPI.RenderConfig,
	standards []standardAPI.Standard,
) error {
	renderer, err := action.createRenderer(renderConfig)
	if err != nil {
		return err
	}

	if err := action.cleanDestination(renderConfig, renderer.FileExtension()); err != nil {
		return fmt.Errorf("failed to clean destination: %w", err)
	}

	for _, std := range standards {
		rendered, err := renderer.Render(std)
		if err != nil {
			return fmt.Errorf("failed to render standard %s: %w", std.Metadata.Name, err)
		}

		filename := string(std.Metadata.Id) + renderer.FileExtension()
		filePath := filepath.Join(renderConfig.Destination, filename)

		if err := afero.WriteFile(action.projectFs, filePath, rendered, 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", filePath, err)
		}

		slog.Debug("Rendered standard to destination.", slog.String("filePath", filePath))
	}

	slog.Info("Rendered standards to destination.",
		slog.Int("count", len(standards)),
		slog.String("destination", renderConfig.Destination),
	)

	return nil
}

func (action *RenderDocStandardAction) createRenderer(renderConfig standardAPI.RenderConfig) (standardAPI.Renderer, error) {
	r, ok := action.renderers[renderConfig.Format]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedFormat, renderConfig.Format)
	}
	return r, nil
}

func (action *RenderDocStandardAction) cleanDestination(renderConfig standardAPI.RenderConfig, fileExtension string) error {
	exists, err := afero.DirExists(action.projectFs, renderConfig.Destination)
	if err != nil {
		return err
	}

	if !exists {
		return action.projectFs.MkdirAll(renderConfig.Destination, 0755)
	}

	files, err := afero.ReadDir(action.projectFs, renderConfig.Destination)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if filepath.Ext(file.Name()) == fileExtension {
			filePath := filepath.Join(renderConfig.Destination, file.Name())
			if err := action.projectFs.Remove(filePath); err != nil {
				return fmt.Errorf("failed to remove file %s: %w", filePath, err)
			}
		}
	}

	return nil
}
