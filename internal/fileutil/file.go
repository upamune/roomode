package fileutil

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// Regular expression for valid filenames
	validFilenameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

// IsValidFilename checks if a filename is valid
func IsValidFilename(filename string) bool {
	return validFilenameRegex.MatchString(filename)
}

// GetModesDir returns the directory for storing mode files
// Uses .roo/modes by default
func GetModesDir() (string, error) {
	modesDir := filepath.Join(".roo", "modes")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(modesDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create modes directory: %w", err)
	}

	return modesDir, nil
}

// GetModeFilePath returns the path to a mode file
func GetModeFilePath(slug string) (string, error) {
	modesDir, err := GetModesDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(modesDir, slug+".md"), nil
}

// ListModeFiles returns all Markdown files in the modes directory
func ListModeFiles() ([]string, error) {
	modesDir, err := GetModesDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(modesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read modes directory: %w", err)
	}

	// Filter to only include Markdown files
	var modeFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			modeFiles = append(modeFiles, filepath.Join(modesDir, entry.Name()))
		}
	}

	return modeFiles, nil
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// WriteFile writes content to a file
func WriteFile(path string, content string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
