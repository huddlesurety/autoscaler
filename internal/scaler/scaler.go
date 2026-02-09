package scaler

import (
	"context"
)

type Scaler interface {
	// Name returns the name of the monitor
	Name() string

	// OnTick runs every time the manager timer ticks.
	// It accepts a single metric value that can be used to scale resource
	OnTick(ctx context.Context, metric int) error
}
