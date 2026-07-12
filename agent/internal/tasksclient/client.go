package tasksclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"agent/internal/commands"
	"agent/internal/config"
)

type Client struct {
	endpoint string
	token    string
	client   *http.Client
}

type Task struct {
	ID       string      `json:"id"`
	ServerID string      `json:"server_id"`
	Name     string      `json:"name"`
	Payload  TaskPayload `json:"payload,omitempty"`
	Status   string      `json:"status"`
}

type TaskPayload struct {
	Service    string   `json:"service,omitempty"`
	Action     string   `json:"action,omitempty"`
	Domains    []string `json:"domains,omitempty"`
	IncidentID string   `json:"incident_id,omitempty"`
}

type TaskResult struct {
	ExitCode   int       `json:"exit_code"`
	Stdout     string    `json:"stdout"`
	Stderr     string    `json:"stderr"`
	DurationMS int64     `json:"duration_ms"`
	StartedAt  time.Time `json:"started_at"`
	Error      string    `json:"error,omitempty"`
}

func New(cfg config.CloudConfig) (*Client, error) {
	endpoint := strings.TrimSpace(cfg.Endpoint)
	if endpoint == "" {
		return nil, fmt.Errorf("cloud endpoint is empty")
	}
	tlsConfig, err := buildTLSConfig(cfg.MTLS)
	if err != nil {
		return nil, err
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if tlsConfig != nil {
		transport.TLSClientConfig = tlsConfig
	}
	return &Client{endpoint: strings.TrimRight(endpoint, "/"), token: cfg.Token, client: &http.Client{Timeout: 10 * time.Second, Transport: transport}}, nil
}

func (c *Client) Poll(ctx context.Context, agentID string, limit int) ([]Task, error) {
	values := url.Values{}
	values.Set("agent_id", agentID)
	if limit > 0 {
		values.Set("limit", fmt.Sprint(limit))
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint+"/v1/agent/tasks?"+values.Encode(), nil)
	if err != nil {
		return nil, err
	}
	c.authorize(req)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("poll tasks failed: %s", resp.Status)
	}
	var out struct {
		Tasks []Task `json:"tasks"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.Tasks, nil
}

func (c *Client) Complete(ctx context.Context, taskID string, result TaskResult) error {
	payload, err := json.Marshal(result)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+"/v1/agent/tasks/"+url.PathEscape(taskID)+"/result", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	c.authorize(req)
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("complete task failed: %s", resp.Status)
	}
	return nil
}

func FromCommandResult(result commands.Result, runErr error) TaskResult {
	out := TaskResult{ExitCode: result.ExitCode, Stdout: result.Stdout, Stderr: result.Stderr, DurationMS: result.Duration.Milliseconds(), StartedAt: result.StartedAt}
	if runErr != nil {
		out.Error = runErr.Error()
	}
	return out
}

func (c *Client) authorize(req *http.Request) {
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
}

func buildTLSConfig(cfg config.MTLS) (*tls.Config, error) {
	if cfg.ServerCAFile == "" && cfg.CertFile == "" && cfg.KeyFile == "" {
		return nil, nil
	}
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
	if cfg.ServerCAFile != "" {
		caPEM, err := os.ReadFile(cfg.ServerCAFile)
		if err != nil {
			return nil, fmt.Errorf("read mtls ca: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(caPEM) {
			return nil, fmt.Errorf("parse mtls ca: no certificates found")
		}
		tlsConfig.RootCAs = pool
	}
	if cfg.CertFile != "" || cfg.KeyFile != "" {
		if cfg.CertFile == "" || cfg.KeyFile == "" {
			return nil, fmt.Errorf("mtls cert_file and key_file must be set together")
		}
		cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("load mtls key pair: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}
	return tlsConfig, nil
}
