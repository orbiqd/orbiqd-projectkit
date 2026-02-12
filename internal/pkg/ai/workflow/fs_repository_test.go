package workflow

import (
	"encoding/json"
	"errors"
	"io/fs"
	"testing"

	workflowAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/workflow"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.nhat.io/aferomock"
)

func TestFsRepository_AddWorkflow_WhenNewWorkflow_ThenStoresSuccessfully(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs, afero.NewMemMapFs())
	workflow := workflowAPI.Workflow{
		Metadata: workflowAPI.Metadata{
			ID:          workflowAPI.WorkflowId("test-workflow"),
			Name:        "Test Workflow",
			Description: "Test workflow description",
			Version:     "1.0.0",
		},
		Steps: []workflowAPI.Step{
			{
				ID:           "step1",
				Name:         "Step 1",
				Description:  "First step",
				Instructions: []string{"Do something"},
			},
		},
	}

	err := repo.AddWorkflow(workflow)
	require.NoError(t, err)

	result, err := repo.GetWorkflowById(workflowAPI.WorkflowId("test-workflow"))
	require.NoError(t, err)
	assert.Equal(t, workflow.Metadata.ID, result.Metadata.ID)
	assert.Equal(t, workflow.Metadata.Name, result.Metadata.Name)
	assert.Equal(t, workflow.Metadata.Description, result.Metadata.Description)
}

func TestFsRepository_AddWorkflow_WhenDuplicateID_ThenReturnsAlreadyExistsError(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs, afero.NewMemMapFs())
	workflow := workflowAPI.Workflow{
		Metadata: workflowAPI.Metadata{
			ID:          workflowAPI.WorkflowId("duplicate-workflow"),
			Name:        "First Workflow",
			Description: "First workflow",
			Version:     "1.0.0",
		},
		Steps: []workflowAPI.Step{
			{
				ID:           "step1",
				Name:         "Step 1",
				Description:  "First step",
				Instructions: []string{"Do something"},
			},
		},
	}

	err := repo.AddWorkflow(workflow)
	require.NoError(t, err)

	duplicateWorkflow := workflowAPI.Workflow{
		Metadata: workflowAPI.Metadata{
			ID:          workflowAPI.WorkflowId("duplicate-workflow"),
			Name:        "Second Workflow",
			Description: "Second workflow",
			Version:     "2.0.0",
		},
		Steps: []workflowAPI.Step{
			{
				ID:           "step2",
				Name:         "Step 2",
				Description:  "Second step",
				Instructions: []string{"Do something else"},
			},
		},
	}

	err = repo.AddWorkflow(duplicateWorkflow)
	require.ErrorIs(t, err, workflowAPI.ErrWorkflowAlreadyExists)
}

func TestFsRepository_AddWorkflow_WhenCalled_ThenCreatesFileWithIDBasedName(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs, afero.NewMemMapFs())
	workflow := workflowAPI.Workflow{
		Metadata: workflowAPI.Metadata{
			ID:          workflowAPI.WorkflowId("my-workflow-id"),
			Name:        "Test Workflow",
			Description: "Test workflow",
			Version:     "1.0.0",
		},
		Steps: []workflowAPI.Step{
			{
				ID:           "step1",
				Name:         "Step 1",
				Description:  "First step",
				Instructions: []string{"Do something"},
			},
		},
	}

	err := repo.AddWorkflow(workflow)
	require.NoError(t, err)

	files, err := afero.ReadDir(fs, ".")
	require.NoError(t, err)
	require.Len(t, files, 1)

	filename := files[0].Name()
	assert.Equal(t, "my-workflow-id.json", filename)
}

