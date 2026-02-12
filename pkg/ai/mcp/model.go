package mcp

type STDIOMCPServer struct {
	ExecutablePath       string            `json:"executablePath" validate:"required"`
	Arguments            []string          `json:"arguments"`
	EnvironmentVariables map[string]string `json:"environmentVariables"`
}

type MCPServer struct {
	Name  string          `json:"name" validate:"required"`
	STDIO *STDIOMCPServer `json:"stdio" validate:"required"`
}
