package source

import (
	"fmt"
	"path/filepath"
	"strings"

	sourceAPI "github.com/orbiqd/orbiqd-projectkit/pkg/source"
	"github.com/spf13/afero"
)

type LocalDriverOpt func(*LocalDriver)

func WithRootFs(fs afero.Fs) LocalDriverOpt {
	return func(driver *LocalDriver) {
		driver.rootFs = fs
	}
}

type LocalDriver struct {
	rootFs afero.Fs
}

var _ sourceAPI.Driver = (*LocalDriver)(nil)

func NewLocalDriver(opts ...LocalDriverOpt) *LocalDriver {
	driver := &LocalDriver{
		rootFs: afero.NewOsFs(),
	}

	for _, opt := range opts {
		opt(driver)
	}

	return driver
}

func (driver *LocalDriver) GetSupportedSchemes() []string {
	return []string{"local"}
}

func (driver *LocalDriver) Resolve(uri string) (afero.Fs, error) {
	path, found := strings.CutPrefix(uri, "local://")
	if !found {
		return nil, fmt.Errorf("uri %s: %w", uri, sourceAPI.ErrUnsupportedScheme)
	}

	if path == "" {
		return nil, fmt.Errorf("empty path in uri %s", uri)
	}

	path = filepath.Clean(path)

	exists, err := afero.DirExists(driver.rootFs, path)
	if err != nil {
		return nil, fmt.Errorf("checking path %s: %w", path, err)
	}
	if !exists {
		return nil, fmt.Errorf("path %s does not exist", path)
	}

	scopedFs := afero.NewBasePathFs(driver.rootFs, path)
	readOnlyFs := afero.NewReadOnlyFs(scopedFs)

	return readOnlyFs, nil
}
