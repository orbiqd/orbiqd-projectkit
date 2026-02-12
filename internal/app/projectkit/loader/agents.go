package loader

import (
	"fmt"
	"log/slog"

	agentAPI "github.com/orbiqd/orbiqd-projectkit/pkg/agent"
	projectAPI "github.com/orbiqd/orbiqd-projectkit/pkg/project"
)

func LoadAgentsFromConfig(config projectAPI.Config, agentRegistry agentAPI.Registry) ([]agentAPI.Agent, error) {
	var agents []agentAPI.Agent

	for _, agentConfig := range config.Agents {
		agentKind := agentConfig.Kind

		provider, err := agentRegistry.GetByKind(agentKind)
		if err != nil {
			return []agentAPI.Agent{}, fmt.Errorf("get provider: %s: %w", agentKind, err)
		}

		agent, err := provider.NewAgent(agentConfig.Options)
		if err != nil {
			return []agentAPI.Agent{}, fmt.Errorf("create agent: %s: %w", agentKind, err)
		}

		agents = append(agents, agent)

		slog.Info("Found agent in config.", slog.String("agentKind", string(agent.GetKind())))
	}

	return agents, nil
}
