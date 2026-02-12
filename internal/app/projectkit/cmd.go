package projectkit

import "github.com/orbiqd/orbiqd-projectkit/internal/pkg/log"

type Cmd struct {
	Log log.Config `embed:"true" prefix:"log-"`

	MCP    MCPCmd    `cmd:"mcp" help:"MCP-related commands."`
	Update UpdateCmd `cmd:"update" help:"Update projectkit repositories from config."`
	Render RenderCmd `cmd:"render" help:"Render all configurations."`
	Agent  AgentCmd  `cmd:"agent" help:"Agent-related commands."`
	Doc    DocCmd    `cmd:"doc" help:"Documentation-related commands."`
}