func TestFsRepository_AddWorkflow_WhenCalled_ThenStoresValidJson(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs, afero.NewMemMapFs())
	workflow := workflowAPI.Workflow{
		Metadata: workflowAPI.Metadata{
			ID:          workflowAPI.WorkflowId("test-workflow"),
			Name:        "Test Workflow",
			Description: "Test workflow",
			Version:     "1.0.0",
		},
		Steps: []workflowAPI.Step{
			{
				ID:           "step1",
				Name:         "Step 1",
				Description:  "First step",
				Instructions: []string{"Do something"},
			},
		},
	}

	err := repo.AddWorkflow(workflow)
	require.NoError(t, err)

	content, err := afero.ReadFile(fs, "test-workflow.json")
	require.NoError(t, err)

	var stored workflowAPI.Workflow
	err = json.Unmarshal(content, &stored)
	require.NoError(t, err)
	assert.Equal(t, workflowAPI.WorkflowId("test-workflow"), stored.Metadata.ID)
	assert.Equal(t, "Test Workflow", stored.Metadata.Name)
	assert.Equal(t, "Test workflow", stored.Metadata.Description)
}

func TestFsRepository_AddWorkflow_WhenInvalidID_ThenReturnsInvalidIDError(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs, afero.NewMemMapFs())
	workflow := workflowAPI.Workflow{
		Metadata: workflowAPI.Metadata{
			ID:          workflowAPI.WorkflowId("invalid/id"),
			Name:        "Test Workflow",
			Description: "Test workflow",
			Version:     "1.0.0",
		},
		Steps: []workflowAPI.Step{
			{
				ID:           "step1",
				Name:         "Step 1",
				Description:  "First step",
				Instructions: []string{"Do something"},
			},
		},
	}

	err := repo.AddWorkflow(workflow)
	require.ErrorIs(t, err, workflowAPI.ErrWorkflowInvalidID)
}

func TestFsRepository_AddWorkflow_WhenExistsFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	existsErr := errors.New("exists check failed")
	base := afero.NewMemMapFs()
	fs := aferomock.OverrideFs(base, aferomock.FsCallbacks{
		StatFunc: func(name string) (fs.FileInfo, error) {
			return nil, existsErr
		},
	})
	repo := NewFsRepository(fs, afero.NewMemMapFs())
	workflow := workflowAPI.Workflow{
		Metadata: workflowAPI.Metadata{
			ID:          workflowAPI.WorkflowId("test-workflow"),
			Name:        "Test Workflow",
			Description: "Test workflow",
			Version:     "1.0.0",
		},
		Steps: []workflowAPI.Step{
			{
				ID:           "step1",
				Name:         "Step 1",
				Description:  "First step",
				Instructions: []string{"Do something"},
			},
		},
	}

	err := repo.AddWorkflow(workflow)

	require.ErrorIs(t, err, existsErr)
}

func TestFsRepository_AddWorkflow_WhenWriteFileFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	writeFileErr := errors.New("write file failed")
	base := afero.NewMemMapFs()
	fs := aferomock.OverrideFs(base, aferomock.FsCallbacks{
		OpenFileFunc: func(name string, flag int, perm fs.FileMode) (afero.File, error) {
			return nil, writeFileErr
		},
	})
	repo := NewFsRepository(fs, afero.NewMemMapFs())
	workflow := workflowAPI.Workflow{
		Metadata: workflowAPI.Metadata{
			ID:          workflowAPI.WorkflowId("test-workflow"),
			Name:        "Test Workflow",
			Description: "Test workflow",
			Version:     "1.0.0",
		},
		Steps: []workflowAPI.Step{
			{
				ID:           "step1",
				Name:         "Step 1",
				Description:  "First step",
				Instructions: []string{"Do something"},
			},
		},
	}

	err := repo.AddWorkflow(workflow)

	require.ErrorIs(t, err, writeFileErr)
}

func TestFsRepository_GetAll_WhenEmpty_ThenReturnsEmptySlice(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs, afero.NewMemMapFs())

	workflows, err := repo.GetAllWorkflows()
	require.NoError(t, err)
	require.NotNil(t, workflows)
	assert.Empty(t, workflows)
}

