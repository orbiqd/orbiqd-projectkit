package projectkit

import (
	"context"
	"os"

	"github.com/orbiqd/orbiqd-projectkit/internal/app/projectkit/action"
)

type DocStandardValidateCmd struct {
	Path string `arg:"" help:"Path to the standard YAML file." type:"existingfile"`
}

func (cmd *DocStandardValidateCmd) Run(ctx context.Context) error {
	return action.NewValidateDocStandardAction(cmd.Path, os.ReadFile).Run()
}
