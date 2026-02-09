package monitor

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/huddlesurety/autoscaler/internal/config"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
)

type TemporalMonitor struct {
	cfg    *config.Config
	name   string
	client client.Client
}

func NewTemporalMonitor(cfg *config.Config) (*TemporalMonitor, error) {
	c, err := client.Dial(client.Options{
		Logger:   slog.Default(),
		HostPort: cfg.Temporal.ServerURL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Temporal client: %w", err)
	}

	return &TemporalMonitor{
		cfg:    cfg,
		name:   "Temporal Monitor",
		client: c,
	}, nil
}

func (m *TemporalMonitor) Name() string {
	return m.name
}

// Emits the number of running temporal workflow executions
func (m *TemporalMonitor) OnTick(ctx context.Context) (int, error) {
	resp, err := m.client.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
		Query: "ExecutionStatus = 'Running'",
	})
	if err != nil {
		return -1, fmt.Errorf("failed to list open workflows: %w", err)
	}

	return len(resp.Executions), nil
}