func TestFsRepository_GetAll_WhenSingleWorkflow_ThenReturnsWorkflow(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs, afero.NewMemMapFs())
	workflow := workflowAPI.Workflow{
		Metadata: workflowAPI.Metadata{
			ID:          workflowAPI.WorkflowId("test-workflow"),
			Name:        "Test Workflow",
			Description: "Test description",
			Version:     "1.0.0",
		},
		Steps: []workflowAPI.Step{
			{
				ID:           "step1",
				Name:         "Step 1",
				Description:  "First step",
				Instructions: []string{"Do something"},
			},
		},
	}

	err := repo.AddWorkflow(workflow)
	require.NoError(t, err)

	workflows, err := repo.GetAllWorkflows()
	require.NoError(t, err)
	require.Len(t, workflows, 1)
	assert.Equal(t, workflow.Metadata.ID, workflows[0].Metadata.ID)
	assert.Equal(t, workflow.Metadata.Name, workflows[0].Metadata.Name)
}

func TestFsRepository_GetAll_WhenMultipleWorkflows_ThenReturnsSortedByName(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs, afero.NewMemMapFs())

	workflow1 := workflowAPI.Workflow{
		Metadata: workflowAPI.Metadata{
			ID:          workflowAPI.WorkflowId("workflow-z"),
			Name:        "Zebra Workflow",
			Description: "Last alphabetically",
			Version:     "1.0.0",
		},
		Steps: []workflowAPI.Step{
			{
				ID:           "step1",
				Name:         "Step 1",
				Description:  "First step",
				Instructions: []string{"Do something"},
			},
		},
	}

	workflow2 := workflowAPI.Workflow{
		Metadata: workflowAPI.Metadata{
			ID:          workflowAPI.WorkflowId("workflow-a"),
			Name:        "Alpha Workflow",
			Description: "First alphabetically",
			Version:     "1.0.0",
		},
		Steps: []workflowAPI.Step{
			{
				ID:           "step2",
				Name:         "Step 2",
				Description:  "Second step",
				Instructions: []string{"Do something else"},
			},
		},
	}

	workflow3 := workflowAPI.Workflow{
		Metadata: workflowAPI.Metadata{
			ID:          workflowAPI.WorkflowId("workflow-m"),
			Name:        "Middle Workflow",
			Description: "Middle alphabetically",
			Version:     "1.0.0",
		},
		Steps: []workflowAPI.Step{
			{
				ID:           "step3",
				Name:         "Step 3",
				Description:  "Third step",
				Instructions: []string{"Do another thing"},
			},
		},
	}

	err := repo.AddWorkflow(workflow1)
	require.NoError(t, err)
	err = repo.AddWorkflow(workflow2)
	require.NoError(t, err)
	err = repo.AddWorkflow(workflow3)
	require.NoError(t, err)

	workflows, err := repo.GetAllWorkflows()
	require.NoError(t, err)
	require.Len(t, workflows, 3)

	assert.Equal(t, "Alpha Workflow", workflows[0].Metadata.Name)
	assert.Equal(t, "Middle Workflow", workflows[1].Metadata.Name)
	assert.Equal(t, "Zebra Workflow", workflows[2].Metadata.Name)
}

func TestFsRepository_GetAll_WhenReadDirFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	readDirErr := errors.New("read dir failed")
	base := afero.NewMemMapFs()
	fs := aferomock.OverrideFs(base, aferomock.FsCallbacks{
		OpenFunc: func(name string) (afero.File, error) {
			if name == "." {
				return nil, readDirErr
			}
			return base.Open(name)
		},
	})
	repo := NewFsRepository(fs, afero.NewMemMapFs())

	result, err := repo.GetAllWorkflows()

	require.ErrorIs(t, err, readDirErr)
	assert.Nil(t, result)
}

