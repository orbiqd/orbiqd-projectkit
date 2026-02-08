package skill

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/afero"
	"sigs.k8s.io/yaml"
)

const metadataFileName = "metadata.yaml"
const instructionsFileName = "instructions.md"
const scriptsDirName = "scripts"

var scriptContentTypes = map[string]string{
	".sh":   "application/x-sh",
	".bash": "application/x-sh",
	".zsh":  "application/x-sh",
	".ksh":  "application/x-sh",
	".csh":  "application/x-csh",
	".fish": "application/x-fish",
	".py":   "text/x-python",
	".rb":   "text/x-ruby",
	".pl":   "text/x-perl",
	".lua":  "text/x-lua",
	".js":   "text/javascript",
	".awk":  "text/x-awk",
	".sed":  "text/x-sed",
	".tcl":  "application/x-tcl",
}

type Loader struct {
	rootFs afero.Fs
}

func NewLoader(rootFs afero.Fs) *Loader {
	return &Loader{
		rootFs: rootFs,
	}
}

func (loader *Loader) Load() ([]Skill, error) {
	skillPaths, err := loader.resolvePaths()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve paths: %w", err)
	}

	var skills []Skill
	for _, skillPath := range skillPaths {
		skill, err := loader.loadSkill(skillPath)
		if err != nil {
			return []Skill{}, fmt.Errorf("failed to load skill: %s: %w", skillPaths, err)
		}

		skills = append(skills, *skill)
	}

	if len(skills) == 0 {
		return []Skill{}, ErrNoSkillsFound
	}

	return skills, nil
}

func (loader *Loader) resolvePaths() ([]string, error) {
	entries, err := afero.ReadDir(loader.rootFs, ".")
	if err != nil {
		return nil, fmt.Errorf("read directory: %w", err)
	}

	var dirPaths []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dirPaths = append(dirPaths, entry.Name())
	}

	return dirPaths, nil
}

func (loader *Loader) loadSkill(path string) (*Skill, error) {
	var err error

	skillFs := afero.NewReadOnlyFs(
		afero.NewBasePathFs(loader.rootFs, path),
	)

	metadata, err := loader.loadMetadata(skillFs)
	if err != nil {
		return nil, fmt.Errorf("failed to load metadata: %w", err)
	}

	err = loader.validateMetadata(*metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to validate metadata: %w", err)
	}

	instructions, err := loader.loadInstructions(skillFs) //nolint:staticcheck // SA4006: false positive, instructions is used in return statement
	if err != nil {
		return nil, fmt.Errorf("failed to load instructions: %w", err)
	}

	scriptsPaths, err := loader.resolveScriptsPaths(skillFs)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve scripts paths: %w", err)
	}

	scripts := make(map[ScriptName]Script)

	for _, scriptPath := range scriptsPaths {
		scriptName, script, err := loader.loadScript(skillFs, scriptPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load script: %s: %w", scriptPath, err)
		}

		scripts[*scriptName] = *script
	}

	return &Skill{
		Metadata:     *metadata,
		Instructions: *instructions,
		Scripts:      scripts,
	}, nil
}

func (loader *Loader) resolveScriptsPaths(skillFs afero.Fs) ([]string, error) {
	exists, err := afero.DirExists(skillFs, scriptsDirName)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrReadFailed, err)
	}

	if !exists {
		return []string{}, nil
	}

	entries, err := afero.ReadDir(skillFs, scriptsDirName)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrReadFailed, err)
	}

	var paths []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		paths = append(paths, scriptsDirName+"/"+entry.Name())
	}

	return paths, nil
}

func resolveContentType(filename string) string {
	ext := filepath.Ext(filename)
	if ct, ok := scriptContentTypes[ext]; ok {
		return ct
	}
	return "application/octet-stream"
}

func (loader *Loader) loadScript(skillFs afero.Fs, path string) (*ScriptName, *Script, error) {
	data, err := afero.ReadFile(skillFs, path)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrReadFailed, err)
	}

	filename := filepath.Base(path)
	name := ScriptName(filename)

	script := &Script{
		ContentType: resolveContentType(filename),
		Content:     data,
	}

	return &name, script, nil
}

func (loader *Loader) loadInstructions(skillFs afero.Fs) (*string, error) {
	data, err := afero.ReadFile(skillFs, instructionsFileName)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrReadFailed, err)
	}

	content := string(data)
	return &content, nil
}

func (loader *Loader) loadMetadata(skillFs afero.Fs) (*Metadata, error) {
	data, err := afero.ReadFile(skillFs, metadataFileName)
	if err != nil {
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

var ErrNoSkillsFound = errors.New("no skills found")
var ErrReadFailed = errors.New("read failed")
var ErrParseFailed = errors.New("parse failed")
var ErrValidationFailed = errors.New("validation failed")
