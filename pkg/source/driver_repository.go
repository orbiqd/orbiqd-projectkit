package source

import "errors"

type DriverRepository interface {
	RegisterDriver(driver Driver) error
	GetDriverByScheme(scheme string) (Driver, error)
}

var ErrSchemeDriverAlreadyRegistered = errors.New("scheme driver already registered")
var ErrSchemeDriverNotRegistered = errors.New("scheme driver not registered")
