// Package config loads and validates taskseed configuration.
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-yaml"
)

//gosec:disable G101 -- These are environment variable names, not credentials
const (
	envUsernameVar = "TASKSEED_CALDAV_USERNAME"
	envPasswordVar = "TASKSEED_CALDAV_PASSWORD"
)

// Config mirrors the user-provided YAML structure.
type Config struct {
	Server   ServerConfig   `yaml:"server" validate:"required"`
	Target   TargetConfig   `yaml:"target" validate:"required"`
	Sync     SyncConfig     `yaml:"sync"`
	Defaults DefaultsConfig `yaml:"defaults"`
	Rules    []Rule         `yaml:"rules" validate:"dive"`
}

// ServerConfig holds CalDAV server connection references.
type ServerConfig struct {
	URL      *url.URL `yaml:"url" validate:"required"`
	Username string   `yaml:"username"`
	Password string   `yaml:"password"` // #nosec G117 -- configuration field, not a hardcoded secret
}

// TargetConfig identifies the calendar to operate on.
type TargetConfig struct {
	URL *url.URL `yaml:"url" validate:"required"`
}

// SyncConfig defines horizon and lookback settings.
type SyncConfig struct {
	HorizonDays  int `yaml:"horizonDays" validate:"gt=0"`
	LookbackDays int `yaml:"lookbackDays" validate:"gt=0"`
}

// DefaultsConfig configures rule defaults.
type DefaultsConfig struct {
	Timezone *time.Location `yaml:"timezone"`
	Due      DuePreference  `yaml:"due"`
}

// DuePreference describes default due-time behavior.
type DuePreference struct {
	Time     ClockTime `yaml:"time"`
	DateOnly bool      `yaml:"dateOnly"`
}

// Rule defines a recurrence rule.
type Rule struct {
	ID       string       `yaml:"id" validate:"required"`
	Title    string       `yaml:"title" validate:"required"`
	Notes    string       `yaml:"notes"`
	Schedule RuleSchedule `yaml:"schedule" validate:"required"`
}

// RuleSchedule holds recurrence parameters.
type RuleSchedule struct {
	Kind             ScheduleKind   `yaml:"kind" validate:"validateFn=IsAScheduleKind"` // technically required through validateFn
	Weekdays         []time.Weekday `yaml:"weekdays" validate:"dive"`
	EveryNDays       int            `yaml:"everyNDays" validate:"gte=0"`
	MonthDays        []int          `yaml:"monthDays" validate:"dive,gte=1,lte=31"`
	Month            int            `yaml:"month" validate:"gte=0,lte=12"`
	Day              int            `yaml:"day" validate:"gte=0,lte=31"`
	Nth              int            `yaml:"nth" validate:"gte=0"`
	NthWeekday       time.Weekday   `yaml:"nthWeekday"`
	YearlyNthWeekday time.Weekday   `yaml:"yearlyNthWeekday"`
}

// ClockTime represents an hour and minute.
type ClockTime struct {
	Hour   int
	Minute int
}

func validateConfig(cfg Config) error {
	v := newValidator()
	if err := v.Struct(cfg); err != nil {
		return err
	}
	return nil
}

func newValidator() *validator.Validate {
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.RegisterValidation("validateFn", validateFn); err != nil {
		panic(err)
	}
	validate.RegisterStructValidation(validateRuleSchedule, RuleSchedule{})
	return validate
}

var enumValidation = map[string]func(any) bool{
	"IsAScheduleKind": func(value any) bool {
		kind, ok := value.(ScheduleKind)
		return ok && kind.IsAScheduleKind()
	},
}

func validateFn(fl validator.FieldLevel) bool {
	name := fl.Param()
	if name != "" {
		fn, ok := enumValidation[name]
		if !ok {
			return false
		}
		return fn(fl.Field().Interface())
	}
	if value, ok := fl.Field().Interface().(interface{ Validate() error }); ok {
		return value.Validate() == nil
	}
	if fl.Field().CanAddr() {
		if value, ok := fl.Field().Addr().Interface().(interface{ Validate() error }); ok {
			return value.Validate() == nil
		}
	}
	return true
}

func finalizeConfig(cfg *Config) error {
	username := os.Getenv(envUsernameVar)
	password := os.Getenv(envPasswordVar)
	if username == "" || password == "" {
		slog.Error("missing server credentials")
		return errors.New("missing server credentials in environment")
	}
	cfg.Server.Username = username
	cfg.Server.Password = password

	if cfg.Defaults.Timezone == nil {
		cfg.Defaults.Timezone = time.UTC
	}

	seen := make(map[string]struct{})
	for i := range cfg.Rules {
		rule := &cfg.Rules[i]
		if _, exists := seen[rule.ID]; exists {
			slog.Error("detected duplicate rule id", "index", i, "id", rule.ID)
			return fmt.Errorf("rules[%d].id must be unique", i)
		}
		seen[rule.ID] = struct{}{}
	}

	return nil
}

// Load reads and validates configuration from disk.
func Load(path string) (Config, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		slog.Error("failed to resolve config path", "error", err)
		return Config{}, fmt.Errorf("resolve config path %q: %w", path, err)
	}
	rootDir := filepath.Dir(absPath)
	fileName := filepath.Base(absPath)

	root, err := os.OpenRoot(rootDir)
	if err != nil {
		slog.Error("failed to open config root", "error", err)
		return Config{}, fmt.Errorf("open config root %q: %w", rootDir, err)
	}
	defer func() {
		_ = root.Close()
	}()

	raw, err := root.ReadFile(fileName)
	if err != nil {
		slog.Error("failed to read config file", "error", err)
		return Config{}, fmt.Errorf("read config file %q: %w", fileName, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		slog.Error("failed to parse config", "error", err)
		return Config{}, fmt.Errorf("parse config %q: %w", fileName, err)
	}

	if err := validateConfig(cfg); err != nil {
		return Config{}, fmt.Errorf("validate config: %w", err)
	}

	if err := finalizeConfig(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
