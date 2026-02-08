package skill

import (
	"sync"

	skillAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/skill"
)

type MemoryRepository struct {
	mutex  sync.RWMutex
	skills map[skillAPI.Name]skillAPI.Skill
}

var _ skillAPI.Repository = (*MemoryRepository)(nil)

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		mutex:  sync.RWMutex{},
		skills: make(map[skillAPI.Name]skillAPI.Skill),
	}
}

func (repository *MemoryRepository) GetSkillByName(name skillAPI.Name) (*skillAPI.Skill, error) {
	repository.mutex.RLock()
	defer repository.mutex.RUnlock()

	skill, found := repository.skills[name]
	if !found {
		return nil, skillAPI.ErrSkillNotFound
	}

	return &skill, nil
}

func (repository *MemoryRepository) AddSkill(skill skillAPI.Skill) error {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	if _, exists := repository.skills[skill.Metadata.Name]; exists {
		return skillAPI.ErrSkillAlreadyExists
	}

	repository.skills[skill.Metadata.Name] = skill

	return nil
}
