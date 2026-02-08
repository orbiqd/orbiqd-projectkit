package skill

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.nhat.io/aferomock"
)

func fsWithOpenError(t *testing.T, base afero.Fs, shouldError func(name string) bool) afero.Fs {
	t.Helper()
	return aferomock.OverrideFs(base, aferomock.FsCallbacks{
		OpenFunc: func(name string) (afero.File, error) {
			if shouldError(name) {
				return nil, errors.New("simulated fs error")
			}
			return base.Open(name)
		},
	})
}

func fsWithStatError(t *testing.T, base afero.Fs) afero.Fs {
	t.Helper()
	return aferomock.OverrideFs(base, aferomock.FsCallbacks{
		StatFunc: func(name string) (os.FileInfo, error) {
			return nil, errors.New("simulated stat error")
		},
	})
}

func TestLoader_resolvePaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupFs    func(fs afero.Fs)
		wantLen    int
		wantErr    error
		errContain string
	}{
		{
			name: "single directory",
			setupFs: func(fs afero.Fs) {
				_ = fs.Mkdir("skill1", 0755)
			},
			wantLen: 1,
			wantErr: nil,
		},
		{
			name: "multiple directories",
			setupFs: func(fs afero.Fs) {
				_ = fs.Mkdir("skill1", 0755)
				_ = fs.Mkdir("skill2", 0755)
				_ = fs.Mkdir("skill3", 0755)
			},
			wantLen: 3,
			wantErr: nil,
		},
		{
			name: "empty filesystem",
			setupFs: func(fs afero.Fs) {
			},
			wantLen: 0,
			wantErr: nil,
		},
		{
			name: "only files, no directories",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "file1.txt", []byte("content"), 0644)
				_ = afero.WriteFile(fs, "file2.yaml", []byte("content"), 0644)
			},
			wantLen: 0,
			wantErr: nil,
		},
		{
			name: "mix of files and directories",
			setupFs: func(fs afero.Fs) {
				_ = fs.Mkdir("skill1", 0755)
				_ = fs.Mkdir("skill2", 0755)
				_ = afero.WriteFile(fs, "file1.txt", []byte("content"), 0644)
				_ = afero.WriteFile(fs, "file2.yaml", []byte("content"), 0644)
			},
			wantLen: 2,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fs := afero.NewMemMapFs()
			if tt.setupFs != nil {
				tt.setupFs(fs)
			}

			loader := NewLoader(fs)
			got, err := loader.resolvePaths()

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else if tt.errContain != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContain)
			} else {
				require.NoError(t, err)
				assert.Len(t, got, tt.wantLen)
			}
		})
	}
}

func TestLoader_resolvePaths_WhenReadDirectoryError_ThenReturnsError(t *testing.T) {
	t.Parallel()

	fs := fsWithOpenError(t, afero.NewMemMapFs(), func(string) bool { return true })
	loader := NewLoader(fs)

	got, err := loader.resolvePaths()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "read directory")
	assert.Len(t, got, 0)
}

func TestLoader_loadMetadata(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupFs     func(fs afero.Fs)
		wantErr     error
		wantName    Name
		wantDesc    string
		errContain  string
	}{
		{
			name: "valid metadata",
			setupFs: func(fs afero.Fs) {
				content := `name: test-skill
description: A test skill description`
				_ = afero.WriteFile(fs, metadataFileName, []byte(content), 0644)
			},
			wantErr:  nil,
			wantName: "test-skill",
			wantDesc: "A test skill description",
		},
		{
			name: "file not found",
			setupFs: func(fs afero.Fs) {
			},
			wantErr: ErrReadFailed,
		},
		{
			name: "invalid YAML",
			setupFs: func(fs afero.Fs) {
				content := `name: test-skill
description: [invalid: yaml: structure`
				_ = afero.WriteFile(fs, metadataFileName, []byte(content), 0644)
			},
			wantErr: ErrParseFailed,
		},
		{
			name: "empty file",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, metadataFileName, []byte(""), 0644)
			},
			wantErr:  nil,
			wantName: "",
			wantDesc: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fs := afero.NewMemMapFs()
			if tt.setupFs != nil {
				tt.setupFs(fs)
			}

			loader := NewLoader(fs)
			got, err := loader.loadMetadata(fs)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantName, got.Name)
				assert.Equal(t, tt.wantDesc, got.Description)
			}
		})
	}
}