func TestFsRepository_GetAll_WhenReadFileFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	base := afero.NewMemMapFs()
	workflow := workflowAPI.Workflow{
		Metadata: workflowAPI.Metadata{
			ID:          workflowAPI.WorkflowId("test-workflow"),
			Name:        "Test Workflow",
			Description: "Test workflow",
			Version:     "1.0.0",
		},
		Steps: []workflowAPI.Step{
			{
				ID:           "step1",
				Name:         "Step 1",
				Description:  "First step",
				Instructions: []string{"Do something"},
			},
		},
	}
	data, err := json.Marshal(workflow)
	require.NoError(t, err)
	err = afero.WriteFile(base, "test-workflow.json", data, 0644)
	require.NoError(t, err)

	readFileErr := errors.New("read file failed")
	fs := aferomock.OverrideFs(base, aferomock.FsCallbacks{
		OpenFunc: func(name string) (afero.File, error) {
			if name == "test-workflow.json" {
				return nil, readFileErr
			}
			return base.Open(name)
		},
	})
	repo := NewFsRepository(fs, afero.NewMemMapFs())

	result, err := repo.GetAllWorkflows()

	require.ErrorIs(t, err, readFileErr)
	assert.Nil(t, result)
}

func TestFsRepository_GetWorkflowById_WhenWorkflowExists_ThenReturnsWorkflow(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs, afero.NewMemMapFs())
	workflow := workflowAPI.Workflow{
		Metadata: workflowAPI.Metadata{
			ID:          workflowAPI.WorkflowId("existing-workflow"),
			Name:        "Existing Workflow",
			Description: "Existing workflow description",
			Version:     "1.0.0",
		},
		Steps: []workflowAPI.Step{
			{
				ID:           "step1",
				Name:         "Step 1",
				Description:  "First step",
				Instructions: []string{"Do something"},
			},
		},
	}

	err := repo.AddWorkflow(workflow)
	require.NoError(t, err)

	result, err := repo.GetWorkflowById(workflowAPI.WorkflowId("existing-workflow"))
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, workflowAPI.WorkflowId("existing-workflow"), result.Metadata.ID)
	assert.Equal(t, "Existing Workflow", result.Metadata.Name)
	assert.Equal(t, "Existing workflow description", result.Metadata.Description)
}

func TestFsRepository_GetWorkflowById_WhenWorkflowNotFound_ThenReturnsError(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs, afero.NewMemMapFs())
	workflow := workflowAPI.Workflow{
		Metadata: workflowAPI.Metadata{
			ID:          workflowAPI.WorkflowId("some-workflow"),
			Name:        "Some Workflow",
			Description: "Some workflow",
			Version:     "1.0.0",
		},
		Steps: []workflowAPI.Step{
			{
				ID:           "step1",
				Name:         "Step 1",
				Description:  "First step",
				Instructions: []string{"Do something"},
			},
		},
	}

	err := repo.AddWorkflow(workflow)
	require.NoError(t, err)

	result, err := repo.GetWorkflowById(workflowAPI.WorkflowId("non-existing-workflow"))
	require.ErrorIs(t, err, workflowAPI.ErrWorkflowNotFound)
	assert.Nil(t, result)
}

func TestFsRepository_GetWorkflowById_WhenEmptyRepository_ThenReturnsError(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs, afero.NewMemMapFs())

	result, err := repo.GetWorkflowById(workflowAPI.WorkflowId("any-workflow"))
	require.ErrorIs(t, err, workflowAPI.ErrWorkflowNotFound)
	assert.Nil(t, result)
}

func TestFsRepository_GetWorkflowById_WhenInvalidID_ThenReturnsInvalidIDError(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs, afero.NewMemMapFs())

	result, err := repo.GetWorkflowById(workflowAPI.WorkflowId("invalid/id"))
	require.ErrorIs(t, err, workflowAPI.ErrWorkflowInvalidID)
	assert.Nil(t, result)
}

