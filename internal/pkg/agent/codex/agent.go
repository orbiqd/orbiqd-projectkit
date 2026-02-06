package codex

import (
	"fmt"
	"strings"

	"github.com/creasty/defaults"
	"github.com/iancoleman/strcase"
	agentAPI "github.com/orbiqd/orbiqd-projectkit/pkg/agent"
	instructionAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/instruction"
	"github.com/spf13/afero"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const Kind = "codex"

type Agent struct {
	options Options
	rootFs  afero.Fs
}

var _ agentAPI.Agent = (*Agent)(nil)

func NewAgent(options Options, rootFs afero.Fs) *Agent {
	defaults.MustSet(&options)

	return &Agent{
		options: options,
		rootFs:  rootFs,
	}
}

func (agent *Agent) RenderInstructions(instructions []instructionAPI.Instructions) error {
	var builder strings.Builder

	builder.WriteString("# Codex Agent Instructions\n\n")

	titleCaser := cases.Title(language.English)

	for _, instruction := range instructions {
		categoryWords := strcase.ToDelimited(string(instruction.Category), ' ')
		heading := titleCaser.String(categoryWords)

		_, _ = fmt.Fprintf(&builder, "## %s\n\n", heading)

		for _, rule := range instruction.Rules {
			_, _ = fmt.Fprintf(&builder, "- %s\n", rule)
		}

		builder.WriteString("\n")
	}

	err := afero.WriteFile(agent.rootFs, agent.options.InstructionsFileName, []byte(builder.String()), 0644)
	if err != nil {
		return fmt.Errorf("instructions file write: %w", err)
	}

	return nil
}
