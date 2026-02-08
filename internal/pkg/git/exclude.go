package git

import (
	"fmt"
	"strings"

	"github.com/spf13/afero"
)

const excludeFilePath = "info/exclude"

func readExcludeLines(gitFs afero.Fs) ([]string, error) {
	exists, err := afero.Exists(gitFs, excludeFilePath)
	if err != nil {
		return nil, fmt.Errorf("exclude file existence check: %w", err)
	}
	if !exists {
		return nil, nil
	}

	content, err := afero.ReadFile(gitFs, excludeFilePath)
	if err != nil {
		return nil, fmt.Errorf("exclude file read: %w", err)
	}

	return strings.Split(string(content), "\n"), nil
}

func matchesPattern(line, pattern string) bool {
	trimmed := strings.TrimRight(line, " \t")
	return trimmed == pattern
}

// IsExcluded checks whether the given pattern exists in .git/info/exclude file.
// Returns false without error if the exclude file does not exist.
func IsExcluded(gitFs afero.Fs, pattern string) (bool, error) {
	lines, err := readExcludeLines(gitFs)
	if err != nil {
		return false, err
	}
	if lines == nil {
		return false, nil
	}

	for _, line := range lines {
		if matchesPattern(line, pattern) {
			return true, nil
		}
	}

	return false, nil
}

// Exclude adds the given pattern to .git/info/exclude file.
// Creates the info directory if it does not exist.
// Operation is idempotent - if pattern already exists, no changes are made.
func Exclude(gitFs afero.Fs, pattern string) error {
	lines, err := readExcludeLines(gitFs)
	if err != nil {
		return err
	}

	for _, line := range lines {
		if matchesPattern(line, pattern) {
			return nil
		}
	}

	if err := gitFs.MkdirAll("info", 0755); err != nil {
		return fmt.Errorf("exclude directory creation: %w", err)
	}

	var content string
	if lines != nil {
		content = strings.Join(lines, "\n")
		if !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
	}
	content += pattern + "\n"

	if err := afero.WriteFile(gitFs, excludeFilePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("exclude file write: %w", err)
	}

	return nil
}

// Unexclude removes all occurrences of the given pattern from .git/info/exclude file.
// Returns nil without error if the exclude file does not exist or pattern is not found.
func Unexclude(gitFs afero.Fs, pattern string) error {
	lines, err := readExcludeLines(gitFs)
	if err != nil {
		return err
	}
	if lines == nil {
		return nil
	}

	var filtered []string
	removed := false
	for _, line := range lines {
		if matchesPattern(line, pattern) {
			removed = true
			continue
		}
		filtered = append(filtered, line)
	}

	if !removed {
		return nil
	}

	content := strings.Join(filtered, "\n")
	if err := afero.WriteFile(gitFs, excludeFilePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("exclude file write: %w", err)
	}

	return nil
}
