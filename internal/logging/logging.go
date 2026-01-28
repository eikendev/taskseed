// Package logging configures application logging.
package logging

import (
	"log/slog"
	"os"
)

// Setup installs the global logger with the desired verbosity.
func Setup(verbose bool) {
	level := slog.LevelInfo

	if verbose {
		level = slog.LevelDebug
	}

	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	slog.SetDefault(slog.New(handler))

	if verbose {
		slog.Info("verbose logging enabled", "level", level)
	}
}
