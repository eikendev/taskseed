// Package commands defines CLI subcommands.
package commands

import (
	"fmt"
	"log/slog"
	"runtime/debug"
)

// VersionCommand prints version information.
type VersionCommand struct{}

// Run executes the version command.
func (*VersionCommand) Run() error {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		slog.Error("failed to read build info")
		return fmt.Errorf("build info not available")
	}

	fmt.Printf("taskseed %s\n", buildInfo.Main.Version)

	return nil
}
