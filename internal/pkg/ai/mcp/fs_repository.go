package mcp

import (
	"encoding/json"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/google/uuid"
	mcpAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/mcp"
	"github.com/spf13/afero"
)

type FsRepository struct {
	mutex sync.RWMutex
	fs    afero.Fs
}

var _ mcpAPI.Repository = (*FsRepository)(nil)

func NewFsRepository(fs afero.Fs) *FsRepository {
	return &FsRepository{
		mutex: sync.RWMutex{},
		fs:    fs,
	}
}

func (repository *FsRepository) listFiles() ([]string, error) {
	entries, err := afero.ReadDir(repository.fs, ".")
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.ToLower(filepath.Ext(entry.Name())) == ".json" {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

func (repository *FsRepository) loadFile(filename string) (mcpAPI.MCPServer, error) {
	data, err := afero.ReadFile(repository.fs, filename)
	if err != nil {
		return mcpAPI.MCPServer{}, err
	}

	var server mcpAPI.MCPServer
	if err := json.Unmarshal(data, &server); err != nil {
		return mcpAPI.MCPServer{}, err
	}

	return server, nil
}

func (repository *FsRepository) saveFile(filename string, server mcpAPI.MCPServer) error {
	data, err := json.Marshal(server)
	if err != nil {
		return err
	}

	return afero.WriteFile(repository.fs, filename, data, 0644)
}

func (repository *FsRepository) GetAll() ([]mcpAPI.MCPServer, error) {
	repository.mutex.RLock()
	defer repository.mutex.RUnlock()

	files, err := repository.listFiles()
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return []mcpAPI.MCPServer{}, nil
	}

	servers := make([]mcpAPI.MCPServer, 0, len(files))
	for _, file := range files {
		server, err := repository.loadFile(file)
		if err != nil {
			return nil, err
		}
		servers = append(servers, server)
	}

	sort.Slice(servers, func(i, j int) bool {
		return servers[i].Name < servers[j].Name
	})

	return servers, nil
}

func (repository *FsRepository) AddMCPServer(server mcpAPI.MCPServer) error {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	filename := uuid.NewString() + ".json"
	return repository.saveFile(filename, server)
}

func (repository *FsRepository) RemoveAll() error {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	files, err := repository.listFiles()
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := repository.fs.Remove(file); err != nil {
			return err
		}
	}

	return nil
}
