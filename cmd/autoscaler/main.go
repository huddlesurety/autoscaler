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
	"github.com/huddlesurety/autoscaler/internal/monitor"
	"github.com/huddlesurety/autoscaler/internal/scaler"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
)

func main() {
	if err := run(); err != nil {
		slog.Error("Autoscaler failed", slog.Any("error", err))
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.New()
	if err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	slog.Info("Autoscaler started")

	manager, err := monitor.NewManager(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize monitor manager")
	}

	temporalMonitor, err := monitor.NewTemporalMonitor(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize Temporal monitor")
	}

	mockScaler := scaler.NewMockScaler()

	manager.Register(temporalMonitor, mockScaler)

	go manager.Run(ctx)

	slog.Info("Manager started", slog.String("interval", fmt.Sprintf("%ds", cfg.IntervalSeconds)))

	<-ctx.Done()

	slog.Info("Autoscaler stopping")

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
