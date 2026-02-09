package monitor

import (
	"context"
)

type Monitor interface {
	// Name returns the name of the monitor
	Name() string

	// OnTick runs every time the monitor timer ticks.
	// It returns a single metric that indicates the resource load
	OnTick(ctx context.Context) (int, error)
}
