package agent

type Config struct {
	Kind    Kind `json:"kind" validate:"required"`
	Options any  `json:"options,omitempty"`
}
