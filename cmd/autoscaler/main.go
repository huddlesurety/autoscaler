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

	manager, err := manager.New(&manager.Config{
		RailwayEnvironmentID: cfg.Railway.EnvironmentID,
		RailwayToken:         cfg.Railway.Token,
		MetricInterval:       cfg.MetricInterval,
		ScaleInterval:        cfg.ScaleInterval,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize manager: %w", err)
	}

	scalers, err := newScalers(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize scalers: %w", err)
	}
	defer scalers.close()

	if err := manager.Register(ctx, cfg.Railway.ServiceRAG, scalers.rag); err != nil {
		return fmt.Errorf("failed to register RAG scaler: %w", err)
	}

	go manager.Run(ctx)

	slog.Info("Manager started", slog.String("interval", cfg.MetricInterval.String()))

	<-ctx.Done()
	stop()

	slog.Info("Autoscaler stopped")

	return nil
}

func init() {
	lg := slog.New(
		tint.NewHandler(os.Stdout, &tint.Options{
			Level:      slog.LevelDebug,
			NoColor:    !isatty.IsTerminal(os.Stdout.Fd()),
			TimeFormat: time.Kitchen,
		}),
	)
	slog.SetDefault(lg)
}
