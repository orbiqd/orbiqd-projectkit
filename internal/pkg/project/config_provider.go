package project

import projectAPI "github.com/orbiqd/orbiqd-projectkit/pkg/project"

func NewConfigProvider() func(projectAPI.ConfigLoader) (*projectAPI.Config, error) {
	return func(configLoader projectAPI.ConfigLoader) (*projectAPI.Config, error) {
		return configLoader.Load()
	}
}
