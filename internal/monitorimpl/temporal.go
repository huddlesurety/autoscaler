package monitorimpl

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/huddlesurety/autoscaler/internal/config"
	"github.com/huddlesurety/autoscaler/pkg/monitor"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
)

type Temporal struct {
	name   string
	client client.Client

	cacheKey int
	cache    int
}

func NewTemporalMonitor(cfg *config.Config) (monitor.Monitor, error) {
	c, err := client.Dial(client.Options{
		Logger:   slog.Default(),
		HostPort: cfg.Temporal.ServerURL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Temporal client: %w", err)
	}

	return &Temporal{
		name:   "Temporal",
		client: c,
	}, nil
}

func (m Temporal) Name() string {
	return m.name
}

// Emits the number of running temporal workflow executions.
// Returns cached value when tickID is identical.
func (m *Temporal) OnTick(ctx context.Context, tickID int) (int, error) {
	if m.cacheKey != tickID {
		resp, err := m.client.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
			Query: "ExecutionStatus = 'Running'",
		})
		if err != nil {
			return -1, fmt.Errorf("failed to list open workflows: %w", err)
		}
		m.cacheKey = tickID
		m.cache = len(resp.Executions)
	}

	return m.cache, nil
}
