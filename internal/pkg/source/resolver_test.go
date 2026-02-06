package source

import (
	"errors"
	"testing"

	sourceAPI "github.com/orbiqd/orbiqd-projectkit/pkg/source"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolver_Resolve(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		uri       string
		setupMock func(*sourceAPI.MockDriverRepository, *sourceAPI.MockDriver)
		assertErr func(*testing.T, error)
		assertFs  func(*testing.T, afero.Fs)
	}{
		{
			name: "WhenValidURI_ThenReturnsFs",
			uri:  "file:///path/to/dir",
			setupMock: func(repo *sourceAPI.MockDriverRepository, driver *sourceAPI.MockDriver) {
				expectedFs := afero.NewMemMapFs()
				repo.EXPECT().GetDriverByScheme("file").Return(driver, nil)
				driver.EXPECT().Resolve("file:///path/to/dir").Return(expectedFs, nil)
			},
			assertErr: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
			assertFs: func(t *testing.T, fs afero.Fs) {
				assert.NotNil(t, fs)
			},
		},
		{
			name: "WhenURIWithoutScheme_ThenReturnsError",
			uri:  "no-scheme-here",
			setupMock: func(repo *sourceAPI.MockDriverRepository, driver *sourceAPI.MockDriver) {
			},
			assertErr: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "uri scheme not found")
			},
			assertFs: func(t *testing.T, fs afero.Fs) {
				assert.Nil(t, fs)
			},
		},
		{
			name: "WhenEmptyURI_ThenReturnsError",
			uri:  "",
			setupMock: func(repo *sourceAPI.MockDriverRepository, driver *sourceAPI.MockDriver) {
			},
			assertErr: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "uri scheme not found")
			},
			assertFs: func(t *testing.T, fs afero.Fs) {
				assert.Nil(t, fs)
			},
		},
		{
			name: "WhenDriverRepositoryReturnsError_ThenReturnsWrappedError",
			uri:  "unknown://res",
			setupMock: func(repo *sourceAPI.MockDriverRepository, driver *sourceAPI.MockDriver) {
				repo.EXPECT().GetDriverByScheme("unknown").Return(nil, sourceAPI.ErrSchemeDriverNotRegistered)
			},
			assertErr: func(t *testing.T, err error) {
				require.Error(t, err)
				require.ErrorIs(t, err, sourceAPI.ErrSchemeDriverNotRegistered)
				require.ErrorContains(t, err, "get driver by scheme")
			},
			assertFs: func(t *testing.T, fs afero.Fs) {
				assert.Nil(t, fs)
			},
		},
		{
			name: "WhenDriverResolveReturnsError_ThenReturnsWrappedError",
			uri:  "file:///bad",
			setupMock: func(repo *sourceAPI.MockDriverRepository, driver *sourceAPI.MockDriver) {
				driverErr := errors.New("driver error")
				repo.EXPECT().GetDriverByScheme("file").Return(driver, nil)
				driver.EXPECT().Resolve("file:///bad").Return(nil, driverErr)
			},
			assertErr: func(t *testing.T, err error) {
				require.Error(t, err)
				require.ErrorContains(t, err, "resolve file:///bad")
				require.ErrorContains(t, err, "driver error")
			},
			assertFs: func(t *testing.T, fs afero.Fs) {
				assert.Nil(t, fs)
			},
		},
		{
			name: "WhenURIHasMultipleSeparators_ThenUsesFirstScheme",
			uri:  "http://host://extra",
			setupMock: func(repo *sourceAPI.MockDriverRepository, driver *sourceAPI.MockDriver) {
				expectedFs := afero.NewMemMapFs()
				repo.EXPECT().GetDriverByScheme("http").Return(driver, nil)
				driver.EXPECT().Resolve("http://host://extra").Return(expectedFs, nil)
			},
			assertErr: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
			assertFs: func(t *testing.T, fs afero.Fs) {
				assert.NotNil(t, fs)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := sourceAPI.NewMockDriverRepository(t)
			mockDriver := sourceAPI.NewMockDriver(t)

			tt.setupMock(mockRepo, mockDriver)

			resolver := NewResolver(mockRepo)

			fs, err := resolver.Resolve(tt.uri)

			tt.assertErr(t, err)
			tt.assertFs(t, fs)
		})
	}
}
