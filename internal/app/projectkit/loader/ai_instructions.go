package loader

import (
	"fmt"

	instructionAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/instruction"
	sourceAPI "github.com/orbiqd/orbiqd-projectkit/pkg/source"
)

func LoadAiInstructionsFromConfig(config instructionAPI.Config, sourceResolver sourceAPI.Resolver) ([]instructionAPI.Instructions, error) {
	var instructions []instructionAPI.Instructions

	for _, instructionSourceConfig := range config.Sources {
		instructionsUri := instructionSourceConfig.URI
		instructionSource, err := sourceResolver.Resolve(instructionsUri)
		if err != nil {
			return nil, fmt.Errorf("resolve: %s: %w", instructionsUri, err)
		}

		instructionLoader := instructionAPI.NewLoader(instructionSource)
		instructionsSet, err := instructionLoader.Load()
		if err != nil {
			return nil, fmt.Errorf("load instructions: %w", err)
		}

		instructions = append(instructions, instructionsSet...)
	}

	return instructions, nil
}
