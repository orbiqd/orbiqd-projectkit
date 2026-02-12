package workflow

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var semverRegex = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

func NewValidator() *validator.Validate {
	v := validator.New()

	_ = v.RegisterValidation("semver", validateSemver)

	return v
}

func validateSemver(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return semverRegex.MatchString(value)
}

func Validate(w Workflow) error {
	v := NewValidator()
	if err := v.Struct(w); err != nil {
		return err
	}
	return nil
}
