package loader

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	agentAPI "github.com/orbiqd/orbiqd-projectkit/pkg/agent"
	projectAPI "github.com/orbiqd/orbiqd-projectkit/pkg/project"
)

func TestLoadAgentsFromConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		agents        []agentAPI.Config
		mockSetup     func(*agentAPI.MockRegistry)
		wantAgentsLen int
	}{
		{
			name:          "WhenNoAgents_ThenReturnsNilAgents",
			agents:        []agentAPI.Config{},
			mockSetup:     func(m *agentAPI.MockRegistry) {},
			wantAgentsLen: 0,
		},
		{
			name: "WhenSingleAgent_ThenReturnsOneAgent",
			agents: []agentAPI.Config{
				{Kind: "test-agent"},
			},
			mockSetup: func(m *agentAPI.MockRegistry) {
				mockProvider := agentAPI.NewMockProvider(t)
				mockAgent := agentAPI.NewMockAgent(t)

				m.EXPECT().GetByKind(agentAPI.Kind("test-agent")).Return(mockProvider, nil)
				mockProvider.EXPECT().NewAgent(nil).Return(mockAgent, nil)
				mockAgent.EXPECT().GetKind().Return(agentAPI.Kind("test-agent"))
			},
			wantAgentsLen: 1,
		},
		{
			name: "WhenMultipleAgents_ThenReturnsAllAgents",
			agents: []agentAPI.Config{
				{Kind: "agent-one"},
				{Kind: "agent-two"},
			},
			mockSetup: func(m *agentAPI.MockRegistry) {
				mockProvider1 := agentAPI.NewMockProvider(t)
				mockAgent1 := agentAPI.NewMockAgent(t)
				m.EXPECT().GetByKind(agentAPI.Kind("agent-one")).Return(mockProvider1, nil)
				mockProvider1.EXPECT().NewAgent(nil).Return(mockAgent1, nil)
				mockAgent1.EXPECT().GetKind().Return(agentAPI.Kind("agent-one"))

				mockProvider2 := agentAPI.NewMockProvider(t)
				mockAgent2 := agentAPI.NewMockAgent(t)
				m.EXPECT().GetByKind(agentAPI.Kind("agent-two")).Return(mockProvider2, nil)
				mockProvider2.EXPECT().NewAgent(nil).Return(mockAgent2, nil)
				mockAgent2.EXPECT().GetKind().Return(agentAPI.Kind("agent-two"))
			},
			wantAgentsLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRegistry := agentAPI.NewMockRegistry(t)
			tt.mockSetup(mockRegistry)

			config := projectAPI.Config{
				Agents: tt.agents,
			}

			agents, err := LoadAgentsFromConfig(config, mockRegistry)

			require.NoError(t, err)
			assert.Len(t, agents, tt.wantAgentsLen)
		})
	}
}

func TestLoadAgentsFromConfig_WhenRegistryGetByKindFails_ThenReturnsGetProviderError(t *testing.T) {
	t.Parallel()

	mockRegistry := agentAPI.NewMockRegistry(t)
	getByKindErr := errors.New("registry get by kind failed")
	mockRegistry.EXPECT().GetByKind(agentAPI.Kind("test-agent")).Return(nil, getByKindErr)

	config := projectAPI.Config{
		Agents: []agentAPI.Config{
			{Kind: "test-agent"},
		},
	}

	agents, err := LoadAgentsFromConfig(config, mockRegistry)

	require.Error(t, err)
	assert.Empty(t, agents)
	assert.Contains(t, err.Error(), "get provider: test-agent")
	assert.ErrorIs(t, err, getByKindErr)
}

func TestLoadAgentsFromConfig_WhenProviderNewAgentFails_ThenReturnsCreateAgentError(t *testing.T) {
	t.Parallel()

	mockRegistry := agentAPI.NewMockRegistry(t)
	mockProvider := agentAPI.NewMockProvider(t)
	newAgentErr := errors.New("provider new agent failed")

	mockRegistry.EXPECT().GetByKind(agentAPI.Kind("test-agent")).Return(mockProvider, nil)
	mockProvider.EXPECT().NewAgent(nil).Return(nil, newAgentErr)

	config := projectAPI.Config{
		Agents: []agentAPI.Config{
			{Kind: "test-agent"},
		},
	}

	agents, err := LoadAgentsFromConfig(config, mockRegistry)

	require.Error(t, err)
	assert.Empty(t, agents)
	assert.Contains(t, err.Error(), "create agent: test-agent")
	assert.ErrorIs(t, err, newAgentErr)
}

func TestLoadAgentsFromConfig_WhenSecondAgentGetByKindFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockRegistry := agentAPI.NewMockRegistry(t)
	mockProvider1 := agentAPI.NewMockProvider(t)
	mockAgent1 := agentAPI.NewMockAgent(t)
	getByKindErr := errors.New("second registry get by kind failed")

	mockRegistry.EXPECT().GetByKind(agentAPI.Kind("agent-one")).Return(mockProvider1, nil)
	mockProvider1.EXPECT().NewAgent(nil).Return(mockAgent1, nil)
	mockAgent1.EXPECT().GetKind().Return(agentAPI.Kind("agent-one"))

	mockRegistry.EXPECT().GetByKind(agentAPI.Kind("failing-agent")).Return(nil, getByKindErr)

	config := projectAPI.Config{
		Agents: []agentAPI.Config{
			{Kind: "agent-one"},
			{Kind: "failing-agent"},
		},
	}

	agents, err := LoadAgentsFromConfig(config, mockRegistry)

	require.Error(t, err)
	assert.Empty(t, agents)
	assert.Contains(t, err.Error(), "get provider: failing-agent")
	assert.ErrorIs(t, err, getByKindErr)
}

func TestLoadAgentsFromConfig_WhenSecondAgentNewAgentFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockRegistry := agentAPI.NewMockRegistry(t)
	mockProvider1 := agentAPI.NewMockProvider(t)
	mockAgent1 := agentAPI.NewMockAgent(t)
	mockProvider2 := agentAPI.NewMockProvider(t)
	newAgentErr := errors.New("second provider new agent failed")

	mockRegistry.EXPECT().GetByKind(agentAPI.Kind("agent-one")).Return(mockProvider1, nil)
	mockProvider1.EXPECT().NewAgent(nil).Return(mockAgent1, nil)
	mockAgent1.EXPECT().GetKind().Return(agentAPI.Kind("agent-one"))

	mockRegistry.EXPECT().GetByKind(agentAPI.Kind("failing-agent")).Return(mockProvider2, nil)
	mockProvider2.EXPECT().NewAgent(nil).Return(nil, newAgentErr)

	config := projectAPI.Config{
		Agents: []agentAPI.Config{
			{Kind: "agent-one"},
			{Kind: "failing-agent"},
		},
	}

	agents, err := LoadAgentsFromConfig(config, mockRegistry)

	require.Error(t, err)
	assert.Empty(t, agents)
	assert.Contains(t, err.Error(), "create agent: failing-agent")
	assert.ErrorIs(t, err, newAgentErr)
}
