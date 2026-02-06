package projectkit

import "github.com/orbiqd/orbiqd-projectkit/internal/pkg/log"

type Cmd struct {
	Log log.Config `embed:"true" prefix:"log-"`

	Setup SetupCmd `cmd:"setup" help:"Setup projectkit into current project."`
	Doc   DocCmd   `cmd:"doc" help:"Documentation-related commands."`
}
