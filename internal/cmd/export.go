package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"

	"github.com/upamune/roomode/internal/fileutil"
	"github.com/upamune/roomode/internal/mode"
)

// ExportCmd is a command to export all modes to a .roomodes JSON file
type ExportCmd struct {
	OutputFile *string `arg:"" optional:"" help:"Output file path (default: .roomodes)."`
}

// Run executes the ExportCmd
func (cmd *ExportCmd) Run() error {
	// 1. Get list of mode files
	files, err := fileutil.ListModeFiles()
	if err != nil {
		return fmt.Errorf("failed to list mode files: %w", err)
	}

	// 2. Handle case when no mode files are found
	if len(files) == 0 {
		log.Info("No custom modes found to export")
		return nil
	}

	// 3. Parse and validate each mode file
	log.Info(fmt.Sprintf("Exporting %d custom modes:", len(files)))

	var validModes []*mode.Config
	invalidCount := 0

	for _, file := range files {

		modeConfig, err := mode.ParseModeFile(file)
		if err != nil {
			log.Error("Failed to parse mode file", "file", file, "error", err)
			invalidCount++
			continue
		}

		err = mode.ValidateMode(modeConfig)
		if err != nil {
			log.Error("Invalid mode file", "file", file, "error", err)
			invalidCount++
			continue
		}

		validModes = append(validModes, modeConfig)
	}

	// 4. Determine output file path
	outputPath := ".roomodes"
	if cmd.OutputFile != nil {
		outputPath = *cmd.OutputFile
	}

	// 5. Create data structure for export
	type ExportedMode struct {
		Slug               string        `json:"slug"`
		Name               string        `json:"name"`
		Groups             []interface{} `json:"groups"`
		CustomInstructions *string       `json:"customInstructions,omitempty"`
		RoleDefinition     string        `json:"roleDefinition"`
		Source             string        `json:"source,omitempty"`
	}

	type ExportData struct {
		CustomModes []ExportedMode `json:"customModes"`
	}

	modes := make([]ExportedMode, 0, len(validModes))
	for _, m := range validModes {
		// Convert ParsedGroupEntry to the expected format for TypeScript schema
		formattedGroups := make([]interface{}, 0, len(m.GroupsParsed))
		for _, g := range m.GroupsParsed {
			if g.Options == nil {
				// Simple string format
				formattedGroups = append(formattedGroups, g.Name)
			} else {
				// Array format [string, options]
				formattedGroups = append(formattedGroups, []interface{}{
					g.Name,
					g.Options,
				})
			}
		}

		modes = append(modes, ExportedMode{
			Slug:               m.Slug,
			Name:               m.Name,
			Groups:             formattedGroups,
			CustomInstructions: m.CustomInstructions,
			RoleDefinition:     m.RoleDefinition,
			Source:             m.Source,
		})
	}

	// 6. Create the final export data structure and convert to JSON
	exportData := ExportData{
		CustomModes: modes,
	}
	jsonData, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// 7. Write to file
	outputDir := filepath.Dir(outputPath)
	if outputDir != "." {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	if err := os.WriteFile(outputPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	// 8. Display results
	log.Info(fmt.Sprintf("Export complete: %d modes exported to %s", len(validModes), outputPath))
	if invalidCount > 0 {
		log.Warn(fmt.Sprintf("%d invalid modes were skipped", invalidCount))
	}

	return nil
}
