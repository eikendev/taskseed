package commands

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/eikendev/taskseed/internal/config"
	"github.com/eikendev/taskseed/internal/runner"
)

// SyncCommand reconciles tasks against CalDAV.
type SyncCommand struct {
	Config       string `name:"config" short:"c" help:"Path to configuration file." default:"config.yaml" env:"TASKSEED_CONFIG"`
	DryRun       bool   `name:"dry-run" help:"Print planned tasks without creating them." env:"TASKSEED_DRY_RUN"`
	HorizonDays  int    `name:"horizon-days" help:"Override planning horizon in days." default:"-1" env:"TASKSEED_HORIZON_DAYS"`
	LookbackDays int    `name:"lookback-days" help:"Override lookback window in days." default:"-1" env:"TASKSEED_LOOKBACK_DAYS"`
}

// Run executes the sync command.
func (cmd *SyncCommand) Run() error {
	cfg, err := config.Load(cmd.Config)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		return fmt.Errorf("load config: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if err := runner.Run(ctx, cfg, runner.Options{
		DryRun: cmd.DryRun,
	}); err != nil {
		slog.Error("sync failed", "error", err)
		return fmt.Errorf("sync failed: %w", err)
	}

	return nil
}
