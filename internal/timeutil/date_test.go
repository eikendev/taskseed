package timeutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDateAtNormalizesToMidnight(t *testing.T) {
	input := time.Date(2023, time.March, 14, 9, 26, 53, 123, time.UTC)
	expected := time.Date(2023, time.March, 14, 0, 0, 0, 0, time.UTC)

	got := DateAt(input)

	assert.Equal(t, expected, got)
}

func TestDateAtKeepsDateWhenAlreadyMidnight(t *testing.T) {
	input := time.Date(2023, time.July, 2, 0, 0, 0, 0, time.UTC)
	expected := time.Date(2023, time.July, 2, 0, 0, 0, 0, time.UTC)

	got := DateAt(input)

	assert.Equal(t, expected, got)
}

func TestFormatDateReturnsFormattedValue(t *testing.T) {
	value := time.Date(2023, time.January, 5, 10, 0, 0, 0, time.UTC)
	expected := "2023-01-05"

	got := FormatDate(&value)

	assert.Equal(t, expected, got)
}

func TestFormatDateNilReturnsEmpty(t *testing.T) {
	var value *time.Time

	got := FormatDate(value)

	assert.Empty(t, got)
}
