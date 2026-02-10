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

	temporal *temporal.Client
}

func newScalers(cfg *config.Config) (*scalers, error) {
	tc, err := temporal.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Temporal client: %w", err)
	}

	rag := scalerimpl.NewRAGScaler(cfg, tc)

	return &scalers{
		rag: rag,
	}, nil
}

func (s *scalers) close() {
	s.temporal.Close()
}
