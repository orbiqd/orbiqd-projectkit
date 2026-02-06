package project

import "errors"

// ConfigLoader loads project configuration.
type ConfigLoader interface {
	// Load returns the resolved project configuration.
	Load() (*Config, error)
}

var defaultLoaderFactory func() ConfigLoader

// RegisterDefaultLoader sets the factory used for the default config loader.
func RegisterDefaultLoader(factory func() ConfigLoader) {
	defaultLoaderFactory = factory
}

// DefaultConfigLoader returns a new default config loader instance.
func DefaultConfigLoader() (ConfigLoader, error) {
	if defaultLoaderFactory == nil {
		return nil, ErrNoDefaultConfigLoaderRegistered
	}
	return defaultLoaderFactory(), nil
}

// ErrNoDefaultConfigLoaderRegistered means no default loader factory was registered.
var ErrNoDefaultConfigLoaderRegistered = errors.New("no default config loader registered")

// ErrNoConfigResolved means no configuration file was found.
var ErrNoConfigResolved = errors.New("config not found")

// ErrConfigLoadFailed means the configuration file could not be loaded.
var ErrConfigLoadFailed = errors.New("config load failed")

// ErrConfigValidationFailed means the configuration failed validation.
var ErrConfigValidationFailed = errors.New("config validation failed")
