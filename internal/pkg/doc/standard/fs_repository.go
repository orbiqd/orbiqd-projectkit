package standard

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/google/uuid"
	standardAPI "github.com/orbiqd/orbiqd-projectkit/pkg/doc/standard"
	projectAPI "github.com/orbiqd/orbiqd-projectkit/pkg/project"
	"github.com/spf13/afero"
)

type FsRepository struct {
	mutex sync.RWMutex
	fs    afero.Fs
}

var _ standardAPI.Repository = (*FsRepository)(nil)

func NewFsRepository(fs afero.Fs) *FsRepository {
	return &FsRepository{
		mutex: sync.RWMutex{},
		fs:    fs,
	}
}

func NewFsRepositoryProvider() func(projectAPI.Fs) (standardAPI.Repository, error) {
	return func(projectFs projectAPI.Fs) (standardAPI.Repository, error) {
		dir := ".projectkit/repository/doc/standard"

		if err := projectFs.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("standard repository directory creation: %w", err)
		}

		scopedFs := afero.NewBasePathFs(projectFs, dir)
		return NewFsRepository(scopedFs), nil
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

func (repository *FsRepository) loadFile(filename string) (standardAPI.Standard, error) {
	data, err := afero.ReadFile(repository.fs, filename)
	if err != nil {
		return standardAPI.Standard{}, err
	}

	var standard standardAPI.Standard
	if err := json.Unmarshal(data, &standard); err != nil {
		return standardAPI.Standard{}, err
	}

	return standard, nil
}

func (repository *FsRepository) saveFile(filename string, standard standardAPI.Standard) error {
	data, err := json.Marshal(standard)
	if err != nil {
		return err
	}

	return afero.WriteFile(repository.fs, filename, data, 0644)
}

func (repository *FsRepository) GetAll() ([]standardAPI.Standard, error) {
	repository.mutex.RLock()
	defer repository.mutex.RUnlock()

	files, err := repository.listFiles()
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return []standardAPI.Standard{}, nil
	}

	standards := make([]standardAPI.Standard, 0, len(files))
	for _, file := range files {
		standard, err := repository.loadFile(file)
		if err != nil {
			return nil, err
		}
		standards = append(standards, standard)
	}

	sort.Slice(standards, func(i, j int) bool {
		return standards[i].Metadata.Name < standards[j].Metadata.Name
	})

	return standards, nil
}

func (repository *FsRepository) AddStandard(standard standardAPI.Standard) error {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	filename := uuid.NewString() + ".json"
	return repository.saveFile(filename, standard)
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
