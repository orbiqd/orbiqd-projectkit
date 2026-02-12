package standard

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var (
	semverRegex     = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)
	iso639_1Regex   = regexp.MustCompile(`^[a-z]{2}$`)
	kebabCaseRegex  = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	nameFormatRegex = regexp.MustCompile(`^[a-zA-Z0-9\s\-]+$`)
)

func NewValidator() *validator.Validate {
	v := validator.New()

	_ = v.RegisterValidation("semver", validateSemver)
	_ = v.RegisterValidation("iso639_1", validateISO639_1)
	_ = v.RegisterValidation("kebab_case", validateKebabCase)
	_ = v.RegisterValidation("name_format", validateNameFormat)

	return v
}

func validateSemver(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return semverRegex.MatchString(value)
}

func validateISO639_1(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return iso639_1Regex.MatchString(value)
}

func validateKebabCase(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return kebabCaseRegex.MatchString(value)
}

func validateNameFormat(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return nameFormatRegex.MatchString(value)
}

func Validate(s Standard) error {
	v := NewValidator()
	if err := v.Struct(s); err != nil {
		return err
	}
	return nil
}
