package rulebook

import (
	"github.com/orbiqd/orbiqd-projectkit/pkg/ai"
	"github.com/orbiqd/orbiqd-projectkit/pkg/ai/instruction"
	"github.com/orbiqd/orbiqd-projectkit/pkg/ai/skill"
)

type AiRulebook struct {
	Instructions []instruction.Instructions
	Skills       []skill.Skill
}

type Rulebook struct {
	AI AiRulebook
}

type Metadata struct {
	AI *ai.Config `json:"ai" validate:"omitempty"`
}
