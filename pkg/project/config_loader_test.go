package project

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockConfigLoader struct {
	id int
}

func (m *mockConfigLoader) Load() (*Config, error) {
	return &Config{}, nil
}

func resetDefaultLoaderFactory() {
	defaultLoaderFactory = nil
}

func TestDefaultConfigLoader_WhenNoLoaderRegistered_ThenReturnsError(t *testing.T) {
	// Arrange
	resetDefaultLoaderFactory()
	defer resetDefaultLoaderFactory()

	// Act
	loader, err := DefaultConfigLoader()

	// Assert
	require.ErrorIs(t, err, ErrNoDefaultConfigLoaderRegistered)
	assert.Nil(t, loader)
}

func TestDefaultConfigLoader_WhenLoaderRegistered_ThenReturnsLoader(t *testing.T) {
	// Arrange
	resetDefaultLoaderFactory()
	defer resetDefaultLoaderFactory()

	expectedLoader := &mockConfigLoader{}
	RegisterDefaultLoader(func() ConfigLoader {
		return expectedLoader
	})

	// Act
	loader, err := DefaultConfigLoader()

	// Assert
	require.NoError(t, err)
	assert.Same(t, expectedLoader, loader)
}

func TestRegisterDefaultLoader_WhenCalled_ThenOverridesPreviousFactory(t *testing.T) {
	// Arrange
	resetDefaultLoaderFactory()
	defer resetDefaultLoaderFactory()

	firstLoader := &mockConfigLoader{id: 1}
	secondLoader := &mockConfigLoader{id: 2}

	RegisterDefaultLoader(func() ConfigLoader {
		return firstLoader
	})

	// Act
	RegisterDefaultLoader(func() ConfigLoader {
		return secondLoader
	})

	// Assert
	loader, err := DefaultConfigLoader()
	require.NoError(t, err)
	assert.Same(t, secondLoader, loader)
	assert.NotSame(t, firstLoader, loader)
}
