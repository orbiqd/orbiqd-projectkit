package agent

import (
	"sync"

	agentAPI "github.com/orbiqd/orbiqd-projectkit/pkg/agent"
)

type StaticRegistry struct {
	mutex     sync.RWMutex
	providers map[agentAPI.Kind]agentAPI.Provider
}

var _ agentAPI.Registry = (*StaticRegistry)(nil)

func NewStaticRegistry() *StaticRegistry {
	return &StaticRegistry{
		mutex:     sync.RWMutex{},
		providers: make(map[agentAPI.Kind]agentAPI.Provider),
	}
}

func (registry *StaticRegistry) GetAllKinds() []agentAPI.Kind {
	registry.mutex.RLock()
	defer registry.mutex.RUnlock()

	kinds := make([]agentAPI.Kind, 0, len(registry.providers))
	for kind := range registry.providers {
		kinds = append(kinds, kind)
	}

	return kinds
}

func (registry *StaticRegistry) Register(provider agentAPI.Provider) error {
	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	kind := provider.GetKind()
	if _, exists := registry.providers[kind]; exists {
		return agentAPI.ErrProviderAlreadyRegistered
	}

	registry.providers[kind] = provider

	return nil
}

func (registry *StaticRegistry) GetByKind(kind agentAPI.Kind) (agentAPI.Provider, error) {
	registry.mutex.RLock()
	defer registry.mutex.RUnlock()

	provider, exists := registry.providers[kind]
	if !exists {
		return nil, agentAPI.ErrProviderNotRegistered
	}

	return provider, nil
}
