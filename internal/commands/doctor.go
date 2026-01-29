package commands

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/eikendev/taskseed/internal/caldav"
	"github.com/eikendev/taskseed/internal/config"
)

// DoctorCommand validates configuration and basic connectivity.
type DoctorCommand struct {
	Config string `name:"config" short:"c" help:"Path to configuration file." default:"config.yaml" env:"TASKSEED_CONFIG"`
}

// Run executes the doctor command.
func (cmd *DoctorCommand) Run() error {
	cfg, err := config.Load(cmd.Config)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		return fmt.Errorf("load config: %w", err)
	}

	client, err := caldav.NewClient(cfg.Server.URL.String(), cfg.Target.URL.String(), cfg.Server.Username, cfg.Server.Password)
	if err != nil {
		slog.Error("failed to create caldav client", "error", err)
		return fmt.Errorf("create caldav client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_, err = client.QueryTasks(ctx, time.Now().Add(-24*time.Hour), time.Now().Add(24*time.Hour))
	if err != nil {
		slog.Error("failed to check connectivity", "error", err)
		return fmt.Errorf("caldav doctor failed: %w", err)
	}

	return nil
}