func TestLoader_validateMetadata(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		metadata Metadata
		wantErr  error
	}{
		{
			name: "valid metadata",
			metadata: Metadata{
				Name:        "test",
				Description: "desc",
			},
			wantErr: nil,
		},
		{
			name: "empty name",
			metadata: Metadata{
				Name:        "",
				Description: "desc",
			},
			wantErr: ErrValidationFailed,
		},
		{
			name: "empty description",
			metadata: Metadata{
				Name:        "test",
				Description: "",
			},
			wantErr: ErrValidationFailed,
		},
		{
			name: "description exceeds max length",
			metadata: Metadata{
				Name:        "test",
				Description: "a" + string(make([]byte, 256)),
			},
			wantErr: ErrValidationFailed,
		},
		{
			name: "description at max length",
			metadata: Metadata{
				Name:        "test",
				Description: string(make([]byte, 256)),
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			loader := NewLoader(afero.NewMemMapFs())
			err := loader.validateMetadata(tt.metadata)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestLoader_loadInstructions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupFs     func(fs afero.Fs)
		wantErr     error
		wantContent *string
	}{
		{
			name: "valid instructions",
			setupFs: func(fs afero.Fs) {
				content := `# Test Instructions

This is a test instruction file with markdown content.

## Steps
1. First step
2. Second step`
				_ = afero.WriteFile(fs, instructionsFileName, []byte(content), 0644)
			},
			wantErr: nil,
			wantContent: func() *string {
				s := `# Test Instructions

This is a test instruction file with markdown content.

## Steps
1. First step
2. Second step`
				return &s
			}(),
		},
		{
			name: "file not found",
			setupFs: func(fs afero.Fs) {
			},
			wantErr: ErrReadFailed,
		},
		{
			name: "empty file",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, instructionsFileName, []byte(""), 0644)
			},
			wantErr: nil,
			wantContent: func() *string {
				s := ""
				return &s
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fs := afero.NewMemMapFs()
			if tt.setupFs != nil {
				tt.setupFs(fs)
			}

			loader := NewLoader(fs)
			got, err := loader.loadInstructions(fs)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, *tt.wantContent, *got)
			}
		})
	}
}

func TestLoader_resolveScriptsPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupFs   func(fs afero.Fs)
		wantPaths []string
		wantErr   error
	}{
		{
			name: "scripts dir with files",
			setupFs: func(fs afero.Fs) {
				_ = fs.Mkdir("scripts", 0755)
				_ = afero.WriteFile(fs, "scripts/a.sh", []byte("#!/bin/bash"), 0755)
				_ = afero.WriteFile(fs, "scripts/b.sh", []byte("#!/bin/bash"), 0755)
			},
			wantPaths: []string{"scripts/a.sh", "scripts/b.sh"},
			wantErr:   nil,
		},
		{
			name: "scripts dir does not exist",
			setupFs: func(fs afero.Fs) {
			},
			wantPaths: []string{},
			wantErr:   nil,
		},
		{
			name: "scripts dir is empty",
			setupFs: func(fs afero.Fs) {
				_ = fs.Mkdir("scripts", 0755)
			},
			wantPaths: []string{},
			wantErr:   nil,
		},
		{
			name: "ignores subdirectories",
			setupFs: func(fs afero.Fs) {
				_ = fs.Mkdir("scripts", 0755)
				_ = afero.WriteFile(fs, "scripts/script.sh", []byte("#!/bin/bash"), 0755)
				_ = fs.Mkdir("scripts/subdir", 0755)
			},
			wantPaths: []string{"scripts/script.sh"},
			wantErr:   nil,
		},
		{
			name: "multiple files",
			setupFs: func(fs afero.Fs) {
				_ = fs.Mkdir("scripts", 0755)
				_ = afero.WriteFile(fs, "scripts/first.sh", []byte("#!/bin/bash"), 0755)
				_ = afero.WriteFile(fs, "scripts/second.sh", []byte("#!/bin/bash"), 0755)
				_ = afero.WriteFile(fs, "scripts/third.py", []byte("#!/usr/bin/env python"), 0755)
			},
			wantPaths: []string{"scripts/first.sh", "scripts/second.sh", "scripts/third.py"},
			wantErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fs := afero.NewMemMapFs()
			if tt.setupFs != nil {
				tt.setupFs(fs)
			}

			loader := NewLoader(fs)
			got, err := loader.resolveScriptsPaths(fs)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.ElementsMatch(t, tt.wantPaths, got)
			}
		})
	}
}

