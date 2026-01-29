// Package timeutil provides time and date helpers.
package timeutil

import "time"

// DateLayout matches the canonical date format used in configuration files.
const DateLayout = "2006-01-02"

// DateAt normalizes a time to midnight in its location.
func DateAt(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

// FormatDate returns a formatted date or an empty string for nil.
func FormatDate(date *time.Time) string {
	if date == nil {
		return ""
	}
	return date.Format(DateLayout)
}
