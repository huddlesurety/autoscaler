package railway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	endpoint = "https://backboard.railway.com/graphql/v2"
)

type request struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

func (c *Client) request(ctx context.Context, reqBody *request, data any) error {
	b, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	r := bytes.NewReader(b)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, r)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.token))

	res, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	// nolint:errcheck
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var m map[string]any
	if err := json.Unmarshal(resBody, &m); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	errors, ok := m["errors"]
	if ok {
		return fmt.Errorf("request error: %v", errors)
	}

	if err := json.Unmarshal(resBody, data); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}
