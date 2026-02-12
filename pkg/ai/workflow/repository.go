package workflow

import (
	"errors"
	"regexp"
)

var idPattern = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)

func (id WorkflowId) Validate() error {
	if !idPattern.MatchString(string(id)) {
		return ErrWorkflowInvalidID
	}
	return nil
}

func (id ExecutionId) Validate() error {
	if !idPattern.MatchString(string(id)) {
		return ErrExecutionInvalidID
	}
	return nil
}

type Repository interface {
	AddWorkflow(workflow Workflow) error
	GetAllWorkflows() ([]Workflow, error)
	GetWorkflowById(id WorkflowId) (*Workflow, error)
	RemoveAllWorkflows() error

	AddExecution(execution Execution) error
	UpdateExecution(execution Execution) error
	GetExecutionById(id ExecutionId) (*Execution, error)
}

var (
	ErrWorkflowNotFound      = errors.New("workflow not found")
	ErrWorkflowAlreadyExists = errors.New("workflow already exists")
	ErrWorkflowInvalidID     = errors.New("workflow id must be alphanumeric with dashes")

	ErrExecutionNotFound      = errors.New("execution not found")
	ErrExecutionAlreadyExists = errors.New("execution already exists")
	ErrExecutionInvalidID     = errors.New("execution id must be alphanumeric with dashes")
)
