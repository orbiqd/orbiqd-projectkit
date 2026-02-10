package codex

type Options struct {
	InstructionsFileName   string `json:"instructionsFileName" validate:"required" default:"AGENTS.md"`
	ProjectSettingsDirName string `json:"projectSettingsDirName" validate:"required" default:".agents"`
	SkillsDirName          string `json:"skillsDirName" validate:"required" default:"skills"`
}
