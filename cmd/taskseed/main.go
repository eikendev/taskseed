// Package main wires the taskseed CLI.
package main

import (
	"github.com/alecthomas/kong"

	"github.com/eikendev/taskseed/internal/commands"
	"github.com/eikendev/taskseed/internal/logging"
)

type CLI struct {
	Verbose bool                    `name:"verbose" help:"Enable verbose (debug) logging." env:"TASKSEED_VERBOSE"`
	Sync    commands.SyncCommand    `cmd:"" help:"Synchronize tasks with CalDAV." default:"1"`
	Doctor  commands.DoctorCommand  `cmd:"" help:"Validate configuration and connectivity."`
	Version commands.VersionCommand `cmd:"" help:"Show version information."`
}

func main() {
	var cli CLI
	kctx := kong.Parse(&cli,
		kong.Description("taskseed materializes recurring tasks into CalDAV task lists."),
		kong.UsageOnError(),
	)

	logging.Setup(cli.Verbose)

	err := kctx.Run()
	kctx.FatalIfErrorf(err)
}
