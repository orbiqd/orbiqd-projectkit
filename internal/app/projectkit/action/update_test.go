package action

import (
	"errors"
	"testing"

	aiAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai"
	instructionAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/instruction"
	skillAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/skill"
	workflowAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/workflow"
	docAPI "github.com/orbiqd/orbiqd-projectkit/pkg/doc"
	standardAPI "github.com/orbiqd/orbiqd-projectkit/pkg/doc/standard"
	projectAPI "github.com/orbiqd/orbiqd-projectkit/pkg/project"
	"github.com/orbiqd/orbiqd-projectkit/pkg/rulebook"
	sourceAPI "github.com/orbiqd/orbiqd-projectkit/pkg/source"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/mock"
)

func validInstructionFs(t *testing.T, category string, rules []string) afero.Fs {
	t.Helper()

	fs := afero.NewMemMapFs()

	content := "category: " + category + "\nrules:\n"
	for _, rule := range rules {
		content += "  - " + rule + "\n"
	}

	require.NoError(t, afero.WriteFile(fs, category+".yaml", []byte(content), 0644))

	return fs
}

func validSkillFs(t *testing.T, name, description, instructions string) afero.Fs {
	t.Helper()

	fs := afero.NewMemMapFs()

	require.NoError(t, fs.Mkdir(name, 0755))

	metadata := `name: ` + name + `
description: ` + description + `
`
	require.NoError(t, afero.WriteFile(fs, name+"/metadata.yaml", []byte(metadata), 0644))
	require.NoError(t, afero.WriteFile(fs, name+"/instructions.md", []byte(instructions), 0644))

	return fs
}

func validWorkflowFs(t *testing.T, id, name, description string) afero.Fs {
	t.Helper()

	fs := afero.NewMemMapFs()

	workflow := `metadata:
  id: ` + id + `
  name: ` + name + `
  description: ` + description + `
  version: 1.0.0
steps:
  - id: step1
    name: Step 1
    description: First step
    instructions:
      - Do something
`
	require.NoError(t, afero.WriteFile(fs, id+".yaml", []byte(workflow), 0644))

	return fs
}

func validStandardFsForUpdate(t *testing.T, name, version string) afero.Fs {
	t.Helper()

	fs := afero.NewMemMapFs()
	content := `metadata:
  name: ` + name + `
  version: ` + version + `
  tags:
    - test-tag
  scope:
    languages:
      - en
  relations:
    standard: []
specification:
  purpose: Test purpose for the standard
  goals:
    - First goal of the standard
requirements:
  rules:
    - level: must
      statement: This is a test requirement
      rationale: This is the rationale for the requirement
examples:
  good:
    - title: Good Example
      language: go
      snippet: |
        package main
        func main() {}
      reason: This is a good example because it follows the standard
`
	require.NoError(t, afero.WriteFile(fs, "standard.yaml", []byte(content), 0644))

	return fs
}

func configWithInstructions(uri string) projectAPI.Config {
	return projectAPI.Config{
		AI: &aiAPI.Config{
			Instruction: &instructionAPI.Config{
				Sources: []instructionAPI.SourceConfig{
					{URI: uri},
				},
			},
		},
	}
}

func configWithSkills(uri string) projectAPI.Config {
	return projectAPI.Config{
		AI: &aiAPI.Config{
			Skill: &skillAPI.Config{
				Sources: []skillAPI.SourceConfig{
					{URI: uri},
				},
			},
		},
	}
}

func configWithWorkflows(uri string) projectAPI.Config {
	return projectAPI.Config{
		AI: &aiAPI.Config{
			Workflows: &workflowAPI.Config{
				Sources: []workflowAPI.SourceConfig{
					{URI: uri},
				},
			},
		},
	}
}

func configWithStandards(uri string) projectAPI.Config {
	return projectAPI.Config{
		Docs: &docAPI.Config{
			Standard: &standardAPI.Config{
				Sources: []standardAPI.SourceConfig{
					{URI: uri},
				},
			},
		},
	}
}

