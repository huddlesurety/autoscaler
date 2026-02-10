package railway

import (
	"context"
	"fmt"
	"net/http"

	"github.com/huddlesurety/autoscaler/internal/config"
)

type Client struct {
	cfg    *config.Config
	client *http.Client
}

type Service struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func NewClient(cfg *config.Config) (*Client, error) {
	c := new(http.Client)
	return &Client{
		cfg:    cfg,
		client: c,
	}, nil
}

func (c *Client) GetService(ctx context.Context, serviceID string) (*Service, error) {
	query := `query service($id: String!) {
service(id: $id) {
  id
  name
  icon
  createdAt
  projectId
}
}`
	body := &request{
		Query: query,
		Variables: map[string]any{
			"id": serviceID,
		},
	}

	var res struct {
		Data struct {
			Service `json:"service"`
		} `json:"data"`
	}

	if err := c.request(ctx, body, &res); err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	return &res.Data.Service, nil
}

func (c *Client) Scale(ctx context.Context, serviceID string, replicas int) error {
	query := `mutation serviceInstanceUpdate($serviceId: String!, $environmentId: String!, $input: ServiceInstanceUpdateInput!) {
serviceInstanceUpdate(serviceId: $serviceId, environmentId: $environmentId, input: $input)
}`

	body := &request{
		Query: query,
		Variables: map[string]any{
			"serviceId":     serviceID,
			"environmentId": c.cfg.Railway.EnvironmentID,
			"input": map[string]any{
				"numReplicas": replicas,
			},
		},
	}

	var res any

	if err := c.request(ctx, body, &res); err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}

	return nil
}
