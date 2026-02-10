package main

import (
	"context"

	"github.com/huddlesurety/autoscaler/internal/config"
)

type controllers struct {
	cfg *config.Config
}

func newControllers(cfg *config.Config) (*controllers, error) {
	return &controllers{
		cfg: cfg,
	}, nil
}

func (c *controllers) rag(ctx context.Context, metric int) (int, error) {
	// TODO
	return 1, nil
}

func (c *controllers) api(ctx context.Context, metric int) (int, error) {
	// TODO
	return 1, nil
}

func (c *controllers) app(ctx context.Context, metric int) (int, error) {
	// TODO
	return 1, nil
}