func TestLoader_resolveScriptsPaths_WhenStatError_ThenReturnsReadFailed(t *testing.T) {
	t.Parallel()

	fs := fsWithStatError(t, afero.NewMemMapFs())
	loader := NewLoader(fs)

	got, err := loader.resolveScriptsPaths(fs)

	require.ErrorIs(t, err, ErrReadFailed)
	assert.Nil(t, got)
}

func TestLoader_resolveScriptsPaths_WhenReadingDirectoryError_ThenReturnsReadFailed(t *testing.T) {
	t.Parallel()

	base := afero.NewMemMapFs()
	_ = base.Mkdir("scripts", 0755)
	fs := fsWithOpenError(t, base, func(name string) bool {
		return name == scriptsDirName || name == "."+string(afero.FilePathSeparator)+scriptsDirName
	})
	loader := NewLoader(fs)

	got, err := loader.resolveScriptsPaths(fs)

	require.ErrorIs(t, err, ErrReadFailed)
	assert.Nil(t, got)
}

func TestResolveContentType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{
			name:     "shell script .sh",
			filename: "script.sh",
			want:     "application/x-sh",
		},
		{
			name:     "bash script .bash",
			filename: "script.bash",
			want:     "application/x-sh",
		},
		{
			name:     "zsh script .zsh",
			filename: "script.zsh",
			want:     "application/x-sh",
		},
		{
			name:     "ksh script .ksh",
			filename: "script.ksh",
			want:     "application/x-sh",
		},
		{
			name:     "csh script .csh",
			filename: "script.csh",
			want:     "application/x-csh",
		},
		{
			name:     "fish script .fish",
			filename: "script.fish",
			want:     "application/x-fish",
		},
		{
			name:     "python script .py",
			filename: "script.py",
			want:     "text/x-python",
		},
		{
			name:     "ruby script .rb",
			filename: "script.rb",
			want:     "text/x-ruby",
		},
		{
			name:     "perl script .pl",
			filename: "script.pl",
			want:     "text/x-perl",
		},
		{
			name:     "lua script .lua",
			filename: "script.lua",
			want:     "text/x-lua",
		},
		{
			name:     "javascript .js",
			filename: "script.js",
			want:     "text/javascript",
		},
		{
			name:     "awk script .awk",
			filename: "script.awk",
			want:     "text/x-awk",
		},
		{
			name:     "sed script .sed",
			filename: "script.sed",
			want:     "text/x-sed",
		},
		{
			name:     "tcl script .tcl",
			filename: "script.tcl",
			want:     "application/x-tcl",
		},
		{
			name:     "unknown extension",
			filename: "script.xyz",
			want:     "application/octet-stream",
		},
		{
			name:     "no extension",
			filename: "script",
			want:     "application/octet-stream",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := resolveContentType(tt.filename)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLoader_loadScript(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		setupFs         func(fs afero.Fs)
		path            string
		wantScriptName  ScriptName
		wantContentType string
		wantContent     []byte
		wantErr         error
	}{
		{
			name: "shell script",
			setupFs: func(fs afero.Fs) {
				_ = fs.Mkdir("scripts", 0755)
				_ = afero.WriteFile(fs, "scripts/run.sh", []byte("#!/bin/bash\necho 'hello'"), 0755)
			},
			path:            "scripts/run.sh",
			wantScriptName:  "run.sh",
			wantContentType: "application/x-sh",
			wantContent:     []byte("#!/bin/bash\necho 'hello'"),
			wantErr:         nil,
		},
		{
			name: "python script",
			setupFs: func(fs afero.Fs) {
				_ = fs.Mkdir("scripts", 0755)
				_ = afero.WriteFile(fs, "scripts/deploy.py", []byte("#!/usr/bin/env python\nprint('deploying')"), 0755)
			},
			path:            "scripts/deploy.py",
			wantScriptName:  "deploy.py",
			wantContentType: "text/x-python",
			wantContent:     []byte("#!/usr/bin/env python\nprint('deploying')"),
			wantErr:         nil,
		},
		{
			name: "unknown extension",
			setupFs: func(fs afero.Fs) {
				_ = fs.Mkdir("scripts", 0755)
				_ = afero.WriteFile(fs, "scripts/task.xyz", []byte("custom content"), 0755)
			},
			path:            "scripts/task.xyz",
			wantScriptName:  "task.xyz",
			wantContentType: "application/octet-stream",
			wantContent:     []byte("custom content"),
			wantErr:         nil,
		},
		{
			name: "file not found",
			setupFs: func(fs afero.Fs) {
			},
			path:    "scripts/nonexistent.sh",
			wantErr: ErrReadFailed,
		},
		{
			name: "bash extension",
			setupFs: func(fs afero.Fs) {
				_ = fs.Mkdir("scripts", 0755)
				_ = afero.WriteFile(fs, "scripts/init.bash", []byte("#!/bin/bash\ninit"), 0755)
			},
			path:            "scripts/init.bash",
			wantScriptName:  "init.bash",
			wantContentType: "application/x-sh",
			wantContent:     []byte("#!/bin/bash\ninit"),
			wantErr:         nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fs := afero.NewMemMapFs()
			if tt.setupFs != nil {
				tt.setupFs(fs)
			}

			loader := NewLoader(fs)
			scriptName, script, err := loader.loadScript(fs, tt.path)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, scriptName)
				assert.Nil(t, script)
			} else {
				require.NoError(t, err)
				require.NotNil(t, scriptName)
				require.NotNil(t, script)
				assert.Equal(t, tt.wantScriptName, *scriptName)
				assert.Equal(t, tt.wantContentType, script.ContentType)
				assert.Equal(t, tt.wantContent, script.Content)
			}
		})
	}
}