func configWithRulebook(uri string) projectAPI.Config {
	return projectAPI.Config{
		Rulebook: &rulebook.Config{
			Sources: []rulebook.SourceConfig{
				{URI: uri},
			},
		},
	}
}

func TestUpdateActionRun_WhenEmptyConfig_ThenClearsAllRepos(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockWorkflowRepo := workflowAPI.NewMockRepository(t)
	mockStandardRepo := standardAPI.NewMockRepository(t)

	mockStandardRepo.EXPECT().RemoveAll().Return(nil)
	mockInstructionRepo.EXPECT().RemoveAll().Return(nil)
	mockSkillRepo.EXPECT().RemoveAll().Return(nil)
	mockWorkflowRepo.EXPECT().RemoveAllWorkflows().Return(nil)

	config := projectAPI.Config{}
	action := NewUpdateAction(config, mockResolver, mockInstructionRepo, mockSkillRepo, mockWorkflowRepo, mockStandardRepo)

	err := action.Run()

	require.NoError(t, err)
}

func TestUpdateActionRun_WhenInstructionsConfigured_ThenUpdatesInstructionRepo(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockWorkflowRepo := workflowAPI.NewMockRepository(t)
	mockStandardRepo := standardAPI.NewMockRepository(t)

	fs := validInstructionFs(t, "test-category", []string{"Rule one"})
	mockResolver.EXPECT().Resolve("file://./instructions").Return(fs, nil)

	mockStandardRepo.EXPECT().RemoveAll().Return(nil)
	mockInstructionRepo.EXPECT().RemoveAll().Return(nil)
	mockInstructionRepo.EXPECT().AddInstructions(mock.AnythingOfType("instruction.Instructions")).Return(nil)
	mockSkillRepo.EXPECT().RemoveAll().Return(nil)
	mockWorkflowRepo.EXPECT().RemoveAllWorkflows().Return(nil)

	config := configWithInstructions("file://./instructions")
	action := NewUpdateAction(config, mockResolver, mockInstructionRepo, mockSkillRepo, mockWorkflowRepo, mockStandardRepo)

	err := action.Run()

	require.NoError(t, err)
}

func TestUpdateActionRun_WhenSkillsConfigured_ThenUpdatesSkillRepo(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockWorkflowRepo := workflowAPI.NewMockRepository(t)
	mockStandardRepo := standardAPI.NewMockRepository(t)

	fs := validSkillFs(t, "test-skill", "Test skill", "Test instructions")
	mockResolver.EXPECT().Resolve("file://./skills").Return(fs, nil)

	mockStandardRepo.EXPECT().RemoveAll().Return(nil)
	mockInstructionRepo.EXPECT().RemoveAll().Return(nil)
	mockSkillRepo.EXPECT().RemoveAll().Return(nil)
	mockSkillRepo.EXPECT().AddSkill(mock.AnythingOfType("skill.Skill")).Return(nil)
	mockWorkflowRepo.EXPECT().RemoveAllWorkflows().Return(nil)

	config := configWithSkills("file://./skills")
	action := NewUpdateAction(config, mockResolver, mockInstructionRepo, mockSkillRepo, mockWorkflowRepo, mockStandardRepo)

	err := action.Run()

	require.NoError(t, err)
}

func TestUpdateActionRun_WhenWorkflowsConfigured_ThenUpdatesWorkflowRepo(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockWorkflowRepo := workflowAPI.NewMockRepository(t)
	mockStandardRepo := standardAPI.NewMockRepository(t)

	fs := validWorkflowFs(t, "test-workflow", "Test Workflow", "Test description")
	mockResolver.EXPECT().Resolve("file://./workflows").Return(fs, nil)

	mockStandardRepo.EXPECT().RemoveAll().Return(nil)
	mockInstructionRepo.EXPECT().RemoveAll().Return(nil)
	mockSkillRepo.EXPECT().RemoveAll().Return(nil)
	mockWorkflowRepo.EXPECT().RemoveAllWorkflows().Return(nil)
	mockWorkflowRepo.EXPECT().AddWorkflow(mock.AnythingOfType("workflow.Workflow")).Return(nil)

	config := configWithWorkflows("file://./workflows")
	action := NewUpdateAction(config, mockResolver, mockInstructionRepo, mockSkillRepo, mockWorkflowRepo, mockStandardRepo)

	err := action.Run()

	require.NoError(t, err)
}

