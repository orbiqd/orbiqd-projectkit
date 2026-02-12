package standard

type SourceConfig struct {
	URI string `json:"uri" validate:"required,uri"`
}

type RenderConfig struct {
	Destination string `json:"destination" validate:"required"`
	Format      string `json:"format" validate:"required"`
}

type Config struct {
	Render  []RenderConfig `json:"render,omitempty" validate:"omitempty,dive"`
	Sources []SourceConfig `json:"sources,omitempty" validate:"omitempty,min=1,dive"`
}
