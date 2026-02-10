package scalerimpl

import (
	"context"
	"fmt"

	"github.com/huddlesurety/autoscaler/internal/config"
	"github.com/huddlesurety/autoscaler/internal/temporal"
	"github.com/huddlesurety/autoscaler/pkg/scaler"
)

type RAGScaler struct {
	temporal       *temporal.Client
	serviceID      string
	workflowsQuery string
}

func NewRAGScaler(cfg *config.Config, temporal *temporal.Client) (scaler.Scaler, error) {
	wq := fmt.Sprintf("ExecutionStatus = '%s' and TaskQueue = '%s'", "Running", cfg.Temporal.TaskQueueRAG)

	return &RAGScaler{
		serviceID:      cfg.Railway.ServiceRAG,
		temporal:       temporal,
		workflowsQuery: wq,
	}, nil
}

func (m RAGScaler) ServiceID() string {
	return m.serviceID
}

func (m RAGScaler) GetMetric(ctx context.Context) (float64, error) {
	workflows, err := m.temporal.GetWorkflowCount(ctx, m.workflowsQuery)
	if err != nil {
		return -1, fmt.Errorf("failed to get workflows count: %w", err)
	}

	return float64(workflows), nil
}

func (m RAGScaler) Scale(avg float64) int {
	switch {
	case avg == 0:
		return 0
	case 0 < avg && avg <= 10:
		return 1
	case 10 < avg && avg < 20:
		return 2
	default:
		return 3
	}
}
