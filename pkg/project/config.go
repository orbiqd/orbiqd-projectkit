package project

import (
	agentAPI "github.com/orbiqd/orbiqd-projectkit/pkg/agent"
	aiAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai"
)

// RulebookSourceConfig defines a single rulebook source entry.
type RulebookSourceConfig struct {
	URI string `json:"uri" validate:"required,uri"`
}

// RulebookConfig defines the rulebook configuration.
type RulebookConfig struct {
	Sources []RulebookSourceConfig `json:"sources,omitempty" validate:"required,min=1,dive"`
}

// Config defines the project configuration.
type Config struct {
	Agents   []agentAPI.Config `json:"agents,omitempty" validate:"omitempty,dive"`
	Rulebook *RulebookConfig   `json:"rulebook,omitempty" validate:"omitempty"`
	AI       *aiAPI.Config     `json:"ai,omitempty" validate:"omitempty"`
}
