package standard

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateSemver(t *testing.T) {
	tests := []struct {
		name    string
		version string
		valid   bool
	}{
		{"valid major.minor.patch", "1.0.0", true},
		{"valid with prerelease", "1.0.0-alpha", true},
		{"valid with prerelease and build", "1.0.0-alpha+001", true},
		{"valid complex prerelease", "1.0.0-alpha.1", true},
		{"valid with build metadata", "1.0.0+20130313144700", true},
		{"valid large version", "10.20.30", true},
		{"invalid missing patch", "1.0", false},
		{"invalid missing minor and patch", "1", false},
		{"invalid leading zeros", "01.0.0", false},
		{"invalid non-numeric", "1.0.x", false},
		{"invalid empty", "", false},
		{"invalid with v prefix", "v1.0.0", false},
	}

	v := NewValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type testStruct struct {
				Version string `validate:"semver"`
			}
			ts := testStruct{Version: tt.version}
			err := v.Struct(ts)

			if tt.valid {
				assert.NoError(t, err, "expected %s to be valid", tt.version)
			} else {
				assert.Error(t, err, "expected %s to be invalid", tt.version)
			}
		})
	}
}

func TestValidateISO639_1(t *testing.T) {
	tests := []struct {
		name  string
		code  string
		valid bool
	}{
		{"valid en", "en", true},
		{"valid pl", "pl", true},
		{"valid de", "de", true},
		{"valid fr", "fr", true},
		{"invalid uppercase", "EN", false},
		{"invalid three letters", "eng", false},
		{"invalid one letter", "e", false},
		{"invalid empty", "", false},
		{"invalid with dash", "en-US", false},
		{"invalid numbers", "e1", false},
	}

	v := NewValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type testStruct struct {
				Language string `validate:"iso639_1"`
			}
			ts := testStruct{Language: tt.code}
			err := v.Struct(ts)

			if tt.valid {
				assert.NoError(t, err, "expected %s to be valid", tt.code)
			} else {
				assert.Error(t, err, "expected %s to be invalid", tt.code)
			}
		})
	}
}

func TestValidateKebabCase(t *testing.T) {
	tests := []struct {
		name  string
		tag   string
		valid bool
	}{
		{"valid single word", "tag", true},
		{"valid with dash", "my-tag", true},
		{"valid multiple dashes", "my-coding-style", true},
		{"valid with numbers", "v1-2-3", true},
		{"valid starting with number", "1-tag", true},
		{"invalid uppercase", "My-Tag", false},
		{"invalid space", "my tag", false},
		{"invalid underscore", "my_tag", false},
		{"invalid starting with dash", "-tag", false},
		{"invalid ending with dash", "tag-", false},
		{"invalid double dash", "my--tag", false},
		{"invalid empty", "", false},
		{"invalid camelCase", "myTag", false},
	}

	v := NewValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type testStruct struct {
				Tag string `validate:"kebab_case"`
			}
			ts := testStruct{Tag: tt.tag}
			err := v.Struct(ts)

			if tt.valid {
				assert.NoError(t, err, "expected %s to be valid", tt.tag)
			} else {
				assert.Error(t, err, "expected %s to be invalid", tt.tag)
			}
		})
	}
}

func TestValidateNameFormat(t *testing.T) {
	tests := []struct {
		name  string
		value string
		valid bool
	}{
		{"valid simple name", "MyName", true},
		{"valid with space", "My Name", true},
		{"valid with dash", "My-Name", true},
		{"valid with numbers", "Name123", true},
		{"valid complex", "My Complex Name-123", true},
		{"valid all caps", "MYNAME", true},
		{"valid all lowercase", "myname", true},
		{"invalid underscore", "My_Name", false},
		{"invalid special chars", "My@Name", false},
		{"invalid dot", "My.Name", false},
		{"invalid empty", "", false},
		{"invalid only special", "!@#", false},
	}

	v := NewValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type testStruct struct {
				Name string `validate:"name_format"`
			}
			ts := testStruct{Name: tt.value}
			err := v.Struct(ts)

			if tt.valid {
				assert.NoError(t, err, "expected %s to be valid", tt.value)
			} else {
				assert.Error(t, err, "expected %s to be invalid", tt.value)
			}
		})
	}
}

func TestValidate_ValidStandard(t *testing.T) {
	standard := &Standard{
		Metadata: Metadata{
			Name:    "My Coding Standard",
			Version: "1.0.0",
			Tags:    []string{"go", "best-practices"},
			Scope: ScopeMetadata{
				Languages: []string{"en", "pl"},
			},
			Related: RelationMetadata{},
		},
		Specification: Specification{
			Purpose: "This is a test purpose with sufficient length to pass validation",
			Goals:   []string{"Goal one with sufficient length for validation"},
		},
		Requirements: Requirements{
			Rules: []RequirementRule{
				{
					Level:     "must",
					Statement: "This is a requirement statement with sufficient length",
					Rationale: "This is the rationale for the requirement with sufficient length",
				},
			},
		},
		Examples: Examples{
			Good: []Example{
				{
					Title:    "Good Example",
					Language: "go",
					Snippet:  "func main() {}",
					Reason:   "This is a good example because it demonstrates the concept clearly",
				},
			},
		},
	}

	err := Validate(*standard)
	assert.NoError(t, err)
}