func TestLoader_loadSkill_WhenCompleteSkillWithScripts_ThenLoadsSuccessfully(t *testing.T) {
	t.Parallel()

	const skillPath = "myskill"
	fs := afero.NewMemMapFs()
	_ = fs.Mkdir(skillPath, 0755)
	metadataContent := `name: test-skill
description: A test skill with scripts`
	_ = afero.WriteFile(fs, skillPath+"/"+metadataFileName, []byte(metadataContent), 0644)
	instructionsContent := "# Test Instructions\n\nThis is a test skill."
	_ = afero.WriteFile(fs, skillPath+"/"+instructionsFileName, []byte(instructionsContent), 0644)
	_ = fs.Mkdir(skillPath+"/scripts", 0755)
	_ = afero.WriteFile(fs, skillPath+"/scripts/run.sh", []byte("#!/bin/bash\necho 'test'"), 0755)

	loader := NewLoader(fs)
	skill, err := loader.loadSkill(skillPath)

	require.NoError(t, err)
	require.NotNil(t, skill)
	assert.Equal(t, Name("test-skill"), skill.Metadata.Name)
	assert.Equal(t, "A test skill with scripts", skill.Metadata.Description)
	assert.Equal(t, "# Test Instructions\n\nThis is a test skill.", skill.Instructions)
	require.Len(t, skill.Scripts, 1)
	script, exists := skill.Scripts["run.sh"]
	require.True(t, exists)
	assert.Equal(t, "application/x-sh", script.ContentType)
	assert.Equal(t, []byte("#!/bin/bash\necho 'test'"), script.Content)
}

