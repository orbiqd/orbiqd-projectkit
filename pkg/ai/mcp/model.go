package mcp

type STDIOMCPServer struct {
	ExecutablePath       string            `json:"executablePath"`
	Arguments            []string          `json:"arguments"`
	EnvironmentVariables map[string]string `json:"environmentVariables"`
}

type MCPServer struct {
	Name  string          `json:"name"`
	STDIO *STDIOMCPServer `json:"stdio"`
}
