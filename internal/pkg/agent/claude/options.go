package claude

type Options struct {
	InstructionsFileName string `json:"instructionsFileName" validate:"required" default:"CLAUDE.md"`
}
