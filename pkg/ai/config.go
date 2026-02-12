package ai

import (
	"github.com/orbiqd/orbiqd-projectkit/pkg/ai/instruction"
	"github.com/orbiqd/orbiqd-projectkit/pkg/ai/mcp"
	"github.com/orbiqd/orbiqd-projectkit/pkg/ai/skill"
	"github.com/orbiqd/orbiqd-projectkit/pkg/ai/workflow"
)

type Config struct {
	Instruction *instruction.Config `json:"instruction,omitempty" validate:"omitempty"`
	Skill       *skill.Config       `json:"skill,omitempty" validate:"omitempty"`
	Workflows   *workflow.Config    `json:"workflow,omitempty" validate:"omitempty"`
	MCP         *mcp.Config         `json:"mcp,omitempty" validate:"omitempty"`
}
