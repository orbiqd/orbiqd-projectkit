package source

import (
	"errors"

	"github.com/spf13/afero"
)

type Driver interface {
	GetSupportedSchemes() []string
	Resolve(uri string) (afero.Fs, error)
}

var ErrUnsupportedScheme = errors.New("unsupported scheme")
