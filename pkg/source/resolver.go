package source

import "github.com/spf13/afero"

type Resolver interface {
	Resolve(uri string) (afero.Fs, error)
}
