package editor

import (
	"fmt"
	"os"
	"os/exec"
)

// GetPreferredEditor returns the user's preferred editor
// Checks the EDITOR or VISUAL environment variables, uses default if neither is set
func GetPreferredEditor() string {
	// Check EDITOR environment variable
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}

	// Check VISUAL environment variable
	if visual := os.Getenv("VISUAL"); visual != "" {
		return visual
	}

	// Default editor
	// Use "nano" as default on macOS and most Linux systems
	return "nano"
}

// OpenInEditor opens the specified file in an editor
func OpenInEditor(filePath string) error {
	editor := GetPreferredEditor()

	cmd := exec.Command(editor, filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to open editor: %w", err)
	}

	return nil
}

// CreateTemplateFile creates a template file and opens it in an editor
func CreateTemplateFile(filePath string, template string) error {
	if err := os.WriteFile(filePath, []byte(template), 0644); err != nil {
		return fmt.Errorf("failed to create template file: %w", err)
	}

	return OpenInEditor(filePath)
}

// GetModeTemplate returns a template for mode files
func GetModeTemplate(name string) string {
	return fmt.Sprintf(`---
name: %s
groups:
  - read
  - edit
customInstructions: null
---

# %s

This mode is... (describe the mode here)

## Usage Examples

- Example 1: ...
- Example 2: ...

## Limitations

- Limitation 1: ...
- Limitation 2: ...
`, name, name)
}
