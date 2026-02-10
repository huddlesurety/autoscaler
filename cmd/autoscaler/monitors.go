package main

import (
	"fmt"

	"github.com/huddlesurety/autoscaler/internal/config"
	"github.com/huddlesurety/autoscaler/internal/monitorimpl"
	"github.com/huddlesurety/autoscaler/pkg/monitor"
)

type monitors struct {
	Mock     monitor.Monitor
	Temporal monitor.Monitor
}

func newMonitors(cfg *config.Config) (*monitors, error) {
	mock, err := monitorimpl.NewMockMonitor(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize mock monitor")
	}

	temporal, err := monitorimpl.NewTemporalMonitor(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Temporal monitor")
	}

	return &monitors{
		Mock:     mock,
		Temporal: temporal,
	}, nil
}
