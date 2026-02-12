package mcp

type Repository interface {
	GetAll() ([]MCPServer, error)
	AddMCPServer(server MCPServer) error
	RemoveAll() error
}
