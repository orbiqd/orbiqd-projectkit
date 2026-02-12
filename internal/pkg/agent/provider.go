package agent

import (
	"fmt"

	agentAPI "github.com/orbiqd/orbiqd-projectkit/pkg/agent"
	projectAPI "github.com/orbiqd/orbiqd-projectkit/pkg/project"
	"github.com/spf13/afero"
)

type ProviderFactory func(rootFs afero.Fs) agentAPI.Provider

func NewRegistryProvider(factories ...ProviderFactory) func(projectAPI.Fs) (agentAPI.Registry, error) {
	return func(projectFs projectAPI.Fs) (agentAPI.Registry, error) {
		registry := NewStaticRegistry()
		for _, factory := range factories {
			provider := factory(projectFs)
			if err := registry.Register(provider); err != nil {
				return nil, fmt.Errorf("register %s provider: %w", provider.GetKind(), err)
			}
		}
		return registry, nil
	}
}
