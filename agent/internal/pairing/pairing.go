package pairing

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"agent/internal/config"
)

type Client struct {
	endpoint string
	token    string
	client   *http.Client
}

type Request struct {
	AgentName string `json:"agent_name"`
	Hostname  string `json:"hostname"`
}

type Response struct {
	AgentID     string    `json:"agent_id"`
	Certificate string    `json:"certificate_pem"`
	PrivateKey  string    `json:"private_key_pem"`
	CACert      string    `json:"ca_cert_pem"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type SaveOptions struct {
	Dir      string
	CAFile   string
	CertFile string
	KeyFile  string
}

type SavedCredentials struct {
	CAFile   string `json:"ca_file"`
	CertFile string `json:"cert_file"`
	KeyFile  string `json:"key_file"`
}

func NewClient(cfg config.CloudConfig) (*Client, error) {
	endpoint := strings.TrimSpace(cfg.Endpoint)
	if endpoint == "" {
		return nil, fmt.Errorf("cloud endpoint is empty")
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if cfg.MTLS.ServerCAFile != "" {
		caPEM, err := os.ReadFile(cfg.MTLS.ServerCAFile)
		if err != nil {
			// The installer preconfigures the destination paths before the first
			// pairing call. Those files do not exist until Claim succeeds.
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("read pairing server ca file: %w", err)
			}
		} else {
			pool := x509.NewCertPool()
			if !pool.AppendCertsFromPEM(caPEM) {
				return nil, fmt.Errorf("parse pairing server ca file: no certificates found")
			}
			transport.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12, RootCAs: pool}
		}
	}
	return &Client{endpoint: strings.TrimRight(endpoint, "/"), token: cfg.Token, client: &http.Client{Timeout: 10 * time.Second, Transport: transport}}, nil
}

func (c *Client) Claim(ctx context.Context, req Request) (Response, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return Response{}, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+"/v1/pairing/claim", bytes.NewReader(payload))
	if err != nil {
		return Response{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return Response{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Response{}, fmt.Errorf("pairing failed: %s", resp.Status)
	}
	var out Response
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return Response{}, err
	}
	if out.Certificate == "" || out.PrivateKey == "" || out.CACert == "" {
		return Response{}, fmt.Errorf("pairing response missing credentials")
	}
	return out, nil
}

func SaveCredentials(resp Response, opts SaveOptions) (SavedCredentials, error) {
	dir := opts.Dir
	if dir == "" {
		var err error
		dir, err = os.UserConfigDir()
		if err != nil {
			return SavedCredentials{}, err
		}
		dir = filepath.Join(dir, "homelytics", "certs")
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return SavedCredentials{}, err
	}
	saved := SavedCredentials{
		CAFile:   firstNonEmpty(opts.CAFile, filepath.Join(dir, "ca.pem")),
		CertFile: firstNonEmpty(opts.CertFile, filepath.Join(dir, "agent.pem")),
		KeyFile:  firstNonEmpty(opts.KeyFile, filepath.Join(dir, "agent-key.pem")),
	}
	if err := writeSecretFile(saved.CAFile, []byte(resp.CACert)); err != nil {
		return SavedCredentials{}, err
	}
	if err := writeSecretFile(saved.CertFile, []byte(resp.Certificate)); err != nil {
		return SavedCredentials{}, err
	}
	if err := writeSecretFile(saved.KeyFile, []byte(resp.PrivateKey)); err != nil {
		return SavedCredentials{}, err
	}
	return saved, nil
}

func writeSecretFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
