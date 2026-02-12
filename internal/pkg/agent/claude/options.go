package claude

type Options struct {
	InstructionsFileName   string `json:"instructionsFileName" validate:"required" default:"CLAUDE.md"`
	ProjectSettingsDirName string `json:"projectSettingsDirName" validate:"required" default:".claude"`
	SkillsDirName          string `json:"skillsDirName" validate:"required" default:"skills"`
	MCPFileName            string `json:"mcpFileName" validate:"required" default:".mcp.json"`
}
