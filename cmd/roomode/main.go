package main

import (
	"os"

	"github.com/alecthomas/kong"
	"github.com/charmbracelet/log"

	"github.com/upamune/roomode/internal/cmd"
)

var cli struct {
	Create  cmd.CreateCmd  `cmd:"" help:"Create a new custom mode markdown file."`
	List    cmd.ListCmd    `cmd:"" help:"List available custom modes."`
	Export  cmd.ExportCmd  `cmd:"" help:"Export all modes to a .roomodes JSON file."`
	Import  cmd.ImportCmd  `cmd:"" help:"Import modes from a .roomodes JSON file into the .roo/modes directory."`
	Version cmd.VersionCmd `cmd:"" help:"Show version information."`
}

func main() {
	// Set default logger
	log.SetDefault(log.NewWithOptions(
		os.Stderr,
		log.Options{
			ReportTimestamp: false,
		},
	))

	ctx := kong.Parse(&cli,
		kong.Name("roomode"),
		kong.Description("A CLI tool to manage RooCode custom modes defined in markdown files."),
		kong.UsageOnError(),
	)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
