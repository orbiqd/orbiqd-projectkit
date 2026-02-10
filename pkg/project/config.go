package project

import (
	agentAPI "github.com/orbiqd/orbiqd-projectkit/pkg/agent"
	aiAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai"
	"github.com/orbiqd/orbiqd-projectkit/pkg/rulebook"
)

// Config defines the project configuration.
type Config struct {
	Agents   []agentAPI.Config `json:"agents,omitempty" validate:"omitempty,dive"`
	Rulebook *rulebook.Config  `json:"rulebook,omitempty" validate:"omitempty"`
	AI       *aiAPI.Config     `json:"ai,omitempty" validate:"omitempty"`
}
