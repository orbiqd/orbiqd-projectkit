package agent

import (
	"testing"

	agentAPI "github.com/orbiqd/orbiqd-projectkit/pkg/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStaticRegistry_Register(t *testing.T) {
	t.Parallel()

	t.Run("WhenNewKind_ThenRegistersSuccessfully", func(t *testing.T) {
		t.Parallel()

		registry := NewStaticRegistry()
		mockProvider := agentAPI.NewMockProvider(t)
		mockProvider.EXPECT().GetKind().Return(agentAPI.Kind("test-agent"))

		err := registry.Register(mockProvider)

		require.NoError(t, err)

		provider, err := registry.GetByKind(agentAPI.Kind("test-agent"))
		require.NoError(t, err)
		assert.Equal(t, mockProvider, provider)
	})

	t.Run("WhenDuplicateKind_ThenReturnsAlreadyRegisteredError", func(t *testing.T) {
		t.Parallel()

		registry := NewStaticRegistry()
		mockProvider1 := agentAPI.NewMockProvider(t)
		mockProvider2 := agentAPI.NewMockProvider(t)
		mockProvider1.EXPECT().GetKind().Return(agentAPI.Kind("test-agent"))
		mockProvider2.EXPECT().GetKind().Return(agentAPI.Kind("test-agent"))

		err := registry.Register(mockProvider1)
		require.NoError(t, err)

		err = registry.Register(mockProvider2)

		require.Error(t, err)
		require.ErrorIs(t, err, agentAPI.ErrProviderAlreadyRegistered)
	})
}

func TestStaticRegistry_GetByKind(t *testing.T) {
	t.Parallel()

	t.Run("WhenKindExists_ThenReturnsProvider", func(t *testing.T) {
		t.Parallel()

		registry := NewStaticRegistry()
		mockProvider := agentAPI.NewMockProvider(t)
		mockProvider.EXPECT().GetKind().Return(agentAPI.Kind("test-agent"))

		err := registry.Register(mockProvider)
		require.NoError(t, err)

		provider, err := registry.GetByKind(agentAPI.Kind("test-agent"))

		require.NoError(t, err)
		assert.Equal(t, mockProvider, provider)
	})

	t.Run("WhenKindNotExists_ThenReturnsNotRegisteredError", func(t *testing.T) {
		t.Parallel()

		registry := NewStaticRegistry()

		provider, err := registry.GetByKind(agentAPI.Kind("non-existent"))

		require.Error(t, err)
		require.ErrorIs(t, err, agentAPI.ErrProviderNotRegistered)
		assert.Nil(t, provider)
	})
}

func TestStaticRegistry_GetAllKinds(t *testing.T) {
	t.Parallel()

	t.Run("WhenEmpty_ThenReturnsEmptySlice", func(t *testing.T) {
		t.Parallel()

		registry := NewStaticRegistry()

		kinds := registry.GetAllKinds()

		assert.NotNil(t, kinds)
		assert.Empty(t, kinds)
	})

	t.Run("WhenMultipleRegistered_ThenReturnsAllKinds", func(t *testing.T) {
		t.Parallel()

		registry := NewStaticRegistry()
		mockProvider1 := agentAPI.NewMockProvider(t)
		mockProvider2 := agentAPI.NewMockProvider(t)
		mockProvider3 := agentAPI.NewMockProvider(t)
		mockProvider1.EXPECT().GetKind().Return(agentAPI.Kind("agent-one"))
		mockProvider2.EXPECT().GetKind().Return(agentAPI.Kind("agent-two"))
		mockProvider3.EXPECT().GetKind().Return(agentAPI.Kind("agent-three"))

		err := registry.Register(mockProvider1)
		require.NoError(t, err)
		err = registry.Register(mockProvider2)
		require.NoError(t, err)
		err = registry.Register(mockProvider3)
		require.NoError(t, err)

		kinds := registry.GetAllKinds()

		assert.Len(t, kinds, 3)
		assert.ElementsMatch(t, []agentAPI.Kind{
			agentAPI.Kind("agent-one"),
			agentAPI.Kind("agent-two"),
			agentAPI.Kind("agent-three"),
		}, kinds)
	})
}
