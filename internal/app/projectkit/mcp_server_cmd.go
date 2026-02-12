package projectkit

import (
	"github.com/mark3labs/mcp-go/server"
)

type MCPServerCmd struct {
}

func (cmd *MCPServerCmd) Run() error {
	mcp := server.NewMCPServer(
		"projectkit",
		"1.0.0",
	)

	return server.ServeStdio(mcp)
}
