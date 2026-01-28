package config

import "github.com/go-playground/validator/v10"

var scheduleValidators = map[ScheduleKind]func(validator.StructLevel, RuleSchedule){
	ScheduleKindWeekly:            validateWeeklySchedule,
	ScheduleKindEveryNDays:        validateEveryNDaysSchedule,
	ScheduleKindMonthlyDay:        validateMonthlyDaySchedule,
	ScheduleKindMonthlyNthWeekday: validateMonthlyNthWeekdaySchedule,
	ScheduleKindYearlyDate:        validateYearlyDateSchedule,
	ScheduleKindYearlyNthWeekday:  validateYearlyNthWeekdaySchedule,
}

func validateRuleSchedule(sl validator.StructLevel) {
	schedule, ok := sl.Current().Interface().(RuleSchedule)
	if !ok {
		return
	}
	fn, ok := scheduleValidators[schedule.Kind]
	if !ok {
		sl.ReportError(schedule.Kind, "Kind", "kind", "unknown", "")
		return
	}
	fn(sl, schedule)
}

func validateWeeklySchedule(sl validator.StructLevel, schedule RuleSchedule) {
	if len(schedule.Weekdays) == 0 {
		sl.ReportError(schedule.Weekdays, "Weekdays", "weekdays", "required", "")
		return
	}
}

func validateEveryNDaysSchedule(sl validator.StructLevel, schedule RuleSchedule) {
	if schedule.EveryNDays <= 0 {
		sl.ReportError(schedule.EveryNDays, "EveryNDays", "everyNDays", "gt0", "")
	}
}

func validateMonthlyDaySchedule(sl validator.StructLevel, schedule RuleSchedule) {
	if len(schedule.MonthDays) == 0 {
		sl.ReportError(schedule.MonthDays, "MonthDays", "monthDays", "required", "")
	}
}

func validateMonthlyNthWeekdaySchedule(sl validator.StructLevel, schedule RuleSchedule) {
	if schedule.Nth <= 0 {
		sl.ReportError(schedule.Nth, "Nth", "nth", "gt0", "")
	}
}

func validateYearlyDateSchedule(sl validator.StructLevel, schedule RuleSchedule) {
	if schedule.Month < 1 || schedule.Month > 12 {
		sl.ReportError(schedule.Month, "Month", "month", "month", "")
	}
	if schedule.Day < 1 || schedule.Day > 31 {
		sl.ReportError(schedule.Day, "Day", "day", "day", "")
	}
}

func validateYearlyNthWeekdaySchedule(sl validator.StructLevel, schedule RuleSchedule) {
	if schedule.Month < 1 || schedule.Month > 12 {
		sl.ReportError(schedule.Month, "Month", "month", "month", "")
	}
	if schedule.Nth <= 0 {
		sl.ReportError(schedule.Nth, "Nth", "nth", "gt0", "")
	}
}
