package standard

type StandardId string

type ScopeMetadata struct {
	Languages       []string `json:"languages" validate:"required,min=1,dive,iso639_1"`
	AppliesTo       []string `json:"appliesTo,omitempty" validate:"omitempty,dive,min=1,max=100"`
	NotApplicableTo []string `json:"notApplicableTo,omitempty" validate:"omitempty,dive,min=1,max=100"`
}

type RelationMetadata struct {
	Standard []string `json:"standard,omitempty" validate:"omitempty,dive,url"`
}

type Metadata struct {
	Id      StandardId       `json:"id" validate:"required,kebab_case,min=1,max=100"`
	Name    string           `json:"name" validate:"required,name_format,min=1,max=200"`
	Version string           `json:"version" validate:"required,semver"`
	Tags    []string         `json:"tags" validate:"required,min=1,dive,kebab_case,min=1,max=50"`
	Scope   ScopeMetadata    `json:"scope" validate:"required"`
	Related RelationMetadata `json:"relations" validate:"required"`
}

type Specification struct {
	Purpose  string   `json:"purpose" validate:"required,min=10,max=500"`
	Goals    []string `json:"goals" validate:"required,min=1,dive,min=10,max=500"`
	NonGoals []string `json:"nonGoals,omitempty" validate:"omitempty,dive,min=10,max=500"`
}

type FieldDefinition struct {
	FieldName string `json:"fieldName" validate:"required,name_format,min=1,max=200"`
}

type TermDefinition struct {
	Abbreviation string `json:"abbreviation" validate:"required,min=1,max=50"`
	Term         string `json:"term" validate:"required,name_format,min=1,max=200"`
	Meaning      string `json:"meaning" validate:"required,min=10,max=500"`
}

type Definitions struct {
	Fields []FieldDefinition `json:"fields,omitempty"`
	Terms  []TermDefinition  `json:"terms,omitempty"`
}

type RequirementVerificationMethod struct {
	Type string `json:"type" validate:"required,min=10,max=500"`
	Hint string `json:"hint" validate:"required,min=10,max=500"`
}

type RequirementException struct {
	When string `json:"when" validate:"required,min=10,max=500"`
}

type RequirementRule struct {
	Level              string                          `json:"level" validate:"required,oneof=must should may recommended optional"`
	Statement          string                          `json:"statement" validate:"required,min=10,max=500"`
	Rationale          string                          `json:"rationale" validate:"required,min=10,max=500"`
	Exceptions         []RequirementException          `json:"exceptions,omitempty" validate:"omitempty,dive"`
	VerificationMethod []RequirementVerificationMethod `json:"verificationMethod,omitempty" validate:"omitempty,dive"`
}

type Requirements struct {
	Rules []RequirementRule `json:"rules" validate:"required,min=1,dive"`
}

type GoldenPathExampleFile struct {
	Path    string `json:"path" validate:"required,min=1,max=500"`
	Snippet string `json:"snippet" validate:"required,min=1"`
}

type GoldenPathExample struct {
	Name     string                  `json:"name" validate:"required,name_format,min=1,max=200"`
	When     []string                `json:"when,omitempty" validate:"omitempty,dive,min=10,max=500"`
	Steps    []string                `json:"steps" validate:"required,min=1,dive,min=10,max=500"`
	Examples []GoldenPathExampleFile `json:"examples,omitempty" validate:"omitempty,dive"`
}

type GoldenPath struct {
	Steps    []string            `json:"steps" validate:"required,min=1,dive,min=10,max=500"`
	Examples []GoldenPathExample `json:"examples,omitempty" validate:"omitempty,dive"`
}

type Reference struct {
	Title string `json:"title" validate:"required,min=1,max=200"`
	Type  string `json:"type" validate:"required,min=1,max=50"`
	URI   string `json:"uri" validate:"required,url"`
}

type Example struct {
	Title    string `json:"title" validate:"required,min=1,max=200"`
	Language string `json:"language" validate:"required,min=2,max=10"`
	Snippet  string `json:"snippet" validate:"required,min=1"`
	Reason   string `json:"reason" validate:"required,min=10,max=500"`
}

type Examples struct {
	Good []Example `json:"good" validate:"required,min=1,dive"`
	Bad  []Example `json:"bad,omitempty" validate:"omitempty,dive"`
}

type Standard struct {
	Metadata      Metadata       `json:"metadata" validate:"required"`
	Specification Specification  `json:"specification" validate:"required"`
	Definitions   *Definitions   `json:"definitions,omitempty" validate:"omitempty"`
	Requirements  Requirements   `json:"requirements" validate:"required"`
	GoldenPath    *GoldenPath    `json:"goldenPath,omitempty" validate:"omitempty"`
	Examples      Examples       `json:"examples" validate:"required"`
	References    []Reference    `json:"references,omitempty" validate:"omitempty,dive"`
}
