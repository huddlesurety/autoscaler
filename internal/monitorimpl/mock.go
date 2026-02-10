package monitorimpl

import (
	"context"

	"github.com/huddlesurety/autoscaler/internal/config"
	"github.com/huddlesurety/autoscaler/pkg/monitor"
)

type Mock struct {
	name string
}

func NewMockMonitor(cfg *config.Config) (monitor.Monitor, error) {
	return &Mock{
		name: "Mock",
	}, nil
}

func (m Mock) Name() string {
	return m.name
}

func (m *Mock) OnTick(ctx context.Context, tickID int) (int, error) {
	return 0, nil
}