func TestFsRepository_GetWorkflowById_WhenReadFileFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	base := afero.NewMemMapFs()
	workflow := workflowAPI.Workflow{
		Metadata: workflowAPI.Metadata{
			ID:          workflowAPI.WorkflowId("test-workflow"),
			Name:        "Test Workflow",
			Description: "Test workflow",
			Version:     "1.0.0",
		},
		Steps: []workflowAPI.Step{
			{
				ID:           "step1",
				Name:         "Step 1",
				Description:  "First step",
				Instructions: []string{"Do something"},
			},
		},
	}
	data, err := json.Marshal(workflow)
	require.NoError(t, err)
	err = afero.WriteFile(base, "test-workflow.json", data, 0644)
	require.NoError(t, err)

	readFileErr := errors.New("read file failed")
	fs := aferomock.OverrideFs(base, aferomock.FsCallbacks{
		OpenFunc: func(name string) (afero.File, error) {
			if name == "test-workflow.json" {
				return nil, readFileErr
			}
			return base.Open(name)
		},
	})
	repo := NewFsRepository(fs, afero.NewMemMapFs())

	result, err := repo.GetWorkflowById(workflowAPI.WorkflowId("test-workflow"))

	require.ErrorIs(t, err, readFileErr)
	assert.Nil(t, result)
}

func TestFsRepository_RemoveAll_WhenEmpty_ThenSucceeds(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs, afero.NewMemMapFs())

	err := repo.RemoveAllWorkflows()
	require.NoError(t, err)

	workflows, err := repo.GetAllWorkflows()
	require.NoError(t, err)
	assert.Empty(t, workflows)
}

func TestFsRepository_RemoveAll_WhenHasWorkflows_ThenRemovesAll(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs, afero.NewMemMapFs())
	workflow1 := workflowAPI.Workflow{
		Metadata: workflowAPI.Metadata{
			ID:          workflowAPI.WorkflowId("workflow1"),
			Name:        "First Workflow",
			Description: "First workflow",
			Version:     "1.0.0",
		},
		Steps: []workflowAPI.Step{
			{
				ID:           "step1",
				Name:         "Step 1",
				Description:  "First step",
				Instructions: []string{"Do something"},
			},
		},
	}
	workflow2 := workflowAPI.Workflow{
		Metadata: workflowAPI.Metadata{
			ID:          workflowAPI.WorkflowId("workflow2"),
			Name:        "Second Workflow",
			Description: "Second workflow",
			Version:     "1.0.0",
		},
		Steps: []workflowAPI.Step{
			{
				ID:           "step2",
				Name:         "Step 2",
				Description:  "Second step",
				Instructions: []string{"Do something else"},
			},
		},
	}

	err := repo.AddWorkflow(workflow1)
	require.NoError(t, err)
	err = repo.AddWorkflow(workflow2)
	require.NoError(t, err)

	workflows, err := repo.GetAllWorkflows()
	require.NoError(t, err)
	require.Len(t, workflows, 2)

	err = repo.RemoveAllWorkflows()
	require.NoError(t, err)

	workflows, err = repo.GetAllWorkflows()
	require.NoError(t, err)
	assert.Empty(t, workflows)
}

func TestFsRepository_RemoveAll_WhenFollowedByAdd_ThenAcceptsNewWorkflow(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(fs, afero.NewMemMapFs())
	oldWorkflow := workflowAPI.Workflow{
		Metadata: workflowAPI.Metadata{
			ID:          workflowAPI.WorkflowId("new-workflow"),
			Name:        "Old Workflow",
			Description: "Old workflow",
			Version:     "1.0.0",
		},
		Steps: []workflowAPI.Step{
			{
				ID:           "step1",
				Name:         "Step 1",
				Description:  "First step",
				Instructions: []string{"Do something"},
			},
		},
	}

	err := repo.AddWorkflow(oldWorkflow)
	require.NoError(t, err)

	err = repo.RemoveAllWorkflows()
	require.NoError(t, err)

	newWorkflow := workflowAPI.Workflow{
		Metadata: workflowAPI.Metadata{
			ID:          workflowAPI.WorkflowId("new-workflow"),
			Name:        "New Workflow",
			Description: "New workflow",
			Version:     "1.0.0",
		},
		Steps: []workflowAPI.Step{
			{
				ID:           "step2",
				Name:         "Step 2",
				Description:  "Second step",
				Instructions: []string{"Do something else"},
			},
		},
	}

	err = repo.AddWorkflow(newWorkflow)
	require.NoError(t, err)

	workflows, err := repo.GetAllWorkflows()
	require.NoError(t, err)
	require.Len(t, workflows, 1)
	assert.Equal(t, workflowAPI.WorkflowId("new-workflow"), workflows[0].Metadata.ID)
	assert.Equal(t, "New Workflow", workflows[0].Metadata.Name)
}

