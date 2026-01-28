package timeutil

import "time"

// DaysBetween returns the number of whole days between start and end.
func DaysBetween(start, end time.Time) int {
	if end.Before(start) {
		return -DaysBetween(end, start)
	}

	days := 0
	for t := start; t.Before(end); t = t.AddDate(0, 0, 1) {
		days++
	}

	return days
}
