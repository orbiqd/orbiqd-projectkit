package codex

type Options struct {
	InstructionsFileName string `json:"instructionsFileName" validate:"required" default:"AGENTS.md"`
}
