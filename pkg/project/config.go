package project

import (
	agentAPI "github.com/orbiqd/orbiqd-projectkit/pkg/agent"
	aiAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai"
	docAPI "github.com/orbiqd/orbiqd-projectkit/pkg/doc"
	"github.com/orbiqd/orbiqd-projectkit/pkg/rulebook"
)

// Config defines the project configuration.
type Config struct {
	Agents   []agentAPI.Config `json:"agents,omitempty" validate:"omitempty,dive"`
	Rulebook *rulebook.Config  `json:"rulebook,omitempty" validate:"omitempty"`
	AI       *aiAPI.Config     `json:"ai,omitempty" validate:"omitempty"`
	Docs     *docAPI.Config    `json:"docs,omitempty" validate:"omitempty"`
}
