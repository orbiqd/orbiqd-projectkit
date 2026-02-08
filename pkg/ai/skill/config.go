package skill

type SourceConfig struct {
	URI string `json:"uri" validate:"required,uri"`
}

type Config struct {
	Sources []SourceConfig `json:"sources,omitempty" validate:"required,min=1,dive"`
}