func TestFsRepository_RemoveAll_WhenReadDirFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	readDirErr := errors.New("read dir failed")
	base := afero.NewMemMapFs()
	fs := aferomock.OverrideFs(base, aferomock.FsCallbacks{
		OpenFunc: func(name string) (afero.File, error) {
			if name == "." {
				return nil, readDirErr
			}
			return base.Open(name)
		},
	})
	repo := NewFsRepository(fs, afero.NewMemMapFs())

	err := repo.RemoveAllWorkflows()

	require.ErrorIs(t, err, readDirErr)
}

func TestFsRepository_RemoveAll_WhenRemoveFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	base := afero.NewMemMapFs()
	workflow := workflowAPI.Workflow{
		Metadata: workflowAPI.Metadata{
			ID:          workflowAPI.WorkflowId("test-workflow"),
			Name:        "Test Workflow",
			Description: "Test workflow",
			Version:     "1.0.0",
		},
		Steps: []workflowAPI.Step{
			{
				ID:           "step1",
				Name:         "Step 1",
				Description:  "First step",
				Instructions: []string{"Do something"},
			},
		},
	}
	data, err := json.Marshal(workflow)
	require.NoError(t, err)
	err = afero.WriteFile(base, "test-workflow.json", data, 0644)
	require.NoError(t, err)

	removeErr := errors.New("remove failed")
	fs := aferomock.OverrideFs(base, aferomock.FsCallbacks{
		RemoveFunc: func(name string) error {
			return removeErr
		},
	})
	repo := NewFsRepository(fs, afero.NewMemMapFs())

	err = repo.RemoveAllWorkflows()

	require.ErrorIs(t, err, removeErr)
}

func TestFsRepository_AddExecution_WhenNewExecution_ThenStoresSuccessfully(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(afero.NewMemMapFs(), fs)
	execution := workflowAPI.Execution{
		Id:         workflowAPI.ExecutionId("test-execution"),
		WorkflowId: workflowAPI.WorkflowId("test-workflow"),
		StateValues: map[string]any{
			"key": "value",
		},
		StepId: "step1",
	}

	err := repo.AddExecution(execution)
	require.NoError(t, err)

	result, err := repo.GetExecutionById(workflowAPI.ExecutionId("test-execution"))
	require.NoError(t, err)
	assert.Equal(t, execution.Id, result.Id)
	assert.Equal(t, execution.WorkflowId, result.WorkflowId)
	assert.Equal(t, execution.StepId, result.StepId)
}

func TestFsRepository_AddExecution_WhenDuplicateID_ThenReturnsAlreadyExistsError(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(afero.NewMemMapFs(), fs)
	execution := workflowAPI.Execution{
		Id:         workflowAPI.ExecutionId("duplicate-execution"),
		WorkflowId: workflowAPI.WorkflowId("test-workflow"),
		StateValues: map[string]any{
			"key": "value1",
		},
		StepId: "step1",
	}

	err := repo.AddExecution(execution)
	require.NoError(t, err)

	duplicateExecution := workflowAPI.Execution{
		Id:         workflowAPI.ExecutionId("duplicate-execution"),
		WorkflowId: workflowAPI.WorkflowId("test-workflow"),
		StateValues: map[string]any{
			"key": "value2",
		},
		StepId: "step2",
	}

	err = repo.AddExecution(duplicateExecution)
	require.ErrorIs(t, err, workflowAPI.ErrExecutionAlreadyExists)
}

