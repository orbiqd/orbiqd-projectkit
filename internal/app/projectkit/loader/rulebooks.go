package loader

import (
	"fmt"
	"log/slog"

	rulebookAPI "github.com/orbiqd/orbiqd-projectkit/pkg/rulebook"
	sourceAPI "github.com/orbiqd/orbiqd-projectkit/pkg/source"
)

func LoadRulebooksFromConfig(config rulebookAPI.Config, sourceResolver sourceAPI.Resolver) ([]rulebookAPI.Rulebook, error) {
	var rulebooks []rulebookAPI.Rulebook

	for _, rulebookSourceConfig := range config.Sources {
		rulebookUri := rulebookSourceConfig.URI

		rulebookSource, err := sourceResolver.Resolve(rulebookUri)
		if err != nil {
			return []rulebookAPI.Rulebook{}, fmt.Errorf("resolve rulebook uri: %w", err)
		}

		rulebook, err := rulebookAPI.NewLoader(rulebookSource).Load()
		if err != nil {
			return []rulebookAPI.Rulebook{}, fmt.Errorf("load rulebook: %w", err)
		}

		rulebooks = append(rulebooks, *rulebook)

		slog.Info("Loaded rulebook.",
			slog.String("sourceUri", rulebookUri),
			slog.Int("aiInstructionsCount", len(rulebook.AI.Instructions)),
			slog.Int("aiSkillsCount", len(rulebook.AI.Skills)),
			slog.Int("aiMCPServersCount", len(rulebook.AI.MCPServers)),
		)
	}

	return rulebooks, nil
}
