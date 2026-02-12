package standard

import "errors"

var ErrStandardInvalidID = errors.New("standard id must be in kebab-case format")

func (id StandardId) Validate() error {
	if !kebabCaseRegex.MatchString(string(id)) {
		return ErrStandardInvalidID
	}
	return nil
}

type Repository interface {
	GetAll() ([]Standard, error)
	AddStandard(standard Standard) error
	RemoveAll() error
}