func TestFsRepository_AddExecution_WhenInvalidID_ThenReturnsInvalidIDError(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(afero.NewMemMapFs(), fs)
	execution := workflowAPI.Execution{
		Id:         workflowAPI.ExecutionId("invalid/id"),
		WorkflowId: workflowAPI.WorkflowId("test-workflow"),
		StateValues: map[string]any{
			"key": "value",
		},
		StepId: "step1",
	}

	err := repo.AddExecution(execution)
	require.ErrorIs(t, err, workflowAPI.ErrExecutionInvalidID)
}

func TestFsRepository_AddExecution_WhenExistsFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	existsErr := errors.New("exists check failed")
	base := afero.NewMemMapFs()
	fs := aferomock.OverrideFs(base, aferomock.FsCallbacks{
		StatFunc: func(name string) (fs.FileInfo, error) {
			return nil, existsErr
		},
	})
	repo := NewFsRepository(afero.NewMemMapFs(), fs)
	execution := workflowAPI.Execution{
		Id:         workflowAPI.ExecutionId("test-execution"),
		WorkflowId: workflowAPI.WorkflowId("test-workflow"),
		StateValues: map[string]any{
			"key": "value",
		},
		StepId: "step1",
	}

	err := repo.AddExecution(execution)

	require.ErrorIs(t, err, existsErr)
}

func TestFsRepository_AddExecution_WhenWriteFileFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	writeFileErr := errors.New("write file failed")
	base := afero.NewMemMapFs()
	fs := aferomock.OverrideFs(base, aferomock.FsCallbacks{
		OpenFileFunc: func(name string, flag int, perm fs.FileMode) (afero.File, error) {
			return nil, writeFileErr
		},
	})
	repo := NewFsRepository(afero.NewMemMapFs(), fs)
	execution := workflowAPI.Execution{
		Id:         workflowAPI.ExecutionId("test-execution"),
		WorkflowId: workflowAPI.WorkflowId("test-workflow"),
		StateValues: map[string]any{
			"key": "value",
		},
		StepId: "step1",
	}

	err := repo.AddExecution(execution)

	require.ErrorIs(t, err, writeFileErr)
}

func TestFsRepository_UpdateExecution_WhenExecutionExists_ThenUpdatesSuccessfully(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(afero.NewMemMapFs(), fs)
	execution := workflowAPI.Execution{
		Id:         workflowAPI.ExecutionId("test-execution"),
		WorkflowId: workflowAPI.WorkflowId("test-workflow"),
		StateValues: map[string]any{
			"key": "value1",
		},
		StepId: "step1",
	}

	err := repo.AddExecution(execution)
	require.NoError(t, err)

	execution.StateValues = map[string]any{
		"key": "value2",
	}
	execution.StepId = "step2"

	err = repo.UpdateExecution(execution)
	require.NoError(t, err)

	result, err := repo.GetExecutionById(workflowAPI.ExecutionId("test-execution"))
	require.NoError(t, err)
	assert.Equal(t, "step2", result.StepId)
	assert.Equal(t, "value2", result.StateValues["key"])
}

func TestFsRepository_UpdateExecution_WhenExecutionNotFound_ThenReturnsError(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(afero.NewMemMapFs(), fs)
	execution := workflowAPI.Execution{
		Id:         workflowAPI.ExecutionId("non-existing"),
		WorkflowId: workflowAPI.WorkflowId("test-workflow"),
		StateValues: map[string]any{
			"key": "value",
		},
		StepId: "step1",
	}

	err := repo.UpdateExecution(execution)
	require.ErrorIs(t, err, workflowAPI.ErrExecutionNotFound)
}

