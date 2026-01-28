package config

//go:generate go run github.com/dmarkham/enumer@v1.6.1 -type=ScheduleKind -trimprefix=ScheduleKind -transform=snake

// ScheduleKind enumerates the supported recurrence schedule types.
type ScheduleKind int

const (
	// ScheduleKindWeekly repeats on specified weekdays.
	ScheduleKindWeekly ScheduleKind = iota
	// ScheduleKindEveryNDays repeats every fixed number of days.
	ScheduleKindEveryNDays
	// ScheduleKindMonthlyDay repeats on specific month days.
	ScheduleKindMonthlyDay
	// ScheduleKindMonthlyNthWeekday repeats on the nth weekday of each month.
	ScheduleKindMonthlyNthWeekday
	// ScheduleKindYearlyDate repeats on a specific date each year.
	ScheduleKindYearlyDate
	// ScheduleKindYearlyNthWeekday repeats on the nth weekday of a given month each year.
	ScheduleKindYearlyNthWeekday
)