func TestUpdateActionRun_WhenStandardsConfigured_ThenUpdatesStandardRepo(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockWorkflowRepo := workflowAPI.NewMockRepository(t)
	mockStandardRepo := standardAPI.NewMockRepository(t)

	fs := validStandardFsForUpdate(t, "Test Standard", "1.0.0")
	mockResolver.EXPECT().Resolve("file://./standards").Return(fs, nil)

	mockStandardRepo.EXPECT().RemoveAll().Return(nil)
	mockStandardRepo.EXPECT().AddStandard(mock.AnythingOfType("standard.Standard")).Return(nil)
	mockInstructionRepo.EXPECT().RemoveAll().Return(nil)
	mockSkillRepo.EXPECT().RemoveAll().Return(nil)
	mockWorkflowRepo.EXPECT().RemoveAllWorkflows().Return(nil)

	config := configWithStandards("file://./standards")
	action := NewUpdateAction(config, mockResolver, mockInstructionRepo, mockSkillRepo, mockWorkflowRepo, mockStandardRepo)

	err := action.Run()

	require.NoError(t, err)
}

func TestUpdateActionRun_WhenRulebookConfigured_ThenMergesAllContent(t *testing.T) {
	t.Skip("Rulebook loader integration is complex - covered by loader tests")
}

func TestUpdateActionRun_WhenLoadInstructionsFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockWorkflowRepo := workflowAPI.NewMockRepository(t)
	mockStandardRepo := standardAPI.NewMockRepository(t)

	loadErr := errors.New("load instructions error")
	mockResolver.EXPECT().Resolve("file://./instructions").Return(nil, loadErr)

	config := configWithInstructions("file://./instructions")
	action := NewUpdateAction(config, mockResolver, mockInstructionRepo, mockSkillRepo, mockWorkflowRepo, mockStandardRepo)

	err := action.Run()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "load instructions from config")
	assert.ErrorIs(t, err, loadErr)
}

func TestUpdateActionRun_WhenLoadSkillsFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockWorkflowRepo := workflowAPI.NewMockRepository(t)
	mockStandardRepo := standardAPI.NewMockRepository(t)

	loadErr := errors.New("load skills error")
	mockResolver.EXPECT().Resolve("file://./skills").Return(nil, loadErr)

	config := configWithSkills("file://./skills")
	action := NewUpdateAction(config, mockResolver, mockInstructionRepo, mockSkillRepo, mockWorkflowRepo, mockStandardRepo)

	err := action.Run()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "load skills from config")
	assert.ErrorIs(t, err, loadErr)
}

func TestUpdateActionRun_WhenLoadWorkflowsFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockWorkflowRepo := workflowAPI.NewMockRepository(t)
	mockStandardRepo := standardAPI.NewMockRepository(t)

	loadErr := errors.New("load workflows error")
	mockResolver.EXPECT().Resolve("file://./workflows").Return(nil, loadErr)

	config := configWithWorkflows("file://./workflows")
	action := NewUpdateAction(config, mockResolver, mockInstructionRepo, mockSkillRepo, mockWorkflowRepo, mockStandardRepo)

	err := action.Run()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "load workflows from config")
	assert.ErrorIs(t, err, loadErr)
}

func TestUpdateActionRun_WhenLoadStandardsFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockWorkflowRepo := workflowAPI.NewMockRepository(t)
	mockStandardRepo := standardAPI.NewMockRepository(t)

	loadErr := errors.New("load standards error")
	mockResolver.EXPECT().Resolve("file://./standards").Return(nil, loadErr)

	config := configWithStandards("file://./standards")
	action := NewUpdateAction(config, mockResolver, mockInstructionRepo, mockSkillRepo, mockWorkflowRepo, mockStandardRepo)

	err := action.Run()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "load standards from config")
	assert.ErrorIs(t, err, loadErr)
}

