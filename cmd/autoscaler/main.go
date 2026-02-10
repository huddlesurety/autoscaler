package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/huddlesurety/autoscaler/internal/config"
	"github.com/huddlesurety/autoscaler/pkg/manager"
)

func main() {
	if err := run(); err != nil {
		slog.Error("Autoscaler failed", slog.Any("error", err))
		os.Exit(1)
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

	man, err := manager.New(&manager.Config{
		RailwayEnvironmentID: cfg.Railway.EnvironmentID,
		RailwayToken:         cfg.Railway.Token,
		MetricInterval:       cfg.IntervalMetric,
		ScaleInterval:        cfg.IntervalScale,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize manager: %w", err)
	}

	scalers, err := newScalers(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize scalers: %w", err)
	}
	defer scalers.close()

	if err := man.Register(ctx, cfg.Railway.ServiceRAG, scalers.rag); err != nil {
		return fmt.Errorf("failed to register RAG scaler: %w", err)
	}

	go man.Run(ctx)

	slog.Info("Manager started",
		slog.Group("interval",
			slog.String("metric", cfg.IntervalMetric.String()),
			slog.String("scale", cfg.IntervalScale.String()),
		),
	)

	<-ctx.Done()
	stop()

	slog.Info("Autoscaler stopped")

	return nil
}

func init() {
	lg := slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}),
	)
	slog.SetDefault(lg)
}