func TestLoader_loadSkill_WhenSkillWithoutScriptsDirectory_ThenLoadsSuccessfully(t *testing.T) {
	t.Parallel()

	const skillPath = "myskill"
	fs := afero.NewMemMapFs()
	_ = fs.Mkdir(skillPath, 0755)
	metadataContent := `name: simple-skill
description: A simple skill without scripts`
	_ = afero.WriteFile(fs, skillPath+"/"+metadataFileName, []byte(metadataContent), 0644)
	instructionsContent := "# Simple Skill\n\nNo scripts here."
	_ = afero.WriteFile(fs, skillPath+"/"+instructionsFileName, []byte(instructionsContent), 0644)

	loader := NewLoader(fs)
	skill, err := loader.loadSkill(skillPath)

	require.NoError(t, err)
	require.NotNil(t, skill)
	assert.Equal(t, Name("simple-skill"), skill.Metadata.Name)
	assert.Equal(t, "A simple skill without scripts", skill.Metadata.Description)
	assert.Equal(t, "# Simple Skill\n\nNo scripts here.", skill.Instructions)
	assert.Empty(t, skill.Scripts)
}

func TestLoader_loadSkill_WhenMissingMetadata_ThenReturnsReadFailed(t *testing.T) {
	t.Parallel()

	const skillPath = "myskill"
	fs := afero.NewMemMapFs()
	_ = fs.Mkdir(skillPath, 0755)
	instructionsContent := "# Instructions Only\n\nNo metadata file."
	_ = afero.WriteFile(fs, skillPath+"/"+instructionsFileName, []byte(instructionsContent), 0644)

	loader := NewLoader(fs)
	skill, err := loader.loadSkill(skillPath)

	require.ErrorIs(t, err, ErrReadFailed)
	assert.Nil(t, skill)
}

func TestLoader_loadSkill_WhenInvalidMetadataYAML_ThenReturnsParseFailed(t *testing.T) {
	t.Parallel()

	const skillPath = "myskill"
	fs := afero.NewMemMapFs()
	_ = fs.Mkdir(skillPath, 0755)
	metadataContent := `name: test
description: [invalid: yaml: structure`
	_ = afero.WriteFile(fs, skillPath+"/"+metadataFileName, []byte(metadataContent), 0644)
	instructionsContent := "# Test"
	_ = afero.WriteFile(fs, skillPath+"/"+instructionsFileName, []byte(instructionsContent), 0644)

	loader := NewLoader(fs)
	skill, err := loader.loadSkill(skillPath)

	require.ErrorIs(t, err, ErrParseFailed)
	assert.Nil(t, skill)
}

func TestLoader_loadSkill_WhenMetadataValidationFails_ThenReturnsValidationFailed(t *testing.T) {
	t.Parallel()

	const skillPath = "myskill"
	fs := afero.NewMemMapFs()
	_ = fs.Mkdir(skillPath, 0755)
	metadataContent := `name: ""
description: Valid description`
	_ = afero.WriteFile(fs, skillPath+"/"+metadataFileName, []byte(metadataContent), 0644)
	instructionsContent := "# Test"
	_ = afero.WriteFile(fs, skillPath+"/"+instructionsFileName, []byte(instructionsContent), 0644)

	loader := NewLoader(fs)
	skill, err := loader.loadSkill(skillPath)

	require.ErrorIs(t, err, ErrValidationFailed)
	assert.Nil(t, skill)
}

func TestLoader_loadSkill_WhenMissingInstructions_ThenReturnsReadFailed(t *testing.T) {
	t.Parallel()

	const skillPath = "myskill"
	fs := afero.NewMemMapFs()
	_ = fs.Mkdir(skillPath, 0755)
	metadataContent := `name: test-skill
description: A test skill`
	_ = afero.WriteFile(fs, skillPath+"/"+metadataFileName, []byte(metadataContent), 0644)

	loader := NewLoader(fs)
	skill, err := loader.loadSkill(skillPath)

	require.ErrorIs(t, err, ErrReadFailed)
	assert.Nil(t, skill)
}

