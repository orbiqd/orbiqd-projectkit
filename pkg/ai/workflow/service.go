package workflow

type Service interface {
	Execute(workflowId WorkflowId) (ExecutionId, error)
}
