package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"

	"github.com/upamune/roomode/internal/config"
	"github.com/upamune/roomode/internal/editor"
	"github.com/upamune/roomode/internal/fileutil"
)

// CreateCmd is a command to create a new custom mode Markdown file
type CreateCmd struct {
	Slug string `arg:"" help:"Slug for the custom mode (used as filename)."`
	Name string `arg:"" optional:"" help:"Name for the custom mode (default: same as slug)."`
}

// Run executes the CreateCmd
func (cmd *CreateCmd) Run() error {
	// 1. Validate slug
	if !fileutil.IsValidFilename(cmd.Slug) {
		return fmt.Errorf("invalid slug: %s (contains invalid characters)", cmd.Slug)
	}

	// 2. Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// 3. Ensure modes directory exists
	if err := config.EnsureModesDir(cfg); err != nil {
		return err
	}

	// 4. Build file path
	filePath := filepath.Join(cfg.ModesDir, cmd.Slug+".md")

	// 5. Check if file already exists
	if fileutil.FileExists(filePath) {
		return fmt.Errorf("mode file already exists: %s", filePath)
	}

	// 6. Use slug as name if not specified
	name := cmd.Name
	if name == "" {
		name = cmd.Slug
	}

	// 7. Create template
	template := editor.GetModeTemplate(name)

	// 8. Confirm with user
	var confirmed bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("Create new mode '%s' at %s?", name, filePath)).
				Value(&confirmed),
		),
	)

	err = form.Run()
	if err != nil {
		return fmt.Errorf("form error: %w", err)
	}

	if !confirmed {
		log.Info("Operation cancelled")
		return nil
	}

	// 9. Create template file and open in editor
	if err := editor.CreateTemplateFile(filePath, template); err != nil {
		return err
	}

	log.Info("Mode file created and opened in editor", "path", filePath)
	return nil
}