func TestLoader_loadSkill_WhenErrorResolvingScriptsPaths_ThenReturnsReadFailed(t *testing.T) {
	t.Parallel()

	const skillPath = "myskill"
	base := afero.NewMemMapFs()
	_ = base.Mkdir(skillPath, 0755)
	metadataContent := `name: test-skill
description: A test skill`
	_ = afero.WriteFile(base, skillPath+"/"+metadataFileName, []byte(metadataContent), 0644)
	instructionsContent := "# Test"
	_ = afero.WriteFile(base, skillPath+"/"+instructionsFileName, []byte(instructionsContent), 0644)
	fs := fsWithStatError(t, base)

	loader := NewLoader(fs)
	skill, err := loader.loadSkill(skillPath)

	require.ErrorIs(t, err, ErrReadFailed)
	assert.Nil(t, skill)
}

func TestLoader_loadSkill_WhenScriptLoadFailure_ThenReturnsReadFailed(t *testing.T) {
	t.Parallel()

	const skillPath = "myskill"
	base := afero.NewMemMapFs()
	_ = base.Mkdir(skillPath, 0755)
	metadataContent := `name: test-skill
description: A test skill`
	_ = afero.WriteFile(base, skillPath+"/"+metadataFileName, []byte(metadataContent), 0644)
	instructionsContent := "# Test"
	_ = afero.WriteFile(base, skillPath+"/"+instructionsFileName, []byte(instructionsContent), 0644)
	_ = base.Mkdir(skillPath+"/scripts", 0755)
	_ = afero.WriteFile(base, skillPath+"/scripts/run.sh", []byte("#!/bin/bash\necho 'test'"), 0755)
	fs := fsWithOpenError(t, base, func(name string) bool {
		if name == skillPath+"/scripts" || name == "./"+skillPath+"/scripts" {
			return false
		}
		return strings.HasPrefix(name, skillPath+"/scripts/") ||
			strings.HasPrefix(name, "./"+skillPath+"/scripts/")
	})

	loader := NewLoader(fs)
	skill, err := loader.loadSkill(skillPath)

	require.ErrorIs(t, err, ErrReadFailed)
	assert.Nil(t, skill)
}

func TestLoader_Load_WhenSingleSkill_ThenLoadsSuccessfully(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	_ = fs.Mkdir("skill1", 0755)
	metadataContent := `name: test-skill
description: A test skill`
	_ = afero.WriteFile(fs, "skill1/"+metadataFileName, []byte(metadataContent), 0644)
	instructionsContent := "# Test Skill\n\nTest instructions."
	_ = afero.WriteFile(fs, "skill1/"+instructionsFileName, []byte(instructionsContent), 0644)

	loader := NewLoader(fs)
	skills, err := loader.Load()

	require.NoError(t, err)
	require.Len(t, skills, 1)
	assert.Equal(t, Name("test-skill"), skills[0].Metadata.Name)
	assert.Equal(t, "A test skill", skills[0].Metadata.Description)
	assert.Equal(t, "# Test Skill\n\nTest instructions.", skills[0].Instructions)
}

func TestLoader_Load_WhenMultipleSkills_ThenLoadsAllSuccessfully(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	_ = fs.Mkdir("skill1", 0755)
	metadata1 := `name: skill-one
description: First skill`
	_ = afero.WriteFile(fs, "skill1/"+metadataFileName, []byte(metadata1), 0644)
	instructions1 := "# Skill One"
	_ = afero.WriteFile(fs, "skill1/"+instructionsFileName, []byte(instructions1), 0644)

	_ = fs.Mkdir("skill2", 0755)
	metadata2 := `name: skill-two
description: Second skill`
	_ = afero.WriteFile(fs, "skill2/"+metadataFileName, []byte(metadata2), 0644)
	instructions2 := "# Skill Two"
	_ = afero.WriteFile(fs, "skill2/"+instructionsFileName, []byte(instructions2), 0644)

	loader := NewLoader(fs)
	skills, err := loader.Load()

	require.NoError(t, err)
	require.Len(t, skills, 2)
	assert.Equal(t, Name("skill-one"), skills[0].Metadata.Name)
	assert.Equal(t, "First skill", skills[0].Metadata.Description)
	assert.Equal(t, Name("skill-two"), skills[1].Metadata.Name)
	assert.Equal(t, "Second skill", skills[1].Metadata.Description)
}

