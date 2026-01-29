package schedule

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eikendev/taskseed/internal/config"
)

func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func TestOccurrencesWeeklyReturnsMatchingDates(t *testing.T) {
	def := config.RuleSchedule{Kind: config.ScheduleKindWeekly, Weekdays: []time.Weekday{time.Monday, time.Wednesday}}
	start := date(2023, time.January, 1)
	end := date(2023, time.January, 8)
	expected := []time.Time{date(2023, time.January, 2), date(2023, time.January, 4)}

	got := Occurrences(def, start, end, time.UTC, nil)

	require.Len(t, got, len(expected))
	assert.Equal(t, expected, got)
}

func TestOccurrencesUnknownKindReturnsNil(t *testing.T) {
	def := config.RuleSchedule{Kind: config.ScheduleKind(999)}
	start := date(2023, time.January, 1)
	end := date(2023, time.January, 8)

	got := Occurrences(def, start, end, time.UTC, nil)

	assert.Nil(t, got)
}

func TestWeeklyMatchesWeekdays(t *testing.T) {
	weekdays := []time.Weekday{time.Tuesday}
	start := date(2023, time.January, 1)
	end := date(2023, time.January, 7)
	expected := []time.Time{date(2023, time.January, 3)}

	got := weekly(weekdays, start, end)

	require.Len(t, got, len(expected))
	assert.Equal(t, expected, got)
}

func TestWeeklyEmptyWeekdaysReturnsEmpty(t *testing.T) {
	weekdays := []time.Weekday{}
	start := date(2023, time.January, 1)
	end := date(2023, time.January, 7)

	got := weekly(weekdays, start, end)

	assert.Nil(t, got)
}

func TestEveryNDaysReturnsIntervalDates(t *testing.T) {
	interval := 2
	start := date(2023, time.January, 1)
	end := date(2023, time.January, 7)
	expected := []time.Time{date(2023, time.January, 1), date(2023, time.January, 3), date(2023, time.January, 5), date(2023, time.January, 7)}

	got := everyNDays(interval, start, end, nil)

	require.Len(t, got, len(expected))
	assert.Equal(t, expected, got)
}

func TestEveryNDaysInvalidIntervalReturnsNil(t *testing.T) {
	interval := 0
	start := date(2023, time.January, 1)
	end := date(2023, time.January, 7)

	got := everyNDays(interval, start, end, nil)

	assert.Nil(t, got)
}

func TestMonthlyDayReturnsValidDates(t *testing.T) {
	days := []int{1, 15}
	start := date(2023, time.January, 10)
	end := date(2023, time.March, 20)
	expected := []time.Time{
		date(2023, time.January, 15),
		date(2023, time.February, 1),
		date(2023, time.February, 15),
		date(2023, time.March, 1),
		date(2023, time.March, 15),
	}

	got := monthlyDay(days, start, end)

	require.Len(t, got, len(expected))
	assert.Equal(t, expected, got)
}

func TestMonthlyDayIgnoresInvalidDays(t *testing.T) {
	days := []int{32}
	start := date(2023, time.January, 1)
	end := date(2023, time.March, 31)

	got := monthlyDay(days, start, end)

	assert.Nil(t, got)
}

func TestMonthlyNthWeekdayReturnsExpectedDates(t *testing.T) {
	nth := 2
	weekday := time.Monday
	start := date(2023, time.January, 1)
	end := date(2023, time.March, 31)
	expected := []time.Time{
		date(2023, time.January, 9),
		date(2023, time.February, 13),
		date(2023, time.March, 13),
	}

	got := monthlyNthWeekday(nth, weekday, start, end)

	require.Len(t, got, len(expected))
	assert.Equal(t, expected, got)
}

func TestMonthlyNthWeekdayOutOfRangeReturnsEmpty(t *testing.T) {
	nth := 6
	weekday := time.Monday
	start := date(2023, time.January, 1)
	end := date(2023, time.March, 31)

	got := monthlyNthWeekday(nth, weekday, start, end)

	assert.Nil(t, got)
}

func TestYearlyDateReturnsExpectedDates(t *testing.T) {
	month := 2
	day := 10
	start := date(2022, time.January, 1)
	end := date(2024, time.December, 31)
	expected := []time.Time{
		date(2022, time.February, 10),
		date(2023, time.February, 10),
		date(2024, time.February, 10),
	}

	got := yearlyDate(month, day, start, end)

	require.Len(t, got, len(expected))
	assert.Equal(t, expected, got)
}

func TestYearlyDateInvalidDayReturnsEmpty(t *testing.T) {
	month := 2
	day := 31
	start := date(2022, time.January, 1)
	end := date(2024, time.December, 31)

	got := yearlyDate(month, day, start, end)

	assert.Nil(t, got)
}

func TestYearlyNthWeekdayReturnsExpectedDates(t *testing.T) {
	month := 5
	nth := 1
	weekday := time.Monday
	start := date(2022, time.January, 1)
	end := date(2024, time.December, 31)
	expected := []time.Time{
		date(2022, time.May, 2),
		date(2023, time.May, 1),
		date(2024, time.May, 6),
	}

	got := yearlyNthWeekday(month, nth, weekday, start, end)

	require.Len(t, got, len(expected))
	assert.Equal(t, expected, got)
}

func TestYearlyNthWeekdayOutOfRangeReturnsEmpty(t *testing.T) {
	month := 5
	nth := 6
	weekday := time.Monday
	start := date(2022, time.January, 1)
	end := date(2024, time.December, 31)

	got := yearlyNthWeekday(month, nth, weekday, start, end)

	assert.Nil(t, got)
}
