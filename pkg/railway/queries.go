package railway

import (
	"context"
	"fmt"
)

type Service struct {
	ID               string
	Name             string `json:"serviceName"`
	Replicas         int    `json:"numReplicas"`
	LatestDeployment struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	} `json:"latestDeployment"`
}

func (c *Client) GetService(ctx context.Context, serviceID string) (*Service, error) {
	query := `query serviceInstance($serviceId: String!, $environmentId: String!) {
serviceInstance(serviceId: $serviceId, environmentId: $environmentId) {
  id
  serviceName
  startCommand
  buildCommand
  rootDirectory
  healthcheckPath
  region
  numReplicas
  restartPolicyType
  restartPolicyMaxRetries
  latestDeployment {
    id
    status
    createdAt
  }
}
}`
	body := &request{
		Query: query,
		Variables: map[string]any{
			"serviceId":     serviceID,
			"environmentId": c.cfg.Railway.EnvironmentID,
		},
	}

	var res struct {
		Data struct {
			Service Service `json:"serviceInstance"`
		} `json:"data"`
	}

	if err := c.request(ctx, body, &res); err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	svc := &res.Data.Service
	svc.ID = serviceID
	if svc.LatestDeployment.ID == "" {
		svc.Replicas = 0
	}

	return svc, nil
}

func (c *Client) Deploy(ctx context.Context, serviceID string) error {
	query := `mutation serviceInstanceDeployV2($serviceId: String!, $environmentId: String!) {
serviceInstanceDeployV2(serviceId: $serviceId, environmentId: $environmentId)
}`
	body := &request{
		Query: query,
		Variables: map[string]any{
			"serviceId":     serviceID,
			"environmentId": c.cfg.Railway.EnvironmentID,
		},
	}

	var res any
	if err := c.request(ctx, body, &res); err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}

	return nil
}

func (c *Client) Scale(ctx context.Context, serviceID string, replicas int) error {
	if replicas == 0 {
		return c.scaleZero(ctx, serviceID)
	}
	return c.scaleNonZero(ctx, serviceID, replicas)
}

// scale by updating the instance config
func (c *Client) scaleNonZero(ctx context.Context, serviceID string, replicas int) error {
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

	svc, err := c.GetService(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}

	if svc.Replicas == 0 {
		if err := c.Deploy(ctx, serviceID); err != nil {
			return fmt.Errorf("failed to deploy service: %w", err)
		}
	}

	return nil
}

// scale by removing the running deployment
func (c *Client) scaleZero(ctx context.Context, serviceID string) error {
	svc, err := c.GetService(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}

	if svc.Replicas == 0 {
		return nil
	}

	query := `mutation deploymentRemove($id: String!) {
deploymentRemove(id: $id)
}`
	body := &request{
		Query: query,
		Variables: map[string]any{
			"id": svc.LatestDeployment.ID,
		},
	}

	var res any
	if err := c.request(ctx, body, &res); err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}

	return nil
}
