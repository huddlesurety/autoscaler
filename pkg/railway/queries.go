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
			"environmentId": c.environmentID,
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

func (c *Client) restartDeployment(ctx context.Context, deploymentID string) error {
	query := `mutation deploymentRestart($id: String!) {
deploymentRestart(id: $id)
}`
	body := &request{
		Query: query,
		Variables: map[string]any{
			"id": deploymentID,
		},
	}

	var res any
	if err := c.request(ctx, body, &res); err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}

	return nil
}

func (c *Client) Scale(ctx context.Context, serviceID string, replicas int) error {
	svc, err := c.GetService(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}

	if replicas == 0 {
		return c.scaleZero(ctx, svc)
	}
	return c.scaleNonZero(ctx, svc, replicas)
}

// scale by updating the instance config
func (c *Client) scaleNonZero(ctx context.Context, svc *Service, replicas int) error {
	if svc.Replicas == replicas {
		return nil
	}

	query := `mutation serviceInstanceUpdate($serviceId: String!, $environmentId: String!, $input: ServiceInstanceUpdateInput!) {
serviceInstanceUpdate(serviceId: $serviceId, environmentId: $environmentId, input: $input)
}`
	body := &request{
		Query: query,
		Variables: map[string]any{
			"serviceId":     svc.ID,
			"environmentId": c.environmentID,
			"input": map[string]any{
				"numReplicas": replicas,
			},
		},
	}

	var res any
	if err := c.request(ctx, body, &res); err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}

	// if previously scaled to zero, restart
	if svc.Replicas == 0 {
		if err := c.restartDeployment(ctx, svc.LatestDeployment.ID); err != nil {
			return fmt.Errorf("failed to deploy service: %w", err)
		}
	}

	return nil
}

// scale by stopping the running deployment
func (c *Client) scaleZero(ctx context.Context, svc *Service) error {
	if svc.Replicas == 0 {
		return nil
	}

	query := `mutation deploymentStop($id: String!) {
deploymentStop(id: $id)
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
