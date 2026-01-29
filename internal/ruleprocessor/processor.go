// Package ruleprocessor coordinates rule evaluation and task creation.
package ruleprocessor

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

// Processor manages rule evaluation state and creates tasks when needed.
type Processor struct {
	calendarURL   string
	windowEnd     time.Time
	existingIDs   map[string]struct{}
	openByRule    map[string]bool
	lastOccByRule map[string]*time.Time
	client        *caldav.Client
	dryRun        bool
	timezone      *time.Location
	due           config.DuePreference
}

// New constructs a Processor using the provided configuration and client.
func New(cfg config.Config, client *caldav.Client, dryRun bool) *Processor {
	timezone := cfg.Defaults.Timezone
	if timezone == nil {
		timezone = time.UTC
	}

	today := timeutil.DateAt(time.Now().In(timezone))
	windowEnd := today.AddDate(0, 0, cfg.Sync.HorizonDays)
	normalizedEnd := timeutil.DateAt(windowEnd.In(timezone))

	return &Processor{
		calendarURL:   cfg.Target.URL.String(),
		windowEnd:     normalizedEnd,
		existingIDs:   make(map[string]struct{}),
		openByRule:    make(map[string]bool),
		lastOccByRule: make(map[string]*time.Time),
		client:        client,
		dryRun:        dryRun,
		timezone:      timezone,
		due:           cfg.Defaults.Due,
	}
}

// LoadExisting summarizes the current task list for rule evaluation.
func (p *Processor) LoadExisting(tasks []caldav.Task) {
	p.existingIDs, p.openByRule, p.lastOccByRule = summarize(tasks, p.timezone)
	slog.Debug("summarized existing tasks", "instances", len(p.existingIDs), "rules_with_open", len(p.openByRule), "rules_with_occurrence", len(p.lastOccByRule))
}

// WindowEnd returns the last day included in the rule evaluation window.
func (p *Processor) WindowEnd() time.Time {
	return p.windowEnd
}

// Timezone returns the location used for date calculations.
func (p *Processor) Timezone() *time.Location {
	return p.timezone
}

// ProcessRule evaluates a rule and creates a task if needed.
func (p *Processor) ProcessRule(ctx context.Context, rule config.Rule) {
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

	if p.dryRun {
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

func (p *Processor) nextCandidate(rule config.Rule, lastOccurrence *time.Time) (time.Time, bool) {
	ruleToday := timeutil.DateAt(time.Now().In(p.timezone))
	ruleEnd := p.windowEnd
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

func summarize(tasks []caldav.Task, timezone *time.Location) (map[string]struct{}, map[string]bool, map[string]*time.Time) {
	ids := make(map[string]struct{})
	open := make(map[string]bool)
	lastOcc := make(map[string]*time.Time)

	for _, t := range tasks {
		if t.InstanceID == "" || t.RuleID == "" || t.Occurrence == "" {
			slog.Warn("missing required fields on task", "instance_id", t.InstanceID, "rule_id", t.RuleID, "occurrence", t.Occurrence)
			continue
		}

		parsed, err := time.ParseInLocation(timeutil.DateLayout, t.Occurrence, timezone)
		if err != nil {
			slog.Warn("found invalid occurrence date on task", "rule", t.RuleID, "occurrence", t.Occurrence, "error", err)
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

func buildTask(rule config.Rule, occ time.Time, calendarURL string, due config.DuePreference, timezone *time.Location) caldav.NewTask {
	id := identity.InstanceID(calendarURL, rule.ID, occ.Format(timeutil.DateLayout))
	dueTime := computeDue(occ, due, timezone)

	return caldav.NewTask{
		UID:        id,
		Summary:    rule.Title,
		Notes:      rule.Notes,
		Due:        dueTime,
		DateOnly:   due.DateOnly,
		InstanceID: id,
		RuleID:     rule.ID,
		Occurrence: occ.Format(timeutil.DateLayout),
		Timezone:   timezone.String(),
	}
}

func computeDue(occ time.Time, due config.DuePreference, loc *time.Location) time.Time {
	if loc == nil {
		loc = time.UTC
	}

	if due.DateOnly {
		return time.Date(occ.Year(), occ.Month(), occ.Day(), 0, 0, 0, 0, loc)
	}

	return time.Date(occ.Year(), occ.Month(), occ.Day(), due.Time.Hour, due.Time.Minute, 0, 0, loc)
}
