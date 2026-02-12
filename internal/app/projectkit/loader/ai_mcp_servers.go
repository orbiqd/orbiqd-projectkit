package loader

import (
	"fmt"

	mcpAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/mcp"
	sourceAPI "github.com/orbiqd/orbiqd-projectkit/pkg/source"
)

func LoadAiMCPServersFromConfig(config mcpAPI.Config, sourceResolver sourceAPI.Resolver) ([]mcpAPI.MCPServer, error) {
	var servers []mcpAPI.MCPServer

	for _, mcpSourceConfig := range config.Sources {
		mcpUri := mcpSourceConfig.URI
		mcpSource, err := sourceResolver.Resolve(mcpUri)
		if err != nil {
			return nil, fmt.Errorf("resolve: %s: %w", mcpUri, err)
		}

		mcpLoader := mcpAPI.NewLoader(mcpSource)
		serversSet, err := mcpLoader.Load()
		if err != nil {
			return nil, fmt.Errorf("load mcp servers: %w", err)
		}

		servers = append(servers, serversSet...)
	}

	return servers, nil
}
