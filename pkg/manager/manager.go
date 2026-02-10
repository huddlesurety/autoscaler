package manager

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/huddlesurety/autoscaler/pkg/railway"
	"github.com/huddlesurety/autoscaler/pkg/scaler"
)

type Manager struct {
	cfg     *Config
	railway *railway.Client
	pairs   []*serviceScaler
}

type serviceScaler struct {
	target *railway.Service
	scaler scaler.Scaler

	metricMu    sync.Mutex
	metricSum   float64
	metricCount int
}

func New(cfg *Config) (*Manager, error) {
	scalers := make([]*serviceScaler, 0)

	rc, err := railway.NewClient(cfg.RailwayEnvironmentID, cfg.RailwayToken)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Railway client: %w", err)
	}

	return &Manager{
		cfg:     cfg,
		railway: rc,
		pairs:   scalers,
	}, nil
}

// Pairs the given service and scaler
func (man *Manager) Register(ctx context.Context, serviceID string, s scaler.Scaler) error {
	svc, err := man.railway.GetService(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}

	man.pairs = append(man.pairs, &serviceScaler{
		target:      svc,
		scaler:      s,
		metricSum:   0,
		metricCount: 0,
	})

	slog.Debug("Scaler registered",
		slog.Group("target", slog.String("id", svc.ID), slog.String("name", svc.Name)),
	)

	return nil
}

// Runs the manager loop until the context is canceled. Executes the registered scalers based on the configured interval.
func (man *Manager) Run(ctx context.Context) {
	metricTicker := time.NewTicker(man.cfg.MetricInterval)
	scaleTicker := time.NewTicker(man.cfg.ScaleInterval)

loop:
	for {
		select {
		case <-metricTicker.C:
			man.onTickMetric(ctx)
		case <-scaleTicker.C:
			man.onTickScale(ctx)
		case <-ctx.Done():
			slog.Info("Manager stopping")
			break loop
		}
	}
}

func (man *Manager) onTickMetric(ctx context.Context) {
	var success atomic.Int64
	ctx, cancel := context.WithTimeout(ctx, man.cfg.MetricInterval)
	defer cancel()

	var wg sync.WaitGroup
	for _, p := range man.pairs {
		wg.Go(func() {
			metric, err := p.scaler.GetMetric(ctx)
			if err != nil {
				slog.Error("Failed to fetch metric",
					slog.String("target", p.target.Name),
					slog.Any("error", err),
				)
				return
			}
			func() {
				p.metricMu.Lock()
				defer p.metricMu.Unlock()
				p.metricCount++
				p.metricSum += metric
			}()

			success.Add(1)

			slog.Debug("Metric fetched",
				slog.String("target", p.target.Name),
				slog.Float64("value", metric),
			)
		})
	}

	wg.Wait()

	slog.Info("Tick metric",
		slog.Int64("success", success.Load()),
		slog.Int("total", len(man.pairs)),
	)
}

func (man Manager) onTickScale(ctx context.Context) {
	var success atomic.Int64
	ctx, cancel := context.WithTimeout(ctx, man.cfg.ScaleInterval)
	defer cancel()

	var wg sync.WaitGroup
	for _, p := range man.pairs {
		wg.Go(func() {
			avg, ok := func() (float64, bool) {
				p.metricMu.Lock()
				defer p.metricMu.Unlock()
				if p.metricCount == 0 {
					return 0, false
				}
				avg := p.metricSum / float64(p.metricCount)
				p.metricSum = 0
				p.metricCount = 0
				return avg, true
			}()
			if !ok {
				slog.Warn("No metrics collected, skipping scale",
					slog.String("target", p.target.Name),
				)
				return
			}

			current, err := man.railway.GetService(ctx, p.target.ID)
			if err != nil {
				slog.Error("Failed to get service", slog.Any("error", err))
				return
			}

			desired := p.scaler.Scale(avg)
			skip := current.Replicas == desired

			if !skip {
				if err := man.railway.Scale(ctx, p.target.ID, desired); err != nil {
					slog.Error("Failed to scale service",
						slog.String("target", p.target.Name),
						slog.Any("error", err),
					)
					return
				}
			}

			success.Add(1)

			slog.Debug("Service scaled",
				slog.String("target", p.target.Name),
				slog.Int("replicas", desired),
				slog.Bool("skip", skip),
			)
		})
	}

	wg.Wait()

	slog.Info("Tick scale",
		slog.Int64("success", success.Load()),
		slog.Int("total", len(man.pairs)),
	)
}
