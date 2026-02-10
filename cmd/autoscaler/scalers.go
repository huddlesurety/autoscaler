package main

import (
	"fmt"

	"github.com/huddlesurety/autoscaler/internal/config"
	"github.com/huddlesurety/autoscaler/internal/scalerimpl"
	"github.com/huddlesurety/autoscaler/internal/temporal"
	"github.com/huddlesurety/autoscaler/pkg/scaler"
)

type scalers struct {
	rag scaler.Scaler
}

func newScalers(cfg *config.Config) (*scalers, error) {
	temporal, err := temporal.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Temporal client: %w", err)
	}

	rag, err := scalerimpl.NewRAGScaler(cfg, temporal)

	return &scalers{
		rag: rag,
	}, nil
}
