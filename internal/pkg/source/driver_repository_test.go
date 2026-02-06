package source

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sourceAPI "github.com/orbiqd/orbiqd-projectkit/pkg/source"
)

func TestDriverRepository_RegisterDriver(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setup          func() (*DriverRepository, *sourceAPI.MockDriver)
		driverSchemes  []string
		expectedError  error
		verifyFn       func(t *testing.T, repo *DriverRepository)
	}{
		{
			name: "WhenSingleScheme_ThenRegisters",
			setup: func() (*DriverRepository, *sourceAPI.MockDriver) {
				repo := NewDriverRepository()
				driver := sourceAPI.NewMockDriver(t)
				return repo, driver
			},
			driverSchemes: []string{"file"},
			expectedError: nil,
			verifyFn: func(t *testing.T, repo *DriverRepository) {
				driver, err := repo.GetDriverByScheme("file")
				require.NoError(t, err)
				assert.NotNil(t, driver)
			},
		},
		{
			name: "WhenMultipleSchemes_ThenRegistersAll",
			setup: func() (*DriverRepository, *sourceAPI.MockDriver) {
				repo := NewDriverRepository()
				driver := sourceAPI.NewMockDriver(t)
				return repo, driver
			},
			driverSchemes: []string{"http", "https"},
			expectedError: nil,
			verifyFn: func(t *testing.T, repo *DriverRepository) {
				httpDriver, err := repo.GetDriverByScheme("http")
				require.NoError(t, err)
				assert.NotNil(t, httpDriver)

				httpsDriver, err := repo.GetDriverByScheme("https")
				require.NoError(t, err)
				assert.NotNil(t, httpsDriver)

				assert.Same(t, httpDriver, httpsDriver)
			},
		},
		{
			name: "WhenSchemeAlreadyRegistered_ThenReturnsError",
			setup: func() (*DriverRepository, *sourceAPI.MockDriver) {
				repo := NewDriverRepository()
				existingDriver := sourceAPI.NewMockDriver(t)
				existingDriver.EXPECT().GetSupportedSchemes().Return([]string{"file"})
				err := repo.RegisterDriver(existingDriver)
				require.NoError(t, err)

				newDriver := sourceAPI.NewMockDriver(t)
				return repo, newDriver
			},
			driverSchemes: []string{"file"},
			expectedError: sourceAPI.ErrSchemeDriverAlreadyRegistered,
		},
		{
			name: "WhenPartialOverlap_ThenReturnsErrorAndNoPartialRegistration",
			setup: func() (*DriverRepository, *sourceAPI.MockDriver) {
				repo := NewDriverRepository()
				existingDriver := sourceAPI.NewMockDriver(t)
				existingDriver.EXPECT().GetSupportedSchemes().Return([]string{"http"})
				err := repo.RegisterDriver(existingDriver)
				require.NoError(t, err)

				newDriver := sourceAPI.NewMockDriver(t)
				return repo, newDriver
			},
			driverSchemes: []string{"https", "http"},
			expectedError: sourceAPI.ErrSchemeDriverAlreadyRegistered,
			verifyFn: func(t *testing.T, repo *DriverRepository) {
				_, err := repo.GetDriverByScheme("https")
				assert.ErrorIs(t, err, sourceAPI.ErrSchemeDriverNotRegistered)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo, driver := tt.setup()
			driver.EXPECT().GetSupportedSchemes().Return(tt.driverSchemes)

			err := repo.RegisterDriver(driver)

			if tt.expectedError != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedError))
			} else {
				require.NoError(t, err)
			}

			if tt.verifyFn != nil {
				tt.verifyFn(t, repo)
			}
		})
	}
}

func TestDriverRepository_GetDriverByScheme(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setup         func() *DriverRepository
		scheme        string
		expectedError error
		expectDriver  bool
	}{
		{
			name: "WhenSchemeRegistered_ThenReturnsDriver",
			setup: func() *DriverRepository {
				repo := NewDriverRepository()
				driver := sourceAPI.NewMockDriver(t)
				driver.EXPECT().GetSupportedSchemes().Return([]string{"file"})
				err := repo.RegisterDriver(driver)
				require.NoError(t, err)
				return repo
			},
			scheme:        "file",
			expectedError: nil,
			expectDriver:  true,
		},
		{
			name: "WhenSchemeNotRegistered_ThenReturnsError",
			setup: func() *DriverRepository {
				return NewDriverRepository()
			},
			scheme:        "unknown",
			expectedError: sourceAPI.ErrSchemeDriverNotRegistered,
			expectDriver:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := tt.setup()

			driver, err := repo.GetDriverByScheme(tt.scheme)

			if tt.expectedError != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedError))
				assert.Nil(t, driver)
			} else {
				require.NoError(t, err)
				if tt.expectDriver {
					assert.NotNil(t, driver)
				}
			}
		})
	}
}
