package projectkit

import (
	"github.com/orbiqd/orbiqd-projectkit/internal/app/projectkit/action"
	"github.com/orbiqd/orbiqd-projectkit/internal/pkg/doc/standard"
	standardAPI "github.com/orbiqd/orbiqd-projectkit/pkg/doc/standard"
	projectAPI "github.com/orbiqd/orbiqd-projectkit/pkg/project"
)

type DocStandardRenderCmd struct{}

func (cmd *DocStandardRenderCmd) Run(
	config *projectAPI.Config,
	projectFs projectAPI.Fs,
	standardRepository standardAPI.Repository,
) error {
	renderers := map[string]standardAPI.Renderer{
		"markdown": standard.NewMarkdownRenderer(),
	}
	return action.NewRenderDocStandardAction(standardRepository, projectFs, config.Docs.Standard.Render, renderers).Run()
}
