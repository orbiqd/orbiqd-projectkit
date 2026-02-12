package loader

import (
	"errors"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mcpAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/mcp"
	sourceAPI "github.com/orbiqd/orbiqd-projectkit/pkg/source"
)

func validMCPServerFs(t *testing.T, serverName string, executablePath string) afero.Fs {
	t.Helper()

	fs := afero.NewMemMapFs()

	content := "name: \"" + serverName + "\"\nstdio:\n  executablePath: \"" + executablePath + "\"\n"

	require.NoError(t, afero.WriteFile(fs, serverName+".yaml", []byte(content), 0644))

	return fs
}

func TestLoadAiMCPServersFromConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		sources        []mcpAPI.SourceConfig
		mockSetup      func(*sourceAPI.MockResolver)
		wantServersLen int
	}{
		{
			name:           "WhenNoSources_ThenReturnsNilServers",
			sources:        []mcpAPI.SourceConfig{},
			mockSetup:      func(m *sourceAPI.MockResolver) {},
			wantServersLen: 0,
		},
		{
			name: "WhenSingleSourceWithOneServer_ThenReturnsServer",
			sources: []mcpAPI.SourceConfig{
				{URI: "file://./mcp-servers"},
			},
			mockSetup: func(m *sourceAPI.MockResolver) {
				fs := validMCPServerFs(t, "test-server", "/usr/local/bin/test")
				m.EXPECT().Resolve("file://./mcp-servers").Return(fs, nil)
			},
			wantServersLen: 1,
		},
		{
			name: "WhenMultipleSources_ThenReturnsCombinedServers",
			sources: []mcpAPI.SourceConfig{
				{URI: "file://./mcp-servers1"},
				{URI: "file://./mcp-servers2"},
			},
			mockSetup: func(m *sourceAPI.MockResolver) {
				fs1 := validMCPServerFs(t, "server-one", "/usr/local/bin/one")
				fs2 := validMCPServerFs(t, "server-two", "/usr/local/bin/two")
				m.EXPECT().Resolve("file://./mcp-servers1").Return(fs1, nil)
				m.EXPECT().Resolve("file://./mcp-servers2").Return(fs2, nil)
			},
			wantServersLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockResolver := sourceAPI.NewMockResolver(t)
			tt.mockSetup(mockResolver)

			config := mcpAPI.Config{
				Sources: tt.sources,
			}

			servers, err := LoadAiMCPServersFromConfig(config, mockResolver)

			require.NoError(t, err)
			assert.Len(t, servers, tt.wantServersLen)
		})
	}
}

func TestLoadAiMCPServersFromConfig_WhenResolverFails_ThenReturnsResolveError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	resolveErr := errors.New("resolver failed")
	mockResolver.EXPECT().Resolve("file://./mcp-servers").Return(nil, resolveErr)

	config := mcpAPI.Config{
		Sources: []mcpAPI.SourceConfig{
			{URI: "file://./mcp-servers"},
		},
	}

	servers, err := LoadAiMCPServersFromConfig(config, mockResolver)

	require.Error(t, err)
	assert.Nil(t, servers)
	assert.Contains(t, err.Error(), "resolve: file://./mcp-servers")
	assert.ErrorIs(t, err, resolveErr)
}

func TestLoadAiMCPServersFromConfig_WhenLoaderFails_ThenReturnsLoadMCPServersError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	emptyFs := afero.NewMemMapFs()
	mockResolver.EXPECT().Resolve("file://./mcp-servers").Return(emptyFs, nil)

	config := mcpAPI.Config{
		Sources: []mcpAPI.SourceConfig{
			{URI: "file://./mcp-servers"},
		},
	}

	servers, err := LoadAiMCPServersFromConfig(config, mockResolver)

	require.Error(t, err)
	assert.Nil(t, servers)
	assert.Contains(t, err.Error(), "load mcp servers:")
	assert.ErrorIs(t, err, mcpAPI.ErrNoMCPServersFound)
}

func TestLoadAiMCPServersFromConfig_WhenSecondSourceResolverFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	fs1 := validMCPServerFs(t, "server-one", "/usr/local/bin/one")
	resolveErr := errors.New("second resolver failed")

	mockResolver.EXPECT().Resolve("file://./mcp-servers1").Return(fs1, nil)
	mockResolver.EXPECT().Resolve("file://./mcp-servers2").Return(nil, resolveErr)

	config := mcpAPI.Config{
		Sources: []mcpAPI.SourceConfig{
			{URI: "file://./mcp-servers1"},
			{URI: "file://./mcp-servers2"},
		},
	}

	servers, err := LoadAiMCPServersFromConfig(config, mockResolver)

	require.Error(t, err)
	assert.Nil(t, servers)
	assert.Contains(t, err.Error(), "resolve: file://./mcp-servers2")
	assert.ErrorIs(t, err, resolveErr)
}

func TestLoadAiMCPServersFromConfig_WhenSecondSourceLoaderFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	fs1 := validMCPServerFs(t, "server-one", "/usr/local/bin/one")
	emptyFs := afero.NewMemMapFs()

	mockResolver.EXPECT().Resolve("file://./mcp-servers1").Return(fs1, nil)
	mockResolver.EXPECT().Resolve("file://./mcp-servers2").Return(emptyFs, nil)

	config := mcpAPI.Config{
		Sources: []mcpAPI.SourceConfig{
			{URI: "file://./mcp-servers1"},
			{URI: "file://./mcp-servers2"},
		},
	}

	servers, err := LoadAiMCPServersFromConfig(config, mockResolver)

	require.Error(t, err)
	assert.Nil(t, servers)
	assert.Contains(t, err.Error(), "load mcp servers:")
	assert.ErrorIs(t, err, mcpAPI.ErrNoMCPServersFound)
}
