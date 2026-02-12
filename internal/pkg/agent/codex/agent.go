package codex

import (
	"fmt"
	"path"
	"strings"

	"github.com/creasty/defaults"
	"github.com/iancoleman/strcase"
	agentAPI "github.com/orbiqd/orbiqd-projectkit/pkg/agent"
	instructionAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/instruction"
	mcpAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/mcp"
	skillAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/skill"
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

func (agent *Agent) RebuildSkills(skillRepository skillAPI.Repository) error {
	skillsDir := path.Join(agent.options.ProjectSettingsDirName, agent.options.SkillsDirName)

	err := agent.rootFs.RemoveAll(skillsDir)
	if err != nil {
		return fmt.Errorf("skills directory removal: %w", err)
	}

	skills, err := skillRepository.GetAll()
	if err != nil {
		return fmt.Errorf("skills retrieval: %w", err)
	}

	if len(skills) == 0 {
		return nil
	}

	for _, skill := range skills {
		err := agent.renderSkill(skillsDir, skill)
		if err != nil {
			return err
		}
	}

	return nil
}

func (agent *Agent) RenderMCPServers(mcpServers []mcpAPI.MCPServer) error {
	return nil
}

func (agent *Agent) renderSkill(skillsDir string, skill skillAPI.Skill) error {
	skillDir := path.Join(skillsDir, string(skill.Metadata.Name))

	err := agent.rootFs.MkdirAll(skillDir, 0755)
	if err != nil {
		return fmt.Errorf("skill directory creation: %w", err)
	}

	skillFileContent := agent.renderSkillFile(skill)
	skillFilePath := path.Join(skillDir, "SKILL.md")

	err = afero.WriteFile(agent.rootFs, skillFilePath, []byte(skillFileContent), 0644)
	if err != nil {
		return fmt.Errorf("skill file write: %w", err)
	}

	if len(skill.Scripts) > 0 {
		scriptsDir := path.Join(skillDir, "scripts")

		err = agent.rootFs.MkdirAll(scriptsDir, 0755)
		if err != nil {
			return fmt.Errorf("scripts directory creation: %w", err)
		}

		for scriptName, script := range skill.Scripts {
			scriptPath := path.Join(scriptsDir, string(scriptName))

			err = afero.WriteFile(agent.rootFs, scriptPath, script.Content, 0755)
			if err != nil {
				return fmt.Errorf("script file write: %w", err)
			}
		}
	}

	return nil
}

func (agent *Agent) renderSkillFile(skill skillAPI.Skill) string {
	var builder strings.Builder

	builder.WriteString("---\n")
	_, _ = fmt.Fprintf(&builder, "name: %s\n", skill.Metadata.Name)
	_, _ = fmt.Fprintf(&builder, "description: %s\n", skill.Metadata.Description)
	builder.WriteString("---\n\n")
	builder.WriteString(skill.Instructions)

	return builder.String()
}

func (agent *Agent) GitIgnorePatterns() []string {
	return []string{
		agent.options.InstructionsFileName,
		agent.options.ProjectSettingsDirName,
	}
}

func (agent *Agent) GetKind() agentAPI.Kind {
	return Kind
}
