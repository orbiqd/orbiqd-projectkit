package instruction

import (
	"sync"

	instructionAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/instruction"
)

// MemoryRepository stores instructions in memory.
type MemoryRepository struct {
	mutex      sync.RWMutex
	categories map[instructionAPI.Category][]instructionAPI.Rule
}

var _ instructionAPI.Repository = (*MemoryRepository)(nil)

// NewMemoryRepository creates a new in-memory instruction repository.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		mutex:      sync.RWMutex{},
		categories: make(map[instructionAPI.Category][]instructionAPI.Rule),
	}
}

// GetAll returns all stored instruction sets.
func (repository *MemoryRepository) GetAll() ([]instructionAPI.Instructions, error) {
	repository.mutex.RLock()
	defer repository.mutex.RUnlock()

	result := make([]instructionAPI.Instructions, 0, len(repository.categories))
	for category, rules := range repository.categories {
		result = append(result, instructionAPI.Instructions{
			Category: category,
			Rules:    rules,
		})
	}

	return result, nil
}

// AddInstructions stores the provided instruction set.
func (repository *MemoryRepository) AddInstructions(instructions instructionAPI.Instructions) error {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	repository.categories[instructions.Category] = append(repository.categories[instructions.Category], instructions.Rules...)

	return nil
}
