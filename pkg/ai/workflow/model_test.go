package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkflowId_Validate_WhenValidID_ThenReturnsNoError(t *testing.T) {
	t.Parallel()

	validIDs := []WorkflowId{
		"simple",
		"with-dashes",
		"with123numbers",
		"MixedCase",
		"a",
		"123",
		"a-b-c-1-2-3",
	}

	for _, id := range validIDs {
		err := id.Validate()
		assert.NoError(t, err, "ID %s should be valid", id)
	}
}

func TestWorkflowId_Validate_WhenInvalidCharacters_ThenReturnsError(t *testing.T) {
	t.Parallel()

	invalidIDs := []WorkflowId{
		"with/slash",
		"with space",
		"with.dot",
		"with_underscore",
		"with@special",
		"with#hash",
		"with$dollar",
		"with%percent",
		"with&ampersand",
		"with*asterisk",
		"with(parens)",
		"with[brackets]",
		"with{braces}",
		"with|pipe",
		"with\\backslash",
		"with:colon",
		"with;semicolon",
		"with'quote",
		"with\"doublequote",
		"with<less",
		"with>greater",
		"with?question",
		"with!exclamation",
		"with~tilde",
		"with`backtick",
		"with^caret",
		"with+plus",
		"with=equals",
		"with,comma",
	}

	for _, id := range invalidIDs {
		err := id.Validate()
		require.ErrorIs(t, err, ErrWorkflowInvalidID, "ID %s should be invalid", id)
	}
}

func TestWorkflowId_Validate_WhenEmptyString_ThenReturnsError(t *testing.T) {
	t.Parallel()

	id := WorkflowId("")
	err := id.Validate()
	require.ErrorIs(t, err, ErrWorkflowInvalidID)
}

func TestExecutionId_Validate_WhenValidID_ThenReturnsNoError(t *testing.T) {
	t.Parallel()

	validIDs := []ExecutionId{
		"simple",
		"with-dashes",
		"with123numbers",
		"MixedCase",
		"a",
		"123",
		"a-b-c-1-2-3",
	}

	for _, id := range validIDs {
		err := id.Validate()
		assert.NoError(t, err, "ID %s should be valid", id)
	}
}

func TestExecutionId_Validate_WhenInvalidCharacters_ThenReturnsError(t *testing.T) {
	t.Parallel()

	invalidIDs := []ExecutionId{
		"with/slash",
		"with space",
		"with.dot",
		"with_underscore",
		"with@special",
		"with#hash",
		"with$dollar",
		"with%percent",
		"with&ampersand",
		"with*asterisk",
		"with(parens)",
		"with[brackets]",
		"with{braces}",
		"with|pipe",
		"with\\backslash",
		"with:colon",
		"with;semicolon",
		"with'quote",
		"with\"doublequote",
		"with<less",
		"with>greater",
		"with?question",
		"with!exclamation",
		"with~tilde",
		"with`backtick",
		"with^caret",
		"with+plus",
		"with=equals",
		"with,comma",
	}

	for _, id := range invalidIDs {
		err := id.Validate()
		require.ErrorIs(t, err, ErrExecutionInvalidID, "ID %s should be invalid", id)
	}
}

func TestExecutionId_Validate_WhenEmptyString_ThenReturnsError(t *testing.T) {
	t.Parallel()

	id := ExecutionId("")
	err := id.Validate()
	require.ErrorIs(t, err, ErrExecutionInvalidID)
}
