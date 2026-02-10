package scaler

import (
	"context"
)

type Scaler interface {
	// ServiceID returns the ID of the target railway service
	ServiceID() string

	// GetMetric emits the metric used to determine the number of desired replicas.
	GetMetric(ctx context.Context) (metric float64, err error)

	// Scale receives the average metric over scale interval and emits the number of desired replicas.
	Scale(avg float64) (desired int)
}
