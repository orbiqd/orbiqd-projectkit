package agent

import (
	instructionAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/instruction"
	skillAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/skill"
)

type Kind string

// Agent renders AI artifacts for a specific provider.
type Agent interface {
	// GetKind of an agent.
	GetKind() Kind

	// RenderInstructions renders agent instructions to the project.
	RenderInstructions(instructions []instructionAPI.Instructions) error

	// RebuildSkills removes existing rendered skills and renders them from the repository.
	RebuildSkills(skillRepository skillAPI.Repository) error

	// GitIgnorePatterns returns patterns that should be excluded from git-commit.
	GitIgnorePatterns() []string
}
