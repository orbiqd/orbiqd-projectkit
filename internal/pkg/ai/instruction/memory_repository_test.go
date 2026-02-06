package instruction

import (
	"testing"

	instructionAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/instruction"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryRepository_AddInstructions_WhenSingleInstruction_ThenStoresRules(t *testing.T) {
	t.Parallel()

	repo := NewMemoryRepository()
	instructions := instructionAPI.Instructions{
		Category: "coding",
		Rules:    []instructionAPI.Rule{"rule1", "rule2"},
	}

	err := repo.AddInstructions(instructions)
	require.NoError(t, err)

	result, err := repo.GetAll()
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, instructionAPI.Category("coding"), result[0].Category)
	assert.Equal(t, []instructionAPI.Rule{"rule1", "rule2"}, result[0].Rules)
}

func TestMemoryRepository_AddInstructions_WhenMultipleCategories_ThenStoresEachSeparately(t *testing.T) {
	t.Parallel()

	repo := NewMemoryRepository()
	instructions1 := instructionAPI.Instructions{
		Category: "coding",
		Rules:    []instructionAPI.Rule{"rule1"},
	}
	instructions2 := instructionAPI.Instructions{
		Category: "testing",
		Rules:    []instructionAPI.Rule{"rule2"},
	}

	err := repo.AddInstructions(instructions1)
	require.NoError(t, err)

	err = repo.AddInstructions(instructions2)
	require.NoError(t, err)

	result, err := repo.GetAll()
	require.NoError(t, err)
	require.Len(t, result, 2)

	categoriesFound := make(map[instructionAPI.Category][]instructionAPI.Rule)
	for _, inst := range result {
		categoriesFound[inst.Category] = inst.Rules
	}

	assert.Equal(t, []instructionAPI.Rule{"rule1"}, categoriesFound["coding"])
	assert.Equal(t, []instructionAPI.Rule{"rule2"}, categoriesFound["testing"])
}

func TestMemoryRepository_AddInstructions_WhenSameCategoryAddedTwice_ThenAppendsRules(t *testing.T) {
	t.Parallel()

	repo := NewMemoryRepository()
	instructions1 := instructionAPI.Instructions{
		Category: "coding",
		Rules:    []instructionAPI.Rule{"rule1"},
	}
	instructions2 := instructionAPI.Instructions{
		Category: "coding",
		Rules:    []instructionAPI.Rule{"rule2", "rule3"},
	}

	err := repo.AddInstructions(instructions1)
	require.NoError(t, err)

	err = repo.AddInstructions(instructions2)
	require.NoError(t, err)

	result, err := repo.GetAll()
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, instructionAPI.Category("coding"), result[0].Category)
	assert.Equal(t, []instructionAPI.Rule{"rule1", "rule2", "rule3"}, result[0].Rules)
}

func TestMemoryRepository_GetAll_WhenEmpty_ThenReturnsEmptySlice(t *testing.T) {
	t.Parallel()

	repo := NewMemoryRepository()
	result, err := repo.GetAll()

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestMemoryRepository_GetAll_WhenHasInstructions_ThenReturnsAll(t *testing.T) {
	t.Parallel()

	repo := NewMemoryRepository()
	instructions1 := instructionAPI.Instructions{
		Category: "coding",
		Rules:    []instructionAPI.Rule{"rule1"},
	}
	instructions2 := instructionAPI.Instructions{
		Category: "testing",
		Rules:    []instructionAPI.Rule{"rule2"},
	}

	err := repo.AddInstructions(instructions1)
	require.NoError(t, err)

	err = repo.AddInstructions(instructions2)
	require.NoError(t, err)

	result, err := repo.GetAll()

	require.NoError(t, err)
	require.Len(t, result, 2)

	categoriesFound := make(map[instructionAPI.Category]bool)
	for _, inst := range result {
		categoriesFound[inst.Category] = true
	}

	assert.True(t, categoriesFound["coding"])
	assert.True(t, categoriesFound["testing"])
}
