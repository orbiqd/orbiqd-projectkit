package skill

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/google/uuid"
	skillAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/skill"
	projectAPI "github.com/orbiqd/orbiqd-projectkit/pkg/project"
	"github.com/spf13/afero"
)

type FsRepository struct {
	mutex sync.RWMutex
	fs    afero.Fs
}

var _ skillAPI.Repository = (*FsRepository)(nil)

func NewFsRepository(fs afero.Fs) *FsRepository {
	return &FsRepository{
		mutex: sync.RWMutex{},
		fs:    fs,
	}
}

func NewFsRepositoryProvider() func(projectAPI.Fs) (skillAPI.Repository, error) {
	return func(projectFs projectAPI.Fs) (skillAPI.Repository, error) {
		dir := ".projectkit/repository/ai/skill"

		if err := projectFs.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("skill repository directory creation: %w", err)
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

func (repository *FsRepository) loadFile(filename string) (skillAPI.Skill, error) {
	data, err := afero.ReadFile(repository.fs, filename)
	if err != nil {
		return skillAPI.Skill{}, err
	}

	var skill skillAPI.Skill
	if err := json.Unmarshal(data, &skill); err != nil {
		return skillAPI.Skill{}, err
	}

	return skill, nil
}

func (repository *FsRepository) saveFile(filename string, skill skillAPI.Skill) error {
	data, err := json.Marshal(skill)
	if err != nil {
		return err
	}

	return afero.WriteFile(repository.fs, filename, data, 0644)
}

func (repository *FsRepository) GetSkillByName(name skillAPI.Name) (*skillAPI.Skill, error) {
	repository.mutex.RLock()
	defer repository.mutex.RUnlock()

	files, err := repository.listFiles()
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		skill, err := repository.loadFile(file)
		if err != nil {
			return nil, err
		}

		if skill.Metadata.Name == name {
			return &skill, nil
		}
	}

	return nil, skillAPI.ErrSkillNotFound
}

func (repository *FsRepository) GetAll() ([]skillAPI.Skill, error) {
	repository.mutex.RLock()
	defer repository.mutex.RUnlock()

	files, err := repository.listFiles()
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return []skillAPI.Skill{}, nil
	}

	skills := make([]skillAPI.Skill, 0, len(files))
	for _, file := range files {
		skill, err := repository.loadFile(file)
		if err != nil {
			return nil, err
		}
		skills = append(skills, skill)
	}

	sort.Slice(skills, func(i, j int) bool {
		return skills[i].Metadata.Name < skills[j].Metadata.Name
	})

	return skills, nil
}

func (repository *FsRepository) AddSkill(skill skillAPI.Skill) error {
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

		if existing.Metadata.Name == skill.Metadata.Name {
			return skillAPI.ErrSkillAlreadyExists
		}
	}

	filename := uuid.NewString() + ".json"
	return repository.saveFile(filename, skill)
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
