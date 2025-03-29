package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"

	"github.com/upamune/roomode/internal/fileutil"
)

// ImportedMode represents the structure of a mode in the .roomodes JSON file
type ImportedMode struct {
	Slug               string        `json:"slug"`
	Name               string        `json:"name"`
	Groups             []interface{} `json:"groups"` // Using interface{} because the format might be different
	CustomInstructions *string       `json:"customInstructions,omitempty"`
	RoleDefinition     string        `json:"roleDefinition"`
	Source             string        `json:"source,omitempty"`
}

// RoomodesFile represents the structure of the .roomodes JSON file
type RoomodesFile struct {
	CustomModes []ImportedMode `json:"customModes"`
}

// ImportCmd is a command to import modes from a .roomodes JSON file into the .roo/modes directory
type ImportCmd struct {
	InputFile *string `arg:"" optional:"" help:"Input JSON file path (default: .roomodes)."`
	Force     bool    `help:"Overwrite existing mode files without confirmation." default:"false"`
}

// Run executes the ImportCmd
func (cmd *ImportCmd) Run() error {
	// 1. Determine input file path
	inputPath := ".roomodes"
	if cmd.InputFile != nil {
		inputPath = *cmd.InputFile
	}

	// 2. Read and parse the JSON file
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Try to parse as a RoomodesFile first (new format)
	var roomodesFile RoomodesFile
	if err := json.Unmarshal(data, &roomodesFile); err != nil {
		// If that fails, try to parse as a direct array of modes (old format)
		var modes []ImportedMode
		if err := json.Unmarshal(data, &modes); err != nil {
			return fmt.Errorf("failed to parse JSON: %w", err)
		}
		// If successful, use the direct array
		roomodesFile.CustomModes = modes
	}

	// 3. Create the modes directory
	modesDir, err := fileutil.GetModesDir()
	if err != nil {
		return fmt.Errorf("failed to create modes directory: %w", err)
	}

	// 4. Process each mode
	imported := 0
	skipped := 0

	for _, mode := range roomodesFile.CustomModes {
		// Validate mode data
		if mode.Slug == "" || mode.Name == "" || mode.RoleDefinition == "" {
			log.Warn("Skipping invalid mode", "slug", mode.Slug)
			skipped++
			continue
		}

		// Generate file path
		filePath := filepath.Join(modesDir, mode.Slug+".md")

		// Check if file exists
		if fileutil.FileExists(filePath) && !cmd.Force {
			// Prompt for confirmation
			fmt.Printf("File exists: %s. Overwrite? [y/N]: ", filePath)

			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(response)

			if response != "y" && response != "Y" {
				log.Info("Skipping existing file", "file", filePath)
				skipped++
				continue
			}
		} else if fileutil.FileExists(filePath) {
			log.Info("Overwriting existing file", "file", filePath)
		}

		// Generate markdown content
		content, err := GenerateModeMarkdown(mode)
		if err != nil {
			log.Error("Failed to generate markdown", "slug", mode.Slug, "error", err)
			skipped++
			continue
		}

		// Write to file
		if err := fileutil.WriteFile(filePath, content); err != nil {
			log.Error("Failed to write file", "file", filePath, "error", err)
			skipped++
			continue
		}

		log.Info("Imported mode", "slug", mode.Slug, "file", filePath)
		imported++
	}

	// 5. Display summary
	log.Info(fmt.Sprintf("Import complete: %d modes imported, %d skipped", imported, skipped))

	return nil
}

// GenerateModeMarkdown creates markdown content with frontmatter from an imported mode
func GenerateModeMarkdown(mode ImportedMode) (string, error) {
	var sb strings.Builder

	// Start frontmatter
	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("name: %s\n", mode.Name))

	// Handle groups
	sb.WriteString("groups:\n")
	for _, groupInterface := range mode.Groups {
		switch group := groupInterface.(type) {
		case string:
			// Simple string group
			sb.WriteString(fmt.Sprintf("  - %s\n", group))
		case map[string]interface{}:
			// Handle map representation of a group
			name, hasName := group["Name"].(string)
			if !hasName {
				name, hasName = group["name"].(string)
			}

			if !hasName {
				// Fallback if we can't find a name
				sb.WriteString(fmt.Sprintf("  - %v\n", groupInterface))
				continue
			}

			// Check for options
			options, hasOptions := group["Options"].(map[string]interface{})
			if !hasOptions {
				options, hasOptions = group["options"].(map[string]interface{})
			}

			if !hasOptions || len(options) == 0 {
				// Simple group entry
				sb.WriteString(fmt.Sprintf("  - %s\n", name))
			} else {
				// Complex group entry with options
				sb.WriteString(fmt.Sprintf("  - [%s, {", name))

				first := true
				if fileRegex, ok := options["FileRegex"]; ok && fileRegex != nil {
					sb.WriteString(fmt.Sprintf("fileRegex: \"%v\"", fileRegex))
					first = false
				} else if fileRegex, ok := options["fileRegex"]; ok && fileRegex != nil {
					sb.WriteString(fmt.Sprintf("fileRegex: \"%v\"", fileRegex))
					first = false
				}

				if description, ok := options["Description"]; ok && description != nil {
					if !first {
						sb.WriteString(", ")
					}
					sb.WriteString(fmt.Sprintf("description: \"%v\"", description))
				} else if description, ok := options["description"]; ok && description != nil {
					if !first {
						sb.WriteString(", ")
					}
					sb.WriteString(fmt.Sprintf("description: \"%v\"", description))
				}

				sb.WriteString("}]\n")
			}
		default:
			// Just convert to string as fallback
			sb.WriteString(fmt.Sprintf("  - %v\n", groupInterface))
		}
	}

	// Add role definition to frontmatter
	sb.WriteString(fmt.Sprintf("roleDefinition: |\n  %s\n", strings.ReplaceAll(mode.RoleDefinition, "\n", "\n  ")))

	// End frontmatter
	sb.WriteString("---\n\n")

	// Add custom instructions to the body if present
	if mode.CustomInstructions != nil {
		sb.WriteString(*mode.CustomInstructions)
	}

	return sb.String(), nil
}
