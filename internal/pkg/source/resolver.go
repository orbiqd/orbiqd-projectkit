package source

import (
	"errors"
	"fmt"
	"strings"

	sourceAPI "github.com/orbiqd/orbiqd-projectkit/pkg/source"
	"github.com/spf13/afero"
)

type Resolver struct {
	driverRepository sourceAPI.DriverRepository
}

var _ sourceAPI.Resolver = (*Resolver)(nil)

func NewResolver(driverRepository sourceAPI.DriverRepository) *Resolver {
	return &Resolver{
		driverRepository: driverRepository,
	}
}

func (resolver *Resolver) Resolve(uri string) (afero.Fs, error) {
	scheme, _, found := strings.Cut(uri, "://")
	if !found {
		return nil, errors.New("uri scheme not found")
	}

	driver, err := resolver.driverRepository.GetDriverByScheme(scheme)
	if err != nil {
		return nil, fmt.Errorf("get driver by scheme: %w", err)
	}

	fs, err := driver.Resolve(uri)
	if err != nil {
		return nil, fmt.Errorf("resolve %s: %w", uri, err)
	}

	return fs, nil
}