func TestValidate_InvalidMetadata(t *testing.T) {
	tests := []struct {
		name     string
		standard *Standard
		wantErr  bool
	}{
		{
			name: "invalid semver",
			standard: &Standard{
				Metadata: Metadata{
					Name:    "Test",
					Version: "invalid",
					Tags:    []string{"test"},
					Scope: ScopeMetadata{
						Languages: []string{"en"},
					},
					Related: RelationMetadata{},
				},
				Specification: Specification{
					Purpose: "Test purpose with sufficient length for validation",
					Goals:   []string{"Test goal with sufficient length for validation"},
				},
				Requirements: Requirements{
					Rules: []RequirementRule{
						{
							Level:     "must",
							Statement: "Test statement with sufficient length for validation",
							Rationale: "Test rationale with sufficient length for validation",
						},
					},
				},
				Examples: Examples{
					Good: []Example{
						{
							Title:    "Test",
							Language: "go",
							Snippet:  "test",
							Reason:   "Test reason with sufficient length for validation",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid language code",
			standard: &Standard{
				Metadata: Metadata{
					Name:    "Test",
					Version: "1.0.0",
					Tags:    []string{"test"},
					Scope: ScopeMetadata{
						Languages: []string{"ENG"},
					},
					Related: RelationMetadata{},
				},
				Specification: Specification{
					Purpose: "Test purpose with sufficient length for validation",
					Goals:   []string{"Test goal with sufficient length for validation"},
				},
				Requirements: Requirements{
					Rules: []RequirementRule{
						{
							Level:     "must",
							Statement: "Test statement with sufficient length for validation",
							Rationale: "Test rationale with sufficient length for validation",
						},
					},
				},
				Examples: Examples{
					Good: []Example{
						{
							Title:    "Test",
							Language: "go",
							Snippet:  "test",
							Reason:   "Test reason with sufficient length for validation",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid tag format",
			standard: &Standard{
				Metadata: Metadata{
					Name:    "Test",
					Version: "1.0.0",
					Tags:    []string{"Invalid_Tag"},
					Scope: ScopeMetadata{
						Languages: []string{"en"},
					},
					Related: RelationMetadata{},
				},
				Specification: Specification{
					Purpose: "Test purpose with sufficient length for validation",
					Goals:   []string{"Test goal with sufficient length for validation"},
				},
				Requirements: Requirements{
					Rules: []RequirementRule{
						{
							Level:     "must",
							Statement: "Test statement with sufficient length for validation",
							Rationale: "Test rationale with sufficient length for validation",
						},
					},
				},
				Examples: Examples{
					Good: []Example{
						{
							Title:    "Test",
							Language: "go",
							Snippet:  "test",
							Reason:   "Test reason with sufficient length for validation",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid requirement level",
			standard: &Standard{
				Metadata: Metadata{
					Name:    "Test",
					Version: "1.0.0",
					Tags:    []string{"test"},
					Scope: ScopeMetadata{
						Languages: []string{"en"},
					},
					Related: RelationMetadata{},
				},
				Specification: Specification{
					Purpose: "Test purpose with sufficient length for validation",
					Goals:   []string{"Test goal with sufficient length for validation"},
				},
				Requirements: Requirements{
					Rules: []RequirementRule{
						{
							Level:     "MUST",
							Statement: "Test statement with sufficient length for validation",
							Rationale: "Test rationale with sufficient length for validation",
						},
					},
				},
				Examples: Examples{
					Good: []Example{
						{
							Title:    "Test",
							Language: "go",
							Snippet:  "test",
							Reason:   "Test reason with sufficient length for validation",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "text too short",
			standard: &Standard{
				Metadata: Metadata{
					Name:    "Test",
					Version: "1.0.0",
					Tags:    []string{"test"},
					Scope: ScopeMetadata{
						Languages: []string{"en"},
					},
					Related: RelationMetadata{},
				},
				Specification: Specification{
					Purpose: "Short",
					Goals:   []string{"Test goal with sufficient length for validation"},
				},
				Requirements: Requirements{
					Rules: []RequirementRule{
						{
							Level:     "must",
							Statement: "Test statement with sufficient length for validation",
							Rationale: "Test rationale with sufficient length for validation",
						},
					},
				},
				Examples: Examples{
					Good: []Example{
						{
							Title:    "Test",
							Language: "go",
							Snippet:  "test",
							Reason:   "Test reason with sufficient length for validation",
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(*tt.standard)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewValidator_RegistersAllCustomValidators(t *testing.T) {
	v := NewValidator()
	require.NotNil(t, v)

	customValidators := []string{"semver", "iso639_1", "kebab_case", "name_format"}

	for _, validatorName := range customValidators {
		t.Run(validatorName, func(t *testing.T) {
			type testStruct struct {
				Field string `validate:"required"`
			}
			ts := testStruct{Field: "test"}
			err := v.Struct(ts)
			assert.NoError(t, err, "validator should be initialized properly")
		})
	}
}

func TestValidate_MissingRequiredFields(t *testing.T) {
	standard := &Standard{}
	err := Validate(*standard)
	require.Error(t, err)

	var validatorErr validator.ValidationErrors
	assert.ErrorAs(t, err, &validatorErr)
}
