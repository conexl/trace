package transport

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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
	tlsConfig, err := buildTLSConfig(cfg.MTLS)
	if err != nil {
		return nil, err
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if tlsConfig != nil {
		transport.TLSClientConfig = tlsConfig
	}
	return &HTTPClient{
		endpoint: strings.TrimRight(endpoint, "/"),
		token:    cfg.Token,
		client:   &http.Client{Timeout: 8 * time.Second, Transport: transport},
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

func buildTLSConfig(cfg config.MTLS) (*tls.Config, error) {
	if cfg.CAFile == "" && cfg.CertFile == "" && cfg.KeyFile == "" {
		return nil, nil
	}
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
	if cfg.CAFile != "" {
		caPEM, err := os.ReadFile(cfg.CAFile)
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
