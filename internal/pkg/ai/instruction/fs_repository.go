package instruction

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/google/uuid"
	instructionAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/instruction"
	projectAPI "github.com/orbiqd/orbiqd-projectkit/pkg/project"
	"github.com/spf13/afero"
)

type FsRepository struct {
	mutex sync.RWMutex
	fs    afero.Fs
}

var _ instructionAPI.Repository = (*FsRepository)(nil)

func NewFsRepository(fs afero.Fs) *FsRepository {
	return &FsRepository{
		mutex: sync.RWMutex{},
		fs:    fs,
	}
}

func NewFsRepositoryProvider() func(projectAPI.Fs) (instructionAPI.Repository, error) {
	return func(projectFs projectAPI.Fs) (instructionAPI.Repository, error) {
		dir := ".projectkit/repository/ai/instruction"

		if err := projectFs.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("instruction repository directory creation: %w", err)
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

func (repository *FsRepository) loadFile(filename string) (instructionAPI.Instructions, error) {
	data, err := afero.ReadFile(repository.fs, filename)
	if err != nil {
		return instructionAPI.Instructions{}, err
	}

	var instructions instructionAPI.Instructions
	if err := json.Unmarshal(data, &instructions); err != nil {
		return instructionAPI.Instructions{}, err
	}

	return instructions, nil
}

func (repository *FsRepository) saveFile(filename string, instructions instructionAPI.Instructions) error {
	data, err := json.Marshal(instructions)
	if err != nil {
		return err
	}

	return afero.WriteFile(repository.fs, filename, data, 0644)
}

func (repository *FsRepository) GetAll() ([]instructionAPI.Instructions, error) {
	repository.mutex.RLock()
	defer repository.mutex.RUnlock()

	files, err := repository.listFiles()
	if err != nil {
		return nil, err
	}

	var result []instructionAPI.Instructions
	for _, file := range files {
		instructions, err := repository.loadFile(file)
		if err != nil {
			return nil, err
		}
		result = append(result, instructions)
	}

	return result, nil
}

func (repository *FsRepository) AddInstructions(instructions instructionAPI.Instructions) error {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	files, err := repository.listFiles()
	if err != nil {
		return err
	}

	for _, file := range files {
		existing, err := repository.loadFile(file)
		if err != nil {
			return err
		}

		if existing.Category == instructions.Category {
			existing.Rules = append(existing.Rules, instructions.Rules...)
			return repository.saveFile(file, existing)
		}
	}

	filename := uuid.NewString() + ".json"
	return repository.saveFile(filename, instructions)
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
