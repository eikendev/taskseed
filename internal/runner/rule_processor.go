package runner

import (
	"context"
	"log/slog"
	"slices"
	"time"

	"github.com/eikendev/taskseed/internal/caldav"
	"github.com/eikendev/taskseed/internal/config"
	"github.com/eikendev/taskseed/internal/identity"
	"github.com/eikendev/taskseed/internal/schedule"
	"github.com/eikendev/taskseed/internal/timeutil"
)

// ruleProcessor holds shared dependencies for per-rule processing.
type ruleProcessor struct {
	calendarURL   string
	windowEnd     time.Time
	existingIDs   map[string]struct{}
	openByRule    map[string]bool
	lastOccByRule map[string]*time.Time
	client        *caldav.Client
	opts          Options
	timezone      *time.Location
	due           config.DuePreference
}

// processRule checks gating conditions, selects the next missing occurrence,
// and creates a task for that occurrence when allowed.
func (p ruleProcessor) processRule(ctx context.Context, rule config.Rule) {
	lastOcc := timeutil.FormatDate(p.lastOccByRule[rule.ID])
	slog.Debug("processing rule", "rule", rule.ID, "schedule_kind", rule.Schedule.Kind, "last_occurrence", lastOcc, "has_open_task", p.openByRule[rule.ID])

	if p.openByRule[rule.ID] {
		slog.Info("skipping as rule is gated by open task", "rule", rule.ID)
		return
	}

	candidate, ok := p.nextCandidate(rule, p.lastOccByRule[rule.ID])
	if !ok {
		slog.Info("found no occurrences to create", "rule", rule.ID, "last_occurrence", lastOcc, "window_end", p.windowEnd.Format(timeutil.DateLayout))
		return
	}

	task := buildTask(rule, candidate, p.calendarURL, p.due, p.timezone)

	if p.opts.DryRun {
		slog.Info("skipping task creation in dry run", "rule", rule.ID)
		return
	}

	err := p.client.CreateTask(ctx, task)
	if err != nil {
		slog.Error("failed to create task", "rule", rule.ID, "error", err)
		return
	}

	slog.Info("created task", "rule", rule.ID, "occurrence", task.Occurrence, "id", task.UID)
}

// nextCandidate returns the earliest missing occurrence for a rule within the window.
func (p ruleProcessor) nextCandidate(rule config.Rule, lastOccurrence *time.Time) (time.Time, bool) {
	ruleToday := timeutil.DateAt(time.Now().In(p.timezone))
	ruleEnd := timeutil.DateAt(p.windowEnd.In(p.timezone))
	occurrences := schedule.Occurrences(rule.Schedule, ruleToday, ruleEnd, p.timezone, lastOccurrence)
	slog.Debug("computed occurrences", "rule", rule.ID, "count", len(occurrences))

	slices.SortFunc(occurrences, func(a, b time.Time) int {
		if a.Before(b) {
			return -1
		}
		if a.After(b) {
			return 1
		}
		return 0
	})

	for _, occ := range occurrences {
		if occ.Before(ruleToday) {
			continue
		}
		occStr := occ.Format(timeutil.DateLayout)
		id := identity.InstanceID(p.calendarURL, rule.ID, occStr)
		if _, exists := p.existingIDs[id]; exists {
			continue
		}
		return occ, true
	}

	return time.Time{}, false
}