func TestUpdateActionRun_WhenLoadRulebooksFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockWorkflowRepo := workflowAPI.NewMockRepository(t)
	mockStandardRepo := standardAPI.NewMockRepository(t)

	loadErr := errors.New("load rulebooks error")
	mockResolver.EXPECT().Resolve("file://./rulebook").Return(nil, loadErr)

	config := configWithRulebook("file://./rulebook")
	action := NewUpdateAction(config, mockResolver, mockInstructionRepo, mockSkillRepo, mockWorkflowRepo, mockStandardRepo)

	err := action.Run()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "load rule books from config")
	assert.ErrorIs(t, err, loadErr)
}

func TestUpdateActionRun_WhenStandardRemoveAllFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockWorkflowRepo := workflowAPI.NewMockRepository(t)
	mockStandardRepo := standardAPI.NewMockRepository(t)

	removeErr := errors.New("remove all standards error")
	mockStandardRepo.EXPECT().RemoveAll().Return(removeErr)

	config := projectAPI.Config{}
	action := NewUpdateAction(config, mockResolver, mockInstructionRepo, mockSkillRepo, mockWorkflowRepo, mockStandardRepo)

	err := action.Run()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "remove all standards from repository")
	assert.ErrorIs(t, err, removeErr)
}

func TestUpdateActionRun_WhenStandardAddFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockWorkflowRepo := workflowAPI.NewMockRepository(t)
	mockStandardRepo := standardAPI.NewMockRepository(t)

	fs := validStandardFsForUpdate(t, "Test Standard", "1.0.0")
	mockResolver.EXPECT().Resolve("file://./standards").Return(fs, nil)

	addErr := errors.New("add standard error")
	mockStandardRepo.EXPECT().RemoveAll().Return(nil)
	mockStandardRepo.EXPECT().AddStandard(mock.AnythingOfType("standard.Standard")).Return(addErr)

	config := configWithStandards("file://./standards")
	action := NewUpdateAction(config, mockResolver, mockInstructionRepo, mockSkillRepo, mockWorkflowRepo, mockStandardRepo)

	err := action.Run()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "add standard to repository")
	assert.ErrorIs(t, err, addErr)
}

func TestUpdateActionRun_WhenInstructionRemoveAllFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockWorkflowRepo := workflowAPI.NewMockRepository(t)
	mockStandardRepo := standardAPI.NewMockRepository(t)

	removeErr := errors.New("remove all instructions error")
	mockStandardRepo.EXPECT().RemoveAll().Return(nil)
	mockInstructionRepo.EXPECT().RemoveAll().Return(removeErr)

	config := projectAPI.Config{}
	action := NewUpdateAction(config, mockResolver, mockInstructionRepo, mockSkillRepo, mockWorkflowRepo, mockStandardRepo)

	err := action.Run()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "remove all instructions from repository")
	assert.ErrorIs(t, err, removeErr)
}

func TestUpdateActionRun_WhenInstructionAddFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockWorkflowRepo := workflowAPI.NewMockRepository(t)
	mockStandardRepo := standardAPI.NewMockRepository(t)

	fs := validInstructionFs(t, "test-category", []string{"Rule one"})
	mockResolver.EXPECT().Resolve("file://./instructions").Return(fs, nil)

	addErr := errors.New("add instructions error")
	mockStandardRepo.EXPECT().RemoveAll().Return(nil)
	mockInstructionRepo.EXPECT().RemoveAll().Return(nil)
	mockInstructionRepo.EXPECT().AddInstructions(mock.AnythingOfType("instruction.Instructions")).Return(addErr)

	config := configWithInstructions("file://./instructions")
	action := NewUpdateAction(config, mockResolver, mockInstructionRepo, mockSkillRepo, mockWorkflowRepo, mockStandardRepo)

	err := action.Run()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "add instructions to repository")
	assert.ErrorIs(t, err, addErr)
}

