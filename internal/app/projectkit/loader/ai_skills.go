package loader

import (
	"fmt"

	skillAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/skill"
	sourceAPI "github.com/orbiqd/orbiqd-projectkit/pkg/source"
)

func LoadAiSkillsFromConfig(config skillAPI.Config, sourceResolver sourceAPI.Resolver) ([]skillAPI.Skill, error) {
	var skills []skillAPI.Skill

	for _, skillConfig := range config.Sources {
		skillsUri := skillConfig.URI
		skillsSource, err := sourceResolver.Resolve(skillConfig.URI)
		if err != nil {
			return []skillAPI.Skill{}, fmt.Errorf("resolve: %s: %w", skillsUri, err)
		}

		skillLoader := skillAPI.NewLoader(skillsSource)

		skillsSet, err := skillLoader.Load()
		if err != nil {
			return []skillAPI.Skill{}, fmt.Errorf("load skills: %w", err)
		}

		skills = append(skills, skillsSet...)
	}

	return skills, nil
}
