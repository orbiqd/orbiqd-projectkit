package codex

import (
	"errors"
	"fmt"

	"github.com/orbiqd/orbiqd-projectkit/internal/pkg/utils"
	agentAPI "github.com/orbiqd/orbiqd-projectkit/pkg/agent"
	"github.com/spf13/afero"
)

type Provider struct {
	rootFs afero.Fs
}

var _ agentAPI.Provider = (*Provider)(nil)

func NewProvider(rootFs afero.Fs) *Provider {
	return &Provider{
		rootFs: rootFs,
	}
}

func (provider *Provider) NewAgent(options any) (agentAPI.Agent, error) {
	agentOptions, err := utils.AnyToStruct[Options](options)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidOptions, err)
	}

	return NewAgent(*agentOptions, provider.rootFs), nil
}

func (provider *Provider) GetKind() agentAPI.Kind {
	return Kind
}

var (
	ErrInvalidOptions = errors.New("invalid options")
)
