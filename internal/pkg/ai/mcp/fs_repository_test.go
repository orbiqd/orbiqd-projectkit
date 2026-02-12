package mcp

import (
	"encoding/json"
	"errors"
	"io/fs"
	"testing"

	mcpAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/mcp"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.nhat.io/aferomock"
)

func TestFsRepository_GetAll_WhenEmpty_ThenReturnsEmptySlice(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)

	result, err := repo.GetAll()

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestFsRepository_GetAll_WhenSingleServer_ThenReturnsIt(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)
	server := mcpAPI.MCPServer{
		Name: "test-server",
		STDIO: &mcpAPI.STDIOMCPServer{
			ExecutablePath:       "/usr/local/bin/test",
			Arguments:            []string{"--port", "8080"},
			EnvironmentVariables: map[string]string{"API_KEY": "secret"},
		},
	}

	err := repo.AddMCPServer(server)
	require.NoError(t, err)

	result, err := repo.GetAll()

	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "test-server", result[0].Name)
}

func TestFsRepository_GetAll_WhenMultipleServers_ThenReturnsSortedByName(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)
	server1 := mcpAPI.MCPServer{
		Name: "zebra-server",
		STDIO: &mcpAPI.STDIOMCPServer{
			ExecutablePath: "/usr/local/bin/zebra",
		},
	}
	server2 := mcpAPI.MCPServer{
		Name: "alpha-server",
		STDIO: &mcpAPI.STDIOMCPServer{
			ExecutablePath: "/usr/local/bin/alpha",
		},
	}

	err := repo.AddMCPServer(server1)
	require.NoError(t, err)
	err = repo.AddMCPServer(server2)
	require.NoError(t, err)

	result, err := repo.GetAll()

	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, "alpha-server", result[0].Name)
	assert.Equal(t, "zebra-server", result[1].Name)
}

func TestFsRepository_GetAll_WhenReadDirFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockedErr := errors.New("read dir error")
	mockFs := aferomock.OverrideFs(afero.NewMemMapFs(), aferomock.FsCallbacks{
		OpenFunc: func(name string) (afero.File, error) {
			return nil, mockedErr
		},
	})
	repo := NewFsRepository(mockFs)

	result, err := repo.GetAll()

	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, mockedErr)
}

func TestFsRepository_GetAll_WhenUnmarshalFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "invalid.json", []byte("invalid json"), 0644)
	repo := NewFsRepository(fs)

	result, err := repo.GetAll()

	require.Error(t, err)
	assert.Nil(t, result)
}

func TestFsRepository_AddMCPServer_WhenValid_ThenPersistsToFs(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)
	server := mcpAPI.MCPServer{
		Name: "test-server",
		STDIO: &mcpAPI.STDIOMCPServer{
			ExecutablePath: "/usr/local/bin/test",
		},
	}

	err := repo.AddMCPServer(server)
	require.NoError(t, err)

	files, err := afero.ReadDir(fs, ".")
	require.NoError(t, err)
	require.Len(t, files, 1)
	assert.Contains(t, files[0].Name(), ".json")

	data, err := afero.ReadFile(fs, files[0].Name())
	require.NoError(t, err)

	var persisted mcpAPI.MCPServer
	err = json.Unmarshal(data, &persisted)
	require.NoError(t, err)
	assert.Equal(t, "test-server", persisted.Name)
}

func TestFsRepository_AddMCPServer_WhenMarshalFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockFs := aferomock.OverrideFs(afero.NewMemMapFs(), aferomock.FsCallbacks{
		OpenFileFunc: func(name string, flag int, perm fs.FileMode) (afero.File, error) {
			return nil, errors.New("write error")
		},
	})
	repo := NewFsRepository(mockFs)
	server := mcpAPI.MCPServer{
		Name: "test-server",
		STDIO: &mcpAPI.STDIOMCPServer{
			ExecutablePath: "/usr/local/bin/test",
		},
	}

	err := repo.AddMCPServer(server)

	require.Error(t, err)
}

func TestFsRepository_RemoveAll_WhenEmpty_ThenNoError(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)

	err := repo.RemoveAll()

	require.NoError(t, err)
}

func TestFsRepository_RemoveAll_WhenMultipleFiles_ThenRemovesAll(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)
	server1 := mcpAPI.MCPServer{
		Name: "server1",
		STDIO: &mcpAPI.STDIOMCPServer{
			ExecutablePath: "/usr/local/bin/server1",
		},
	}
	server2 := mcpAPI.MCPServer{
		Name: "server2",
		STDIO: &mcpAPI.STDIOMCPServer{
			ExecutablePath: "/usr/local/bin/server2",
		},
	}

	_ = repo.AddMCPServer(server1)
	_ = repo.AddMCPServer(server2)

	err := repo.RemoveAll()
	require.NoError(t, err)

	result, err := repo.GetAll()
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestFsRepository_RemoveAll_ThenAdd_ThenGetAll(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs)
	server := mcpAPI.MCPServer{
		Name: "test-server",
		STDIO: &mcpAPI.STDIOMCPServer{
			ExecutablePath: "/usr/local/bin/test",
		},
	}

	_ = repo.AddMCPServer(server)
	err := repo.RemoveAll()
	require.NoError(t, err)

	err = repo.AddMCPServer(server)
	require.NoError(t, err)

	result, err := repo.GetAll()
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "test-server", result[0].Name)
}
