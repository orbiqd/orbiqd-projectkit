package rulebook

import (
	"github.com/orbiqd/orbiqd-projectkit/pkg/ai"
	"github.com/orbiqd/orbiqd-projectkit/pkg/ai/instruction"
	"github.com/orbiqd/orbiqd-projectkit/pkg/ai/mcp"
	"github.com/orbiqd/orbiqd-projectkit/pkg/ai/skill"
	"github.com/orbiqd/orbiqd-projectkit/pkg/ai/workflow"
	"github.com/orbiqd/orbiqd-projectkit/pkg/doc"
	"github.com/orbiqd/orbiqd-projectkit/pkg/doc/standard"
)

type AiRulebook struct {
	Instructions []instruction.Instructions
	Skills       []skill.Skill
	Workflows    []workflow.Workflow
	MCPServers   []mcp.MCPServer
}

type DocRulebook struct {
	Standards []standard.Standard
}

type Rulebook struct {
	AI  AiRulebook
	Doc DocRulebook
}

type Metadata struct {
	AI  *ai.Config  `json:"ai" validate:"omitempty"`
	Doc *doc.Config `json:"doc" validate:"omitempty"`
}
