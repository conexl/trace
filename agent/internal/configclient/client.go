package configclient

import (
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

	"agent/internal/config"
)

type DesiredConfig struct {
	Agent       AgentConfig       `json:"agent"`
	Logging     LoggingConfig     `json:"logging"`
	Watchdog    WatchdogConfig    `json:"watchdog"`
	Performance PerformanceConfig `json:"performance"`
	Network     NetworkConfig     `json:"network"`
	Processes   []ProcessConfig   `json:"processes"`
	LogStreams  []LogStream       `json:"log_streams"`
	Remote      RemoteConfig      `json:"remote"`
	Update      UpdateConfig      `json:"update"`
	Hardware    HardwareConfig    `json:"hardware"`
	Power       PowerConfig       `json:"power"`
	Buffer      BufferConfig      `json:"buffer"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type AgentConfig struct {
	Name     string        `json:"name"`
	Interval time.Duration `json:"interval"`
}

type LoggingConfig struct {
	Level string `json:"level"`
}

type WatchdogConfig struct {
	PollingSeconds int `json:"polling_seconds"`
	TimeoutSeconds int `json:"timeout_seconds"`
}

type PerformanceConfig struct {
	Mode     string `json:"mode"`
	FanCurve string `json:"fan_curve"`
}

type NetworkConfig struct {
	PublicIPURL string      `json:"public_ip_url"`
	DNSChecks   []DNSCheck  `json:"dns_checks"`
	PortChecks  []PortCheck `json:"port_checks"`
	SpeedTests  []SpeedTest `json:"speed_tests"`
}

type DNSCheck struct {
	Name     string `json:"name"`
	Domain   string `json:"domain"`
	Group    string `json:"group,omitempty"`
	Critical bool   `json:"critical,omitempty"`
}

type PortCheck struct {
	Name    string        `json:"name"`
	Address string        `json:"address"`
	Timeout time.Duration `json:"timeout"`
}

type SpeedTest struct {
	Name     string        `json:"name"`
	URL      string        `json:"url"`
	MaxBytes int64         `json:"max_bytes"`
	Timeout  time.Duration `json:"timeout"`
}

type ProcessConfig struct {
	Name            string        `json:"name"`
	Match           string        `json:"match"`
	Service         string        `json:"service"`
	Critical        bool          `json:"critical"`
	Restart         bool          `json:"restart"`
	RemoteControl   bool          `json:"remote_control"`
	RestartCommand  []string      `json:"restart_command"`
	GracePeriod     time.Duration `json:"grace_period"`
	MaxRestarts     int           `json:"max_restarts"`
	RestartWindow   time.Duration `json:"restart_window"`
	RestartCooldown time.Duration `json:"restart_cooldown"`
	CPUThreshold    int           `json:"cpu_threshold"`
	MemoryThreshold int           `json:"memory_threshold"`
}

type LogStream struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	MaxBytes int64  `json:"max_bytes"`
}

type RemoteConfig struct {
	TasksEnabled bool          `json:"tasks_enabled"`
	ShellEnabled bool          `json:"shell_enabled"`
	AuditPath    string        `json:"audit_path"`
	PollEvery    time.Duration `json:"poll_every"`
}

type UpdateConfig struct {
	Policy           string `json:"policy"`
	URL              string `json:"url"`
	SHA256           string `json:"sha256"`
	SignatureURL     string `json:"signature_url"`
	Ed25519PublicKey string `json:"ed25519_public_key"`
}

type HardwareConfig struct {
	SmartDevices []string `json:"smart_devices"`
}

type PowerConfig struct {
	PreventSleep bool   `json:"prevent_sleep"`
	SleepAt      string `json:"sleep_at,omitempty"`
	WakeAt       string `json:"wake_at,omitempty"`
}

type BufferConfig struct {
	Path           string `json:"path"`
	MaxEvents      int    `json:"max_events"`
	MirrorToStdout bool   `json:"mirror_to_stdout"`
}

type Client struct {
	endpoint string
	agentID  string
	token    string
	client   *http.Client
}

func New(cfg config.CloudConfig, agentID string) (*Client, error) {
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
	return &Client{
		endpoint: strings.TrimRight(endpoint, "/"),
		agentID:  agentID,
		token:    cfg.Token,
		client:   &http.Client{Timeout: 8 * time.Second, Transport: transport},
	}, nil
}

func (c *Client) Fetch(ctx context.Context) (DesiredConfig, error) {
	values := url.Values{}
	values.Set("agent_id", c.agentID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint+"/v1/agent/config?"+values.Encode(), nil)
	if err != nil {
		return DesiredConfig{}, err
	}
	c.authorize(req)
	resp, err := c.client.Do(req)
	if err != nil {
		return DesiredConfig{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return DesiredConfig{}, fmt.Errorf("no desired config on server")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return DesiredConfig{}, fmt.Errorf("fetch config failed: %s", resp.Status)
	}
	var cfg DesiredConfig
	if err := json.NewDecoder(resp.Body).Decode(&cfg); err != nil {
		return DesiredConfig{}, err
	}
	return cfg, nil
}

func (c *Client) authorize(req *http.Request) {
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
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
