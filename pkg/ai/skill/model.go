package skill

type Name string

type Metadata struct {
	Name        Name   `json:"name" validate:"required"`
	Description string `json:"description" validate:"required,max=256"`
}

type ScriptName string

type Script struct {
	ContentType string `json:"contentType"`
	Content     []byte `json:"content"`
}

type Skill struct {
	Metadata     Metadata              `json:"metadata" validate:"required"`
	Instructions string                `json:"instructions" validate:"required"`
	Scripts      map[ScriptName]Script `json:"scripts,omitempty" validate:"omitempty,dive"`
}
