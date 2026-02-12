package mcp

import (
	"fmt"

	mcpAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/mcp"
	projectAPI "github.com/orbiqd/orbiqd-projectkit/pkg/project"
	"github.com/spf13/afero"
)

func NewFsRepositoryProvider() func(projectAPI.Fs) (mcpAPI.Repository, error) {
	return func(projectFs projectAPI.Fs) (mcpAPI.Repository, error) {
		dir := ".projectkit/repository/ai/mcp"

		if err := projectFs.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("mcp repository directory creation: %w", err)
		}

		scopedFs := afero.NewBasePathFs(projectFs, dir)
		return NewFsRepository(scopedFs), nil
	}
}
