package project

import (
	"errors"
	"testing"

	projectAPI "github.com/orbiqd/orbiqd-projectkit/pkg/project"
	rulebookAPI "github.com/orbiqd/orbiqd-projectkit/pkg/rulebook"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfigProvider_WhenLoadSucceeds_ThenReturnsConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		mockConfig *projectAPI.Config
	}{
		{
			name:       "WhenConfigIsNonNil_ThenReturnsConfig",
			mockConfig: &projectAPI.Config{},
		},
		{
			name: "WhenConfigHasRulebook_ThenReturnsConfig",
			mockConfig: &projectAPI.Config{
				Rulebook: &rulebookAPI.Config{
					Sources: []rulebookAPI.SourceConfig{{URI: "file://test"}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			mockLoader := projectAPI.NewMockConfigLoader(t)
			mockLoader.EXPECT().Load().Return(tt.mockConfig, nil)

			provider := NewConfigProvider()

			// Act
			config, err := provider(mockLoader)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tt.mockConfig, config)
		})
	}
}

func TestNewConfigProvider_WhenLoadFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		mockErr       error
		expectedError error
	}{
		{
			name:          "WhenLoadReturnsNoConfigError_ThenReturnsError",
			mockErr:       projectAPI.ErrNoConfigResolved,
			expectedError: projectAPI.ErrNoConfigResolved,
		},
		{
			name:          "WhenLoadReturnsLoadFailedError_ThenReturnsError",
			mockErr:       projectAPI.ErrConfigLoadFailed,
			expectedError: projectAPI.ErrConfigLoadFailed,
		},
		{
			name:          "WhenLoadReturnsValidationError_ThenReturnsError",
			mockErr:       projectAPI.ErrConfigValidationFailed,
			expectedError: projectAPI.ErrConfigValidationFailed,
		},
		{
			name:          "WhenLoadReturnsGenericError_ThenReturnsError",
			mockErr:       errors.New("generic error"),
			expectedError: errors.New("generic error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			mockLoader := projectAPI.NewMockConfigLoader(t)
			mockLoader.EXPECT().Load().Return(nil, tt.mockErr)

			provider := NewConfigProvider()

			// Act
			config, err := provider(mockLoader)

			// Assert
			require.Error(t, err)
			assert.Nil(t, config)
			assert.Equal(t, tt.expectedError.Error(), err.Error())
		})
	}
}
