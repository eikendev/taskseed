// Package schedule expands recurrence rules into dates.
package schedule

import (
	"slices"
	"time"

	"github.com/eikendev/taskseed/internal/config"
	"github.com/eikendev/taskseed/internal/timeutil"
)

// Occurrences returns all occurrence dates for a rule between startDate and endDate.
func Occurrences(def config.RuleSchedule, startDate, endDate time.Time, tz *time.Location, anchor *time.Time) []time.Time {
	if tz == nil {
		tz = time.Local
	}

	start := timeutil.DateAt(startDate.In(tz))
	end := timeutil.DateAt(endDate.In(tz))

	switch def.Kind {
	case config.ScheduleKindWeekly:
		return weekly(def.Weekdays, start, end)
	case config.ScheduleKindEveryNDays:
		return everyNDays(def.EveryNDays, start, end, anchor)
	case config.ScheduleKindMonthlyDay:
		return monthlyDay(def.MonthDays, start, end)
	case config.ScheduleKindMonthlyNthWeekday:
		return monthlyNthWeekday(def.Nth, def.NthWeekday, start, end)
	case config.ScheduleKindYearlyDate:
		return yearlyDate(def.Month, def.Day, start, end)
	case config.ScheduleKindYearlyNthWeekday:
		return yearlyNthWeekday(def.Month, def.Nth, def.YearlyNthWeekday, start, end)
	default:
		return nil
	}
}

func weekly(weekdays []time.Weekday, start, end time.Time) []time.Time {
	targets := make(map[time.Weekday]struct{})

	for _, day := range weekdays {
		targets[day] = struct{}{}
	}

	var out []time.Time
	for t := start; !t.After(end); t = t.AddDate(0, 0, 1) {
		if _, ok := targets[t.Weekday()]; ok {
			out = append(out, t)
		}
	}

	return out
}

func everyNDays(interval int, start, end time.Time, anchor *time.Time) []time.Time {
	if interval <= 0 {
		return nil
	}

	anchorDate := start
	if anchor != nil {
		anchorDate = timeutil.DateAt(anchor.In(start.Location()))
	}

	start = timeutil.DateAt(start)
	end = timeutil.DateAt(end)

	var out []time.Time
	first := anchorDate

	if anchorDate.Before(start) {
		days := timeutil.DaysBetween(anchorDate, start)
		skip := days / interval
		first = anchorDate.AddDate(0, 0, skip*interval)
		if first.Before(start) {
			first = first.AddDate(0, 0, interval)
		}
	}

	for t := first; !t.After(end); t = t.AddDate(0, 0, interval) {
		out = append(out, t)
	}

	return out
}

func monthlyDay(days []int, start, end time.Time) []time.Time {
	slices.Sort(days)
	var out []time.Time

	for t := start; !t.After(end); t = t.AddDate(0, 1, 0) {
		y, m, _ := t.Date()

		for _, d := range days {
			if d < 1 || d > 31 {
				continue
			}

			date := time.Date(y, m, d, 0, 0, 0, 0, t.Location())
			if date.Month() != m {
				continue
			}
			if date.Before(start) || date.After(end) {
				continue
			}

			out = append(out, date)
		}
	}

	return out
}

func monthlyNthWeekday(n int, weekday time.Weekday, start, end time.Time) []time.Time {
	var out []time.Time

	for t := start; !t.After(end); t = t.AddDate(0, 1, 0) {
		y, m, _ := t.Date()
		firstOfMonth := time.Date(y, m, 1, 0, 0, 0, 0, t.Location())
		offset := (int(weekday) - int(firstOfMonth.Weekday()) + 7) % 7
		day := 1 + offset + (n-1)*7

		date := time.Date(y, m, day, 0, 0, 0, 0, t.Location())
		if date.Month() != m {
			continue
		}
		if date.Before(start) || date.After(end) {
			continue
		}

		out = append(out, date)
	}

	return out
}

func yearlyDate(month int, day int, start, end time.Time) []time.Time {
	var out []time.Time

	for y := start.Year(); y <= end.Year(); y++ {
		date := time.Date(y, time.Month(month), day, 0, 0, 0, 0, start.Location())
		if date.Month() != time.Month(month) || date.Before(start) || date.After(end) {
			continue
		}

		out = append(out, date)
	}

	return out
}

func yearlyNthWeekday(month int, n int, weekday time.Weekday, start, end time.Time) []time.Time {
	var out []time.Time

	for y := start.Year(); y <= end.Year(); y++ {
		firstOfMonth := time.Date(y, time.Month(month), 1, 0, 0, 0, 0, start.Location())
		offset := (int(weekday) - int(firstOfMonth.Weekday()) + 7) % 7
		day := 1 + offset + (n-1)*7
		date := time.Date(y, time.Month(month), day, 0, 0, 0, 0, start.Location())

		if date.Month() != time.Month(month) || date.Before(start) || date.After(end) {
			continue
		}

		out = append(out, date)
	}

	return out
}
