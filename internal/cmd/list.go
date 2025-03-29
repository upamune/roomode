package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"

	"github.com/upamune/roomode/internal/fileutil"
	"github.com/upamune/roomode/internal/mode"
)

// ListCmd is a command to list available custom modes
type ListCmd struct {
	Verbose bool `short:"v" help:"Show detailed information about each mode."`
}

// Run executes the ListCmd
func (cmd *ListCmd) Run() error {
	// 1. Get list of mode files
	files, err := fileutil.ListModeFiles()
	if err != nil {
		return fmt.Errorf("failed to list mode files: %w", err)
	}

	// 2. Handle case when no mode files are found
	if len(files) == 0 {
		log.Info("No custom modes found")
		return nil
	}

	// 3. Display information for each mode file
	log.Info(fmt.Sprintf("Found %d custom modes:", len(files)))

	for i, file := range files {

		base := filepath.Base(file)
		slug := strings.TrimSuffix(base, filepath.Ext(base))

		modeConfig, err := mode.ParseModeFile(file)
		if err != nil {
			log.Error("Failed to parse mode file", "file", file, "error", err)
			continue
		}

		if cmd.Verbose {
			// Detailed display mode
			fmt.Printf("%d. %s (%s)\n", i+1, modeConfig.Name, slug)
			fmt.Printf("   Path: %s\n", file)
			fmt.Printf("   Groups: ")
			for j, group := range modeConfig.GroupsParsed {
				if j > 0 {
					fmt.Print(", ")
				}
				fmt.Print(group.Name)
				if group.Options != nil && group.Options.FileRegex != nil {
					fmt.Printf(" (fileRegex: %s)", *group.Options.FileRegex)
				}
			}
			fmt.Println()
			if modeConfig.CustomInstructions != nil {
				fmt.Printf("   Custom Instructions: %s\n", *modeConfig.CustomInstructions)
			}
			fmt.Println()
		} else {
			// Simple display mode
			fmt.Printf("%d. %s (%s)\n", i+1, modeConfig.Name, slug)
		}
	}

	return nil
}
