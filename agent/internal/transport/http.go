package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"agent/internal/collectors"
	"agent/internal/config"
)

type HTTPClient struct {
	endpoint string
	token    string
	client   *http.Client
}

func NewHTTPClient(cfg config.CloudConfig) (*HTTPClient, error) {
	endpoint := strings.TrimSpace(cfg.Endpoint)
	if endpoint == "" {
		return nil, fmt.Errorf("cloud endpoint is empty")
	}
	return &HTTPClient{
		endpoint: strings.TrimRight(endpoint, "/"),
		token:    cfg.Token,
		client:   &http.Client{Timeout: 8 * time.Second},
	}, nil
}

func (c *HTTPClient) SendSnapshots(ctx context.Context, snapshots []collectors.Snapshot) error {
	if len(snapshots) == 0 {
		return nil
	}
	payload, err := json.Marshal(map[string]any{"snapshots": snapshots})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+"/v1/agent/snapshots", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("snapshot upload failed: %s", resp.Status)
	}
	return nil
}
