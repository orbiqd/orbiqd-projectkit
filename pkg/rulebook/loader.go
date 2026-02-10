package rulebook

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	instructionAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/instruction"
	skillAPI "github.com/orbiqd/orbiqd-projectkit/pkg/ai/skill"
	"github.com/spf13/afero"
	"sigs.k8s.io/yaml"
)

const rulebookFileName = "rulebook.yaml"

type Loader struct {
	fs afero.Fs
}

func NewLoader(fs afero.Fs) *Loader {
	return &Loader{
		fs: fs,
	}
}

func (loader *Loader) Load() (*Rulebook, error) {
	metadata, err := loader.loadMetadata()
	if err != nil {
		return nil, fmt.Errorf("load metadata: %w", err)
	}

	err = loader.validateMetadata(*metadata)
	if err != nil {
		return nil, fmt.Errorf("validate metadata: %w", err)
	}

	rulebook := Rulebook{
		AI: AiRulebook{
			Instructions: []instructionAPI.Instructions{},
			Skills:       []skillAPI.Skill{},
		},
	}

	if metadata.AI != nil && metadata.AI.Instruction != nil {
		for _, aiInstructionsSource := range metadata.AI.Instruction.Sources {
			aiInstructionsPath, err := loader.resolveSourceUri(aiInstructionsSource.URI)
			if err != nil {
				return nil, fmt.Errorf("ai instructions: resolve source path: %w", err)
			}

			aiInstructions, err := instructionAPI.NewLoader(
				afero.NewBasePathFs(loader.fs, aiInstructionsPath),
			).Load()
			if err != nil {
				return nil, fmt.Errorf("ai instructions: load ai instructions: %w", err)
			}

			rulebook.AI.Instructions = append(rulebook.AI.Instructions, aiInstructions...)
		}
	}

	if metadata.AI != nil && metadata.AI.Skill != nil {
		for _, aiSkillsSource := range metadata.AI.Skill.Sources {
			aiSkillsPath, err := loader.resolveSourceUri(aiSkillsSource.URI)
			if err != nil {
				return nil, fmt.Errorf("ai skills: resolve source path: %w", err)
			}

			aiSkills, err := skillAPI.NewLoader(
				afero.NewBasePathFs(loader.fs, aiSkillsPath),
			).Load()
			if err != nil {
				return nil, fmt.Errorf("ai skills: load ai skills: %w", err)
			}

			rulebook.AI.Skills = append(rulebook.AI.Skills, aiSkills...)
		}
	}

	return &rulebook, nil
}

func (loader *Loader) loadMetadata() (*Metadata, error) {
	data, err := afero.ReadFile(loader.fs, rulebookFileName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrMissingMetadataFile
		}
		return nil, fmt.Errorf("%w: %v", ErrReadFailed, err)
	}

	var metadata Metadata
	if err := yaml.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParseFailed, err)
	}

	return &metadata, nil
}

func (loader *Loader) validateMetadata(metadata Metadata) error {
	validate := validator.New()
	if err := validate.Struct(metadata); err != nil {
		return fmt.Errorf("%w: %v", ErrValidationFailed, err)
	}
	return nil
}

func (loader *Loader) resolveSourceUri(uri string) (string, error) {
	path, found := strings.CutPrefix(uri, "rulebook://")
	if !found {
		return "", fmt.Errorf("%w: %s", ErrUnsupportedScheme, uri)
	}
	if path == "" {
		return "", fmt.Errorf("%w: %s", ErrEmptyPath, uri)
	}
	return "/" + path, nil
}

var ErrMissingMetadataFile = errors.New("missing metadata file")
var ErrReadFailed = errors.New("read failed")
var ErrParseFailed = errors.New("parse failed")
var ErrValidationFailed = errors.New("validation failed")
var ErrUnsupportedScheme = errors.New("unsupported scheme")
var ErrEmptyPath = errors.New("empty path")
