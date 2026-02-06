package agent

type Provider interface {
	NewAgent(options any) (Agent, error)
	GetKind() Kind
}
