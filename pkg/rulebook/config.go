package rulebook

type SourceConfig struct {
	URI string `json:"uri"`
}

// Config defines the rulebook configuration.
type Config struct {
	Sources []SourceConfig `json:"sources,omitempty" validate:"required,min=1,dive"`
}
