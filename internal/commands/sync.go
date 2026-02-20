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
	Config string `name:"config" short:"c" help:"Path to configuration file." default:"config.yaml" env:"TASKSEED_CONFIG"`
	DryRun bool   `name:"dry-run" help:"Print planned tasks without creating them." env:"TASKSEED_DRY_RUN"`
}

// Run executes the sync command.
func (cmd *SyncCommand) Run() error {
	start := time.Now()
	slog.Info("starting sync", "config", cmd.Config, "dry_run", cmd.DryRun)

	cfg, err := config.Load(cmd.Config)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		return fmt.Errorf("load config: %w", err)
	}

	slog.Debug("loaded config", "rules", len(cfg.Rules), "timezone", cfg.Defaults.Timezone.String(), "horizon_days", cfg.Sync.HorizonDays, "lookback_days", cfg.Sync.LookbackDays)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if err := runner.Run(ctx, cfg, runner.Options{
		DryRun: cmd.DryRun,
	}); err != nil {
		slog.Error("failed to sync", "error", err)
		return fmt.Errorf("sync failed: %w", err)
	}

	slog.Info("completed sync", "duration", time.Since(start).String(), "dry_run", cmd.DryRun)

	return nil
}
