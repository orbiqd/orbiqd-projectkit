package workflow

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	workflowAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/workflow"
	projectAPI "github.com/orbiqd/orbiqd-projectkit/pkg/project"
	"github.com/spf13/afero"
)

type FsRepository struct {
	mutex       sync.RWMutex
	workflowFs  afero.Fs
	executionFs afero.Fs
}

var _ workflowAPI.Repository = (*FsRepository)(nil)

func NewFsRepository(workflowFs afero.Fs, executionFs afero.Fs) *FsRepository {
	return &FsRepository{
		mutex:       sync.RWMutex{},
		workflowFs:  workflowFs,
		executionFs: executionFs,
	}
}

func NewFsRepositoryProvider() func(projectAPI.Fs) (workflowAPI.Repository, error) {
	return func(projectFs projectAPI.Fs) (workflowAPI.Repository, error) {
		baseDir := ".projectkit/repository/ai/workflow"
		workflowDir := filepath.Join(baseDir, "workflows")
		executionDir := filepath.Join(baseDir, "executions")

		if err := projectFs.MkdirAll(workflowDir, 0755); err != nil {
			return nil, fmt.Errorf("workflow directory creation: %w", err)
		}

		if err := projectFs.MkdirAll(executionDir, 0755); err != nil {
			return nil, fmt.Errorf("execution directory creation: %w", err)
		}

		workflowFs := afero.NewBasePathFs(projectFs, workflowDir)
		executionFs := afero.NewBasePathFs(projectFs, executionDir)

		return NewFsRepository(workflowFs, executionFs), nil
	}
}

func (repository *FsRepository) listWorkflowFiles() ([]string, error) {
	entries, err := afero.ReadDir(repository.workflowFs, ".")
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.ToLower(filepath.Ext(entry.Name())) == ".json" {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

func (repository *FsRepository) loadWorkflowFile(filename string) (workflowAPI.Workflow, error) {
	data, err := afero.ReadFile(repository.workflowFs, filename)
	if err != nil {
		return workflowAPI.Workflow{}, err
	}

	var workflow workflowAPI.Workflow
	if err := json.Unmarshal(data, &workflow); err != nil {
		return workflowAPI.Workflow{}, err
	}

	return workflow, nil
}

func (repository *FsRepository) saveWorkflowFile(filename string, workflow workflowAPI.Workflow) error {
	data, err := json.Marshal(workflow)
	if err != nil {
		return err
	}

	return afero.WriteFile(repository.workflowFs, filename, data, 0644)
}

func (repository *FsRepository) AddWorkflow(workflow workflowAPI.Workflow) error {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	if err := workflow.Metadata.ID.Validate(); err != nil {
		return err
	}

	filename := string(workflow.Metadata.ID) + ".json"
	exists, err := afero.Exists(repository.workflowFs, filename)
	if err != nil {
		return err
	}
	if exists {
		return workflowAPI.ErrWorkflowAlreadyExists
	}

	return repository.saveWorkflowFile(filename, workflow)
}

func (repository *FsRepository) GetAllWorkflows() ([]workflowAPI.Workflow, error) {
	repository.mutex.RLock()
	defer repository.mutex.RUnlock()

	files, err := repository.listWorkflowFiles()
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return []workflowAPI.Workflow{}, nil
	}

	workflows := make([]workflowAPI.Workflow, 0, len(files))
	for _, file := range files {
		workflow, err := repository.loadWorkflowFile(file)
		if err != nil {
			return nil, err
		}
		workflows = append(workflows, workflow)
	}

	sort.Slice(workflows, func(i, j int) bool {
		return workflows[i].Metadata.Name < workflows[j].Metadata.Name
	})

	return workflows, nil
}

func (repository *FsRepository) GetWorkflowById(id workflowAPI.WorkflowId) (*workflowAPI.Workflow, error) {
	repository.mutex.RLock()
	defer repository.mutex.RUnlock()

	if err := id.Validate(); err != nil {
		return nil, err
	}

	filename := string(id) + ".json"
	exists, err := afero.Exists(repository.workflowFs, filename)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, workflowAPI.ErrWorkflowNotFound
	}

	workflow, err := repository.loadWorkflowFile(filename)
	if err != nil {
		return nil, err
	}

	return &workflow, nil
}

func (repository *FsRepository) RemoveAllWorkflows() error {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	files, err := repository.listWorkflowFiles()
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := repository.workflowFs.Remove(file); err != nil {
			return err
		}
	}

	return nil
}

func (repository *FsRepository) loadExecutionFile(filename string) (workflowAPI.Execution, error) {
	data, err := afero.ReadFile(repository.executionFs, filename)
	if err != nil {
		return workflowAPI.Execution{}, err
	}

	var execution workflowAPI.Execution
	if err := json.Unmarshal(data, &execution); err != nil {
		return workflowAPI.Execution{}, err
	}

	return execution, nil
}

func (repository *FsRepository) saveExecutionFile(filename string, execution workflowAPI.Execution) error {
	data, err := json.Marshal(execution)
	if err != nil {
		return err
	}

	return afero.WriteFile(repository.executionFs, filename, data, 0644)
}

func (repository *FsRepository) AddExecution(execution workflowAPI.Execution) error {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	if err := execution.Id.Validate(); err != nil {
		return err
	}

	filename := string(execution.Id) + ".json"
	exists, err := afero.Exists(repository.executionFs, filename)
	if err != nil {
		return err
	}
	if exists {
		return workflowAPI.ErrExecutionAlreadyExists
	}

	return repository.saveExecutionFile(filename, execution)
}

func (repository *FsRepository) UpdateExecution(execution workflowAPI.Execution) error {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	if err := execution.Id.Validate(); err != nil {
		return err
	}

	filename := string(execution.Id) + ".json"
	exists, err := afero.Exists(repository.executionFs, filename)
	if err != nil {
		return err
	}
	if !exists {
		return workflowAPI.ErrExecutionNotFound
	}

	return repository.saveExecutionFile(filename, execution)
}

func (repository *FsRepository) GetExecutionById(id workflowAPI.ExecutionId) (*workflowAPI.Execution, error) {
	repository.mutex.RLock()
	defer repository.mutex.RUnlock()

	if err := id.Validate(); err != nil {
		return nil, err
	}

	filename := string(id) + ".json"
	exists, err := afero.Exists(repository.executionFs, filename)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, workflowAPI.ErrExecutionNotFound
	}

	execution, err := repository.loadExecutionFile(filename)
	if err != nil {
		return nil, err
	}

	return &execution, nil
}
