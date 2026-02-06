package agent

import "errors"

type Registry interface {
	GetAllKinds() []Kind
	Register(provider Provider) error
	GetByKind(kind Kind) (Provider, error)
}

var (
	ErrProviderNotRegistered     = errors.New("provider not registered")
	ErrProviderAlreadyRegistered = errors.New("provider already registered")
)
