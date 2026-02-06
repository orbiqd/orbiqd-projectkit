package source

import (
	"fmt"
	"sync"

	sourceAPI "github.com/orbiqd/orbiqd-projectkit/pkg/source"
)

type DriverRepository struct {
	mutex   sync.RWMutex
	drivers map[string]sourceAPI.Driver
}

var _ sourceAPI.DriverRepository = (*DriverRepository)(nil)

func NewDriverRepository() *DriverRepository {
	return &DriverRepository{
		drivers: make(map[string]sourceAPI.Driver),
	}
}

func (repository *DriverRepository) RegisterDriver(driver sourceAPI.Driver) error {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	schemes := driver.GetSupportedSchemes()

	for _, scheme := range schemes {
		if _, exists := repository.drivers[scheme]; exists {
			return fmt.Errorf("scheme %s: %w", scheme, sourceAPI.ErrSchemeDriverAlreadyRegistered)
		}
	}

	for _, scheme := range schemes {
		repository.drivers[scheme] = driver
	}

	return nil
}

func (repository *DriverRepository) GetDriverByScheme(scheme string) (sourceAPI.Driver, error) {
	repository.mutex.RLock()
	defer repository.mutex.RUnlock()

	driver, exists := repository.drivers[scheme]
	if !exists {
		return nil, fmt.Errorf("scheme %s: %w", scheme, sourceAPI.ErrSchemeDriverNotRegistered)
	}

	return driver, nil
}
