package scaler

import (
	"context"
	"log/slog"
)

type Scaler interface {
	// Name returns the name of the monitor
	Name() string

	// OnTick runs every time the manager timer ticks.
	// It accepts a single metric value that can be used to scale resource
	OnTick(ctx context.Context, metric int) error
}

type MockScaler struct {
	name string
}

func NewMockScaler() *MockScaler {
	return &MockScaler{}
}

func (m *MockScaler) Name() string {
	return m.name
}

func (m *MockScaler) OnTick(ctx context.Context, metric int) error {
	slog.Info("Mock scaler triggered", slog.Int("metric", metric))
	return nil
}
