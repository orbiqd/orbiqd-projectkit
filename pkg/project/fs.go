package project

import "github.com/spf13/afero"

// Fs represents the project filesystem abstraction.
type Fs interface {
	afero.Fs
}
