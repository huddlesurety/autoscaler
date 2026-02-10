package temporal

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/huddlesurety/autoscaler/internal/config"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
)

type Client struct {
	client client.Client
}

func NewClient(cfg *config.Config) (*Client, error) {
	c, err := client.Dial(client.Options{
		Logger:   slog.Default(),
		HostPort: cfg.Temporal.ServerURL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Temporal client: %w", err)
	}

	return &Client{
		client: c,
	}, nil
}

func (c *Client) GetWorkflowCount(ctx context.Context, query string) (int64, error) {
	resp, err := c.client.CountWorkflow(ctx, &workflowservice.CountWorkflowExecutionsRequest{
		Query: query,
	})
	if err != nil {
		return -1, fmt.Errorf("failed to count workflows: %w", err)
	}

	return resp.Count, nil
}

func (c *Client) Close() {
	c.client.Close()
}
