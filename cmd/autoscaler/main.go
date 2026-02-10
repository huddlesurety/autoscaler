package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/huddlesurety/autoscaler/internal/config"
	"github.com/huddlesurety/autoscaler/pkg/manager"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
)

func main() {
	if err := run(); err != nil {
		slog.Error("Autoscaler failed", slog.Any("error", err))
	}
}

func run() error {
	slog.Info("Autoscaler started")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.New()
	if err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	manager, err := manager.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize monitor manager: %w", err)
	}

	monitors, err := newMonitors(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize monitors: %w", err)
	}

	controllers, err := newControllers(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize monitors: %w", err)
	}

	manager.Register(cfg.Railway.ServiceRAG, monitors.Temporal, controllers.rag)
	manager.Register(cfg.Railway.ServiceAPI, monitors.Mock, controllers.api)
	manager.Register(cfg.Railway.ServiceApp, monitors.Mock, controllers.app)

	go manager.Run(ctx)

	slog.Info("Manager started", slog.String("interval", cfg.Interval.String()))

	<-ctx.Done()
	stop()

	slog.Info("Autoscaler stopped")

	return nil
}

func init() {
	lg := slog.New(
		tint.NewHandler(os.Stdout, &tint.Options{
			Level:      slog.LevelInfo,
			NoColor:    !isatty.IsTerminal(os.Stdout.Fd()),
			TimeFormat: time.Kitchen,
		}),
	)
	slog.SetDefault(lg)
}
