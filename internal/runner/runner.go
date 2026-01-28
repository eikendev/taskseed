// Package runner orchestrates reconciliation between rules and CalDAV.
package runner

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/eikendev/taskseed/internal/caldav"
	"github.com/eikendev/taskseed/internal/config"
	"github.com/eikendev/taskseed/internal/identity"
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
	today := timeutil.DateAt(time.Now().In(cfg.Defaults.Timezone))
	windowStart := today.AddDate(0, 0, -cfg.Sync.LookbackDays)
	windowEnd := today.AddDate(0, 0, cfg.Sync.HorizonDays)

	client, err := caldav.NewClient(cfg.Server.URL.String(), cfg.Target.URL.String(), cfg.Server.Username, cfg.Server.Password)
	if err != nil {
		slog.Error("failed to create caldav client", "error", err)
		return fmt.Errorf("create caldav client: %w", err)
	}

	existing, err := client.QueryTasks(ctx, windowStart, windowEnd)
	if err != nil {
		slog.Error("failed to query existing tasks", "error", err)
		return fmt.Errorf("query existing tasks: %w", err)
	}

	existingIDs, openByRule, lastOccByRule := summarize(existing, cfg.Defaults.Timezone)

	processor := ruleProcessor{
		calendarURL:   cfg.Target.URL.String(),
		windowEnd:     windowEnd,
		existingIDs:   existingIDs,
		openByRule:    openByRule,
		lastOccByRule: lastOccByRule,
		client:        client,
		opts:          opts,
		timezone:      cfg.Defaults.Timezone,
		due:           cfg.Defaults.Due,
	}

	for _, rule := range cfg.Rules {
		processor.processRule(ctx, rule)
	}

	return nil
}

// summarize derives task instance IDs, open-task flags, and the latest occurrence per rule.
// Input: a slice of CalDAV tasks and the timezone used for occurrence normalization.
// Output: existing instance IDs, rule ID -> has-open-task, rule ID -> last occurrence (normalized to midnight in timezone).
func summarize(tasks []caldav.Task, timezone *time.Location) (map[string]struct{}, map[string]bool, map[string]*time.Time) {
	ids := make(map[string]struct{})
	open := make(map[string]bool)
	lastOcc := make(map[string]*time.Time)

	for _, t := range tasks {
		if t.InstanceID == "" || t.RuleID == "" || t.Occurrence == "" {
			slog.Warn("task missing required fields", "instance_id", t.InstanceID, "rule_id", t.RuleID, "occurrence", t.Occurrence)
			continue
		}

		parsed, err := time.ParseInLocation("2006-01-02", t.Occurrence, timezone)
		if err != nil {
			slog.Warn("invalid occurrence date on task", "rule", t.RuleID, "occurrence", t.Occurrence, "error", err)
			continue
		}

		ids[t.InstanceID] = struct{}{}

		if !t.Completed {
			open[t.RuleID] = true
		}

		parsedDate := timeutil.DateAt(parsed.In(timezone))
		if prev, ok := lastOcc[t.RuleID]; !ok || prev == nil || parsedDate.After(*prev) {
			tmp := parsedDate
			lastOcc[t.RuleID] = &tmp
		}
	}

	return ids, open, lastOcc
}

// buildTask constructs a new CalDAV task from a rule and occurrence date.
// Inputs: rule, occurrence date, calendar URL, due preferences, and timezone.
// Output: a populated caldav.NewTask ready to create.
func buildTask(rule config.Rule, occ time.Time, calendarURL string, due config.DuePreference, timezone *time.Location) caldav.NewTask {
	id := identity.InstanceID(calendarURL, rule.ID, occ.Format("2006-01-02"))
	dueTime := computeDue(occ, due, timezone)

	return caldav.NewTask{
		UID:        id,
		Summary:    rule.Title,
		Notes:      rule.Notes,
		Due:        dueTime,
		DateOnly:   due.DateOnly,
		InstanceID: id,
		RuleID:     rule.ID,
		Occurrence: occ.Format("2006-01-02"),
		Timezone:   timezone.String(),
	}
}

// computeDue builds the due datetime for an occurrence using preference settings.
// Inputs: occurrence date, due preferences, and timezone (nil uses UTC).
// Output: a due time with either date-only or time-of-day semantics.
func computeDue(occ time.Time, due config.DuePreference, loc *time.Location) time.Time {
	if loc == nil {
		loc = time.UTC
	}

	if due.DateOnly {
		return time.Date(occ.Year(), occ.Month(), occ.Day(), 0, 0, 0, 0, loc)
	}

	return time.Date(occ.Year(), occ.Month(), occ.Day(), due.Time.Hour, due.Time.Minute, 0, 0, loc)
}
