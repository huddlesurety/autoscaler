package monitor

import (
	"context"
	"log/slog"
	"time"

	"github.com/huddlesurety/autoscaler/internal/config"
	"github.com/huddlesurety/autoscaler/internal/scaler"
)

type Monitor interface {
	// Name returns the name of the monitor
	Name() string

	// OnTick runs every time the monitor timer ticks.
	// It returns a single metric that indicates the resource load
	OnTick(ctx context.Context) (int, error)
}

type Manager struct {
	cfg   *config.Config
	pairs []pair
}

type pair struct {
	monitor Monitor
	scaler  scaler.Scaler
}

func NewManager(cfg *config.Config) (*Manager, error) {
	pairs := make([]pair, 0)

	return &Manager{
		cfg:   cfg,
		pairs: pairs,
	}, nil
}

// Pairs the given monitor and scaler by feeding the metric from the monitor to the scaler
func (man *Manager) Register(m Monitor, s scaler.Scaler) {
	man.pairs = append(man.pairs, pair{m, s})
}

// Runs the manager loop indefinitely. Executes the regiestered monitors and tickers on tick.
func (man *Manager) Run(ctx context.Context) {
	interval := time.Second * time.Duration(man.cfg.IntervalSeconds)
	ticker := time.NewTicker(interval)

loop:
	for {
		select {
		case <-ticker.C:
			success := 0
			ctx, cancel := context.WithTimeout(ctx, interval)
			for _, p := range man.pairs {
				metric, err := p.monitor.OnTick(ctx)
				if err != nil {
					slog.Error("Failed to retrieve metric",
						slog.String("monitor", p.monitor.Name()),
						slog.Any("error", err),
					)
					continue
				}

				if err := p.scaler.OnTick(ctx, metric); err != nil {
					slog.Error("Failed to scale",
						slog.String("scaler", p.scaler.Name()),
						slog.Any("error", err),
					)
					continue
				}

				success++
			}
			slog.Info("Tick", slog.Int("success", success), slog.Int("failure", len(man.pairs)-success))
			cancel()
		case <-ctx.Done():
			slog.Info("Manager stopping")
			break loop
		}
	}
}