func TestUpdateActionRun_WhenSkillRemoveAllFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockWorkflowRepo := workflowAPI.NewMockRepository(t)
	mockStandardRepo := standardAPI.NewMockRepository(t)

	removeErr := errors.New("remove all skills error")
	mockStandardRepo.EXPECT().RemoveAll().Return(nil)
	mockInstructionRepo.EXPECT().RemoveAll().Return(nil)
	mockSkillRepo.EXPECT().RemoveAll().Return(removeErr)

	config := projectAPI.Config{}
	action := NewUpdateAction(config, mockResolver, mockInstructionRepo, mockSkillRepo, mockWorkflowRepo, mockStandardRepo)

	err := action.Run()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "remove all skills from repository")
	assert.ErrorIs(t, err, removeErr)
}

func TestUpdateActionRun_WhenSkillAddFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockWorkflowRepo := workflowAPI.NewMockRepository(t)
	mockStandardRepo := standardAPI.NewMockRepository(t)

	fs := validSkillFs(t, "test-skill", "Test skill", "Test instructions")
	mockResolver.EXPECT().Resolve("file://./skills").Return(fs, nil)

	addErr := errors.New("add skill error")
	mockStandardRepo.EXPECT().RemoveAll().Return(nil)
	mockInstructionRepo.EXPECT().RemoveAll().Return(nil)
	mockSkillRepo.EXPECT().RemoveAll().Return(nil)
	mockSkillRepo.EXPECT().AddSkill(mock.AnythingOfType("skill.Skill")).Return(addErr)

	config := configWithSkills("file://./skills")
	action := NewUpdateAction(config, mockResolver, mockInstructionRepo, mockSkillRepo, mockWorkflowRepo, mockStandardRepo)

	err := action.Run()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "add skill")
	assert.ErrorIs(t, err, addErr)
}

func TestUpdateActionRun_WhenWorkflowRemoveAllFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockWorkflowRepo := workflowAPI.NewMockRepository(t)
	mockStandardRepo := standardAPI.NewMockRepository(t)

	removeErr := errors.New("remove all workflows error")
	mockStandardRepo.EXPECT().RemoveAll().Return(nil)
	mockInstructionRepo.EXPECT().RemoveAll().Return(nil)
	mockSkillRepo.EXPECT().RemoveAll().Return(nil)
	mockWorkflowRepo.EXPECT().RemoveAllWorkflows().Return(removeErr)

	config := projectAPI.Config{}
	action := NewUpdateAction(config, mockResolver, mockInstructionRepo, mockSkillRepo, mockWorkflowRepo, mockStandardRepo)

	err := action.Run()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "remove all workflows from repository")
	assert.ErrorIs(t, err, removeErr)
}

func TestUpdateActionRun_WhenWorkflowAddFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	mockResolver := sourceAPI.NewMockResolver(t)
	mockInstructionRepo := instructionAPI.NewMockRepository(t)
	mockSkillRepo := skillAPI.NewMockRepository(t)
	mockWorkflowRepo := workflowAPI.NewMockRepository(t)
	mockStandardRepo := standardAPI.NewMockRepository(t)

	fs := validWorkflowFs(t, "test-workflow", "Test Workflow", "Test description")
	mockResolver.EXPECT().Resolve("file://./workflows").Return(fs, nil)

	addErr := errors.New("add workflow error")
	mockStandardRepo.EXPECT().RemoveAll().Return(nil)
	mockInstructionRepo.EXPECT().RemoveAll().Return(nil)
	mockSkillRepo.EXPECT().RemoveAll().Return(nil)
	mockWorkflowRepo.EXPECT().RemoveAllWorkflows().Return(nil)
	mockWorkflowRepo.EXPECT().AddWorkflow(mock.AnythingOfType("workflow.Workflow")).Return(addErr)

	config := configWithWorkflows("file://./workflows")
	action := NewUpdateAction(config, mockResolver, mockInstructionRepo, mockSkillRepo, mockWorkflowRepo, mockStandardRepo)

	err := action.Run()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "add workflow")
	assert.ErrorIs(t, err, addErr)
}
