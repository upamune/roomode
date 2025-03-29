package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"gopkg.in/yaml.v3"

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
	// Create frontmatter data structure
	frontmatterData := map[string]interface{}{
		"name":          mode.Name,
		"roleDefinition": mode.RoleDefinition,
	}

	// Process groups to ensure proper YAML formatting
	processedGroups := make([]interface{}, 0, len(mode.Groups))
	for _, groupInterface := range mode.Groups {
		switch group := groupInterface.(type) {
		case string:
			// Simple string group
			processedGroups = append(processedGroups, group)
		case []interface{}:
			// Array format [string, options]
			if len(group) == 2 {
				name, ok := group[0].(string)
				if ok {
					switch options := group[1].(type) {
					case map[string]interface{}:
						// Convert array format to map format
						groupMap := map[string]interface{}{
							name: options,
						}
						processedGroups = append(processedGroups, groupMap)
					case map[interface{}]interface{}:
						// Convert interface{} keys to string keys
						stringOptions := make(map[string]interface{})
						for k, v := range options {
							if keyStr, ok := k.(string); ok {
								stringOptions[keyStr] = v
							}
						}
						groupMap := map[string]interface{}{
							name: stringOptions,
						}
						processedGroups = append(processedGroups, groupMap)
					default:
						// Just add the name as a simple string
						processedGroups = append(processedGroups, name)
					}
				} else {
					// Just add as is if not properly structured
					processedGroups = append(processedGroups, groupInterface)
				}
			} else {
				// Just add as is if not properly structured
				processedGroups = append(processedGroups, groupInterface)
			}
		case map[string]interface{}:
			// Map format - already structured correctly
			processedGroups = append(processedGroups, group)
		default:
			// Just add as is for any other type
			processedGroups = append(processedGroups, groupInterface)
		}
	}

	// Add processed groups to frontmatter
	frontmatterData["groups"] = processedGroups

	// Generate YAML frontmatter manually since the frontmatter package doesn't provide a Marshal function
	var buf bytes.Buffer
	
	// Start frontmatter
	buf.WriteString("---\n")
	
	// Marshal to YAML
	yamlData, err := yaml.Marshal(frontmatterData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal frontmatter: %w", err)
	}
	
	// Write YAML content
	buf.Write(yamlData)
	
	// End frontmatter
	buf.WriteString("---\n")

	// Add custom instructions to the body if present
	if mode.CustomInstructions != nil {
		buf.WriteString("\n")
		buf.WriteString(*mode.CustomInstructions)
	}

	return buf.String(), nil
}
