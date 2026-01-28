// Package timeutil provides time and date helpers.
package timeutil

import "time"

// DateAt normalizes a time to midnight in its location.
func DateAt(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}
