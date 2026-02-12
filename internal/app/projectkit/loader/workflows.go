package loader

import (
	"fmt"

	sourceAPI "github.com/orbiqd/orbiqd-projectkit/pkg/source"
	workflowAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/workflow"
)

func LoadWorkflowsFromConfig(config workflowAPI.Config, sourceResolver sourceAPI.Resolver) ([]workflowAPI.Workflow, error) {
	var workflows []workflowAPI.Workflow

	for _, workflowConfig := range config.Sources {
		workflowsUri := workflowConfig.URI
		workflowsSource, err := sourceResolver.Resolve(workflowConfig.URI)
		if err != nil {
			return []workflowAPI.Workflow{}, fmt.Errorf("resolve: %s: %w", workflowsUri, err)
		}

		workflowLoader := workflowAPI.NewLoader(workflowsSource)

		workflowsSet, err := workflowLoader.Load()
		if err != nil {
			return []workflowAPI.Workflow{}, fmt.Errorf("load workflows: %w", err)
		}

		workflows = append(workflows, workflowsSet...)
	}

	return workflows, nil
}