func TestFsRepository_UpdateExecution_WhenInvalidID_ThenReturnsInvalidIDError(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(afero.NewMemMapFs(), fs)
	execution := workflowAPI.Execution{
		Id:         workflowAPI.ExecutionId("invalid/id"),
		WorkflowId: workflowAPI.WorkflowId("test-workflow"),
		StateValues: map[string]any{
			"key": "value",
		},
		StepId: "step1",
	}

	err := repo.UpdateExecution(execution)
	require.ErrorIs(t, err, workflowAPI.ErrExecutionInvalidID)
}

func TestFsRepository_GetExecutionById_WhenExecutionExists_ThenReturnsExecution(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(afero.NewMemMapFs(), fs)
	execution := workflowAPI.Execution{
		Id:         workflowAPI.ExecutionId("existing-execution"),
		WorkflowId: workflowAPI.WorkflowId("test-workflow"),
		StateValues: map[string]any{
			"key": "value",
		},
		StepId: "step1",
	}

	err := repo.AddExecution(execution)
	require.NoError(t, err)

	result, err := repo.GetExecutionById(workflowAPI.ExecutionId("existing-execution"))
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, workflowAPI.ExecutionId("existing-execution"), result.Id)
	assert.Equal(t, workflowAPI.WorkflowId("test-workflow"), result.WorkflowId)
	assert.Equal(t, "step1", result.StepId)
}

func TestFsRepository_GetExecutionById_WhenExecutionNotFound_ThenReturnsError(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(afero.NewMemMapFs(), fs)
	execution := workflowAPI.Execution{
		Id:         workflowAPI.ExecutionId("some-execution"),
		WorkflowId: workflowAPI.WorkflowId("test-workflow"),
		StateValues: map[string]any{
			"key": "value",
		},
		StepId: "step1",
	}

	err := repo.AddExecution(execution)
	require.NoError(t, err)

	result, err := repo.GetExecutionById(workflowAPI.ExecutionId("non-existing-execution"))
	require.ErrorIs(t, err, workflowAPI.ErrExecutionNotFound)
	assert.Nil(t, result)
}

func TestFsRepository_GetExecutionById_WhenEmptyRepository_ThenReturnsError(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(afero.NewMemMapFs(), fs)

	result, err := repo.GetExecutionById(workflowAPI.ExecutionId("any-execution"))
	require.ErrorIs(t, err, workflowAPI.ErrExecutionNotFound)
	assert.Nil(t, result)
}

func TestFsRepository_GetExecutionById_WhenInvalidID_ThenReturnsInvalidIDError(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	repo := NewFsRepository(afero.NewMemMapFs(), fs)

	result, err := repo.GetExecutionById(workflowAPI.ExecutionId("invalid/id"))
	require.ErrorIs(t, err, workflowAPI.ErrExecutionInvalidID)
	assert.Nil(t, result)
}

func TestFsRepository_GetExecutionById_WhenReadFileFails_ThenReturnsError(t *testing.T) {
	t.Parallel()

	base := afero.NewMemMapFs()
	execution := workflowAPI.Execution{
		Id:         workflowAPI.ExecutionId("test-execution"),
		WorkflowId: workflowAPI.WorkflowId("test-workflow"),
		StateValues: map[string]any{
			"key": "value",
		},
		StepId: "step1",
	}
	data, err := json.Marshal(execution)
	require.NoError(t, err)
	err = afero.WriteFile(base, "test-execution.json", data, 0644)
	require.NoError(t, err)

	readFileErr := errors.New("read file failed")
	fs := aferomock.OverrideFs(base, aferomock.FsCallbacks{
		OpenFunc: func(name string) (afero.File, error) {
			if name == "test-execution.json" {
				return nil, readFileErr
			}
			return base.Open(name)
		},
	})
	repo := NewFsRepository(afero.NewMemMapFs(), fs)

	result, err := repo.GetExecutionById(workflowAPI.ExecutionId("test-execution"))

	require.ErrorIs(t, err, readFileErr)
	assert.Nil(t, result)
}
