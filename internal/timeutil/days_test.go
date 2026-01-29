package timeutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDaysBetweenCountsForward(t *testing.T) {
	start := time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2023, time.January, 4, 0, 0, 0, 0, time.UTC)
	expected := 3

	got := DaysBetween(start, end)

	assert.Equal(t, expected, got)
}

func TestDaysBetweenReturnsNegativeWhenReversed(t *testing.T) {
	start := time.Date(2023, time.January, 4, 0, 0, 0, 0, time.UTC)
	end := time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC)
	expected := -3

	got := DaysBetween(start, end)

	assert.Equal(t, expected, got)
}
