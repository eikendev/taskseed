package config

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
)

func init() {
	yaml.RegisterCustomUnmarshaler(clockTimeUnmarshal)
	yaml.RegisterCustomUnmarshaler(locationUnmarshal)
	yaml.RegisterCustomUnmarshaler(urlUnmarshal)
	yaml.RegisterCustomUnmarshaler(weekdayUnmarshal)
	yaml.RegisterCustomUnmarshaler(scheduleKindUnmarshal)
}

func unmarshalStringInto[T any](target *T, data []byte, parse func(string) (*T, error)) error {
	var raw string
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("decode string: %w", err)
	}

	raw = strings.TrimSpace(raw)

	parsed, err := parse(raw)
	if err != nil {
		return err
	}

	*target = *parsed
	return nil
}

func clockTimeUnmarshal(ct *ClockTime, data []byte) error {
	return unmarshalStringInto(ct, data, parseClockTime)
}

func locationUnmarshal(loc *time.Location, data []byte) error {
	return unmarshalStringInto(loc, data, time.LoadLocation)
}

func urlUnmarshal(u *url.URL, data []byte) error {
	return unmarshalStringInto(u, data, url.Parse)
}

func weekdayUnmarshal(w *time.Weekday, data []byte) error {
	return unmarshalStringInto(w, data, parseWeekday)
}

func scheduleKindUnmarshal(kind *ScheduleKind, data []byte) error {
	return unmarshalStringInto(kind, data, parseScheduleKind)
}

func parseClockTime(val string) (*ClockTime, error) {
	t, err := time.Parse("15:04", val)
	if err != nil {
		return &ClockTime{}, err
	}
	return &ClockTime{Hour: t.Hour(), Minute: t.Minute()}, nil
}

var weekdayValues = map[string]time.Weekday{
	"sunday":    time.Sunday,
	"monday":    time.Monday,
	"tuesday":   time.Tuesday,
	"wednesday": time.Wednesday,
	"thursday":  time.Thursday,
	"friday":    time.Friday,
	"saturday":  time.Saturday,
}

func parseWeekday(name string) (*time.Weekday, error) {
	weekday, ok := weekdayValues[strings.ToLower(name)]
	if !ok {
		return pointerTo(time.Weekday(0)), fmt.Errorf("invalid weekday %q", name)
	}
	return new(weekday), nil
}

func parseScheduleKind(name string) (*ScheduleKind, error) {
	kind, err := ScheduleKindString(name)
	if err != nil {
		return new(ScheduleKindWeekly), fmt.Errorf("invalid schedule kind %q", name)
	}
	return new(kind), nil
}

//go:fix inline
func pointerTo[T any](val T) *T {
	return new(val)
}
