package workflow

import "github.com/invopop/jsonschema"

type WorkflowId string
type ExecutionId string

type Metadata struct {
	ID          WorkflowId `json:"id" validate:"required"`
	Name        string     `json:"name" validate:"required"`
	Description string     `json:"description" validate:"required"`
	Version     string     `json:"version" validate:"required,semver"`
}

type Step struct {
	ID           string   `json:"id" validate:"required"`
	Name         string   `json:"name" validate:"required"`
	Description  string   `json:"description" validate:"required"`
	Instructions []string `json:"instructions" validate:"required,min=1"`
}

type Workflow struct {
	Metadata Metadata                      `json:"metadata" validate:"required"`
	State    map[string]*jsonschema.Schema `json:"state,omitempty" validate:"omitempty"`
	Steps    []Step                        `json:"steps" validate:"required,min=1,dive"`
}

type Execution struct {
	Id         ExecutionId `json:"id" validate:"required"`
	WorkflowId WorkflowId  `json:"workflowId" validate:"required"`

	StateValues map[string]any `json:"stateValues" validate:"required"`
	StepId      string         `json:"stepId" validate:"required"`
}
