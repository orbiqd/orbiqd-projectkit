package agent

import instructionAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/instruction"

type Kind string

type Agent interface {
	RenderInstructions(instructions []instructionAPI.Instructions) error

	GitIgnorePatterns() []string
}
