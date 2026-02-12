package doc

import "github.com/orbiqd/orbiqd-projectkit/pkg/doc/standard"

type Config struct {
	Standard *standard.Config `json:"standard,omitempty" validate:"omitempty"`
}
