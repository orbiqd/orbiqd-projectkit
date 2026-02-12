package standard

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStandardId_Validate_WhenValidKebabCase_ThenReturnsNoError(t *testing.T) {
	testCases := []struct {
		name string
		id   StandardId
	}{
		{name: "simple", id: "simple"},
		{name: "with-dashes", id: "with-dashes"},
		{name: "my-coding-style", id: "my-coding-style"},
		{name: "v1-2-3", id: "v1-2-3"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.id.Validate()

			assert.NoError(t, err)
		})
	}
}

func TestStandardId_Validate_WhenInvalidFormat_ThenReturnsError(t *testing.T) {
	testCases := []struct {
		name string
		id   StandardId
	}{
		{name: "uppercase", id: "My-Tag"},
		{name: "space", id: "my tag"},
		{name: "underscore", id: "my_tag"},
		{name: "leading-dash", id: "-tag"},
		{name: "trailing-dash", id: "tag-"},
		{name: "double-dash", id: "my--tag"},
		{name: "camelCase", id: "myTag"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.id.Validate()

			assert.ErrorIs(t, err, ErrStandardInvalidID)
		})
	}
}

func TestStandardId_Validate_WhenEmpty_ThenReturnsError(t *testing.T) {
	id := StandardId("")

	err := id.Validate()

	assert.ErrorIs(t, err, ErrStandardInvalidID)
}
