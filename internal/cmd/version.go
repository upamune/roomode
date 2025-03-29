package cmd

import (
	"fmt"
)

// Version information
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

// VersionCmd is a command to display version information
type VersionCmd struct{}

// Run executes the VersionCmd
func (cmd *VersionCmd) Run() error {
	fmt.Printf("roomode version %s\n", Version)
	fmt.Printf("commit: %s\n", Commit)
	fmt.Printf("build date: %s\n", BuildDate)
	return nil
}
