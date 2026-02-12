package loader

import (
	"fmt"

	standardAPI "github.com/orbiqd/orbiqd-projectkit/pkg/doc/standard"
	sourceAPI "github.com/orbiqd/orbiqd-projectkit/pkg/source"
)

func LoadDocStandardsFromConfig(config standardAPI.Config, sourceResolver sourceAPI.Resolver) ([]standardAPI.Standard, error) {
	var standards []standardAPI.Standard

	for _, standardSourceConfig := range config.Sources {
		standardUri := standardSourceConfig.URI
		standardSource, err := sourceResolver.Resolve(standardUri)
		if err != nil {
			return []standardAPI.Standard{}, fmt.Errorf("resolve: %s: %w", standardUri, err)
		}

		standardLoader := standardAPI.NewLoader(standardSource)
		loadedStandards, err := standardLoader.Load()
		if err != nil {
			return []standardAPI.Standard{}, fmt.Errorf("load standard: %w", err)
		}

		standards = append(standards, loadedStandards...)
	}

	return standards, nil
}