func TestLoader_Load_WhenSkillWithScripts_ThenLoadsScriptsSuccessfully(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	_ = fs.Mkdir("skill1", 0755)
	metadataContent := `name: scripted-skill
description: Skill with scripts`
	_ = afero.WriteFile(fs, "skill1/"+metadataFileName, []byte(metadataContent), 0644)
	instructionsContent := "# Scripted Skill"
	_ = afero.WriteFile(fs, "skill1/"+instructionsFileName, []byte(instructionsContent), 0644)
	_ = fs.Mkdir("skill1/scripts", 0755)
	_ = afero.WriteFile(fs, "skill1/scripts/run.sh", []byte("#!/bin/bash\necho 'running'"), 0755)

	loader := NewLoader(fs)
	skills, err := loader.Load()

	require.NoError(t, err)
	require.Len(t, skills, 1)
	assert.Equal(t, Name("scripted-skill"), skills[0].Metadata.Name)
	require.Len(t, skills[0].Scripts, 1)
	script, exists := skills[0].Scripts["run.sh"]
	require.True(t, exists)
	assert.Equal(t, "application/x-sh", script.ContentType)
	assert.Equal(t, []byte("#!/bin/bash\necho 'running'"), script.Content)
}

func TestLoader_Load(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupFs    func(fs afero.Fs)
		wantErr    error
		errContain string
	}{
		{
			name:    "empty filesystem",
			setupFs: func(fs afero.Fs) {},
			wantErr: ErrNoSkillsFound,
		},
		{
			name: "loadSkill fails",
			setupFs: func(fs afero.Fs) {
				_ = fs.Mkdir("skill1", 0755)
				instructionsContent := "# Test"
				_ = afero.WriteFile(fs, "skill1/"+instructionsFileName, []byte(instructionsContent), 0644)
			},
			wantErr: ErrReadFailed,
		},
		{
			name: "second skill fails",
			setupFs: func(fs afero.Fs) {
				_ = fs.Mkdir("skill1", 0755)
				metadata1 := `name: valid-skill
description: Valid skill`
				_ = afero.WriteFile(fs, "skill1/"+metadataFileName, []byte(metadata1), 0644)
				instructions1 := "# Valid Skill"
				_ = afero.WriteFile(fs, "skill1/"+instructionsFileName, []byte(instructions1), 0644)

				_ = fs.Mkdir("skill2", 0755)
				metadata2 := `name: invalid
description: [invalid: yaml: structure`
				_ = afero.WriteFile(fs, "skill2/"+metadataFileName, []byte(metadata2), 0644)
				instructions2 := "# Invalid"
				_ = afero.WriteFile(fs, "skill2/"+instructionsFileName, []byte(instructions2), 0644)
			},
			wantErr: ErrParseFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fs := afero.NewMemMapFs()
			if tt.setupFs != nil {
				tt.setupFs(fs)
			}

			loader := NewLoader(fs)
			got, err := loader.Load()

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else if tt.errContain != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContain)
			}
			_ = got
		})
	}
}

func TestLoader_Load_WhenResolvePathsError_ThenReturnsError(t *testing.T) {
	t.Parallel()

	fs := fsWithOpenError(t, afero.NewMemMapFs(), func(string) bool { return true })
	loader := NewLoader(fs)

	got, err := loader.Load()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to resolve paths")
	assert.Nil(t, got)
}
