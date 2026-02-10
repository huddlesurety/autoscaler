package manager

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/huddlesurety/autoscaler/internal/config"
	"github.com/huddlesurety/autoscaler/pkg/monitor"
	"github.com/huddlesurety/autoscaler/pkg/railway"
)

type Manager struct {
	cfg     *config.Config
	railway *railway.Client
	scalers []scaler
}

// Based on the metric, determines the desired replicas
type ControllerFunc func(ctx context.Context, metric int) (desired int, err error)

type scaler struct {
	serviceID string
	monitor   monitor.Monitor
	control   ControllerFunc
}

func New(cfg *config.Config) (*Manager, error) {
	scalers := make([]scaler, 0)

	rc, err := railway.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Railway client")
	}

	return &Manager{
		cfg:     cfg,
		railway: rc,
		scalers: scalers,
	}, nil
}

// Pairs the given monitor and controller by feeding the metric from the monitor to the scaler
func (man *Manager) Register(serviceID string, m monitor.Monitor, control ControllerFunc) {
	man.scalers = append(man.scalers, scaler{serviceID, m, control})
}

// Runs the manager loop indefinitely. Executes the regiestered monitors and tickers on tick.
func (man *Manager) Run(ctx context.Context) {
	ticker := time.NewTicker(man.cfg.Interval)
	tickID := 0

loop:
	for {
		select {
		case <-ticker.C:
			man.onTick(ctx, tickID)
			tickID++
		case <-ctx.Done():
			slog.Info("Manager stopping")
			break loop
		}
	}
}

func (man *Manager) onTick(ctx context.Context, tickID int) {
	success := 0
	ctx, cancel := context.WithTimeout(ctx, man.cfg.Interval)

	var wg sync.WaitGroup
	for _, s := range man.scalers {
		wg.Go(func() {
			before, err := man.railway.GetService(ctx, s.serviceID)
			if err != nil {
				slog.Error("Failed to get service", slog.Any("error", err))
				return
			}

			attrService := slog.String("service", before.Name)
			attrMonitor := slog.String("monitor", s.monitor.Name())

			metric, err := s.monitor.OnTick(ctx, tickID)
			if err != nil {
				slog.Error("Failed to retrieve metric",
					attrService,
					attrMonitor,
					slog.Any("error", err),
				)
				return
			}

			desired, err := s.control(ctx, metric)
			if err != nil {
				slog.Error("Failed to get desired replicas",
					attrService,
					attrMonitor,
					slog.Any("error", err),
				)
				return
			}

			if err := man.railway.Scale(ctx, s.serviceID, desired); err != nil {
				slog.Error("Failed to scale",
					attrService,
					attrMonitor,
					slog.Any("error", err),
				)
				return
			}

			after, err := man.railway.GetService(ctx, s.serviceID)
			if err != nil {
				slog.Error("Failed to get service", slog.Any("error", err))
				return
			}

			attrReplicasBefore := slog.Int("before", before.Replicas)
			attrReplicasAfter := slog.Int("after", after.Replicas)

			fmt.Println(after.LatestDeployment.Status)

			slog.Info("Scaled successfully",
				attrService,
				attrMonitor,
				attrReplicasBefore,
				attrReplicasAfter,
				slog.Int("desired", desired),
			)
		})
		success++
	}

	wg.Wait()

	slog.Info("Tick",
		slog.Int("id", tickID),
		slog.Int("success", success),
		slog.Int("failure", len(man.scalers)-success),
	)
	cancel()
}
