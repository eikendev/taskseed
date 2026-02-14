// Package runner orchestrates reconciliation between rules and CalDAV.
package runner

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/eikendev/taskseed/internal/caldav"
	"github.com/eikendev/taskseed/internal/config"
	"github.com/eikendev/taskseed/internal/ruleprocessor"
	"github.com/eikendev/taskseed/internal/timeutil"
)

// Options control synchronization behavior.
type Options struct {
	DryRun bool
}

// Run performs a reconciliation cycle that reads existing tasks, checks rules,
// and creates at most one new task per rule when needed.
// Inputs: context for cancellation, a validated config, and runtime options.
// Output: error when client setup or query fails; per-rule creation errors are logged.
func Run(ctx context.Context, cfg config.Config, opts Options) error {
	client, err := caldav.NewClient(cfg.Server.URL.String(), cfg.Target.URL.String(), cfg.Server.Username, cfg.Server.Password)
	if err != nil {
		slog.Error("failed to create caldav client", "error", err)
		return fmt.Errorf("create caldav client: %w", err)
	}

	processor := ruleprocessor.New(cfg, client, opts.DryRun)

	today := timeutil.DateAt(time.Now().In(processor.Timezone()))
	windowStart := today.AddDate(0, 0, -cfg.Sync.LookbackDays)
	windowEnd := processor.WindowEnd()

	slog.Debug("querying existing tasks")
	existing, err := client.QueryTasks(ctx, windowStart, windowEnd)
	if err != nil {
		slog.Error("failed to query existing tasks", "error", err)
		return fmt.Errorf("query existing tasks: %w", err)
	}

	slog.Info("fetched existing tasks")

	processor.LoadExisting(existing)

	for _, rule := range cfg.Rules {
		processor.ProcessRule(ctx, rule)
	}

	return nil
}
