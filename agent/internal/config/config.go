package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Agent      AgentConfig     `yaml:"agent"`
	Cloud      CloudConfig     `yaml:"cloud"`
	Network    NetworkConfig   `yaml:"network"`
	Processes  []ProcessConfig `yaml:"processes"`
	LogStreams []LogStream     `yaml:"log_streams"`
	Tasks      []TaskConfig    `yaml:"tasks"`
	Remote     RemoteConfig    `yaml:"remote"`
	Hardware   HardwareConfig  `yaml:"hardware"`
	Power      PowerConfig     `yaml:"power"`
	Buffer     BufferConfig    `yaml:"buffer"`
}

type AgentConfig struct {
	ID       string        `yaml:"id"`
	Name     string        `yaml:"name"`
	Interval time.Duration `yaml:"interval"`
	Once     bool          `yaml:"once"`
}

type CloudConfig struct {
	Endpoint    string        `yaml:"endpoint"`
	Token       string        `yaml:"token"`
	Transport   string        `yaml:"transport"`
	ReplayBatch int           `yaml:"replay_batch"`
	ReplayEvery time.Duration `yaml:"replay_every"`
	MTLS        MTLS          `yaml:"mtls"`
}

type MTLS struct {
	CAFile   string `yaml:"ca_file"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

type NetworkConfig struct {
	PublicIPURL string      `yaml:"public_ip_url"`
	DNSChecks   []DNSCheck  `yaml:"dns_checks"`
	PortChecks  []PortCheck `yaml:"port_checks"`
	SpeedTests  []SpeedTest `yaml:"speed_tests"`
}

type DNSCheck struct {
	Name   string `yaml:"name"`
	Domain string `yaml:"domain"`
}

type PortCheck struct {
	Name    string        `yaml:"name"`
	Address string        `yaml:"address"`
	Timeout time.Duration `yaml:"timeout"`
}

type SpeedTest struct {
	Name     string        `yaml:"name"`
	URL      string        `yaml:"url"`
	MaxBytes int64         `yaml:"max_bytes"`
	Timeout  time.Duration `yaml:"timeout"`
}

type ProcessConfig struct {
	Name           string        `yaml:"name"`
	Match          string        `yaml:"match"`
	Service        string        `yaml:"service"`
	Critical       bool          `yaml:"critical"`
	Restart        bool          `yaml:"restart"`
	RestartCommand []string      `yaml:"restart_command"`
	GracePeriod    time.Duration `yaml:"grace_period"`
}

type LogStream struct {
	Name     string `yaml:"name"`
	Path     string `yaml:"path"`
	MaxBytes int64  `yaml:"max_bytes"`
}

type TaskConfig struct {
	Name        string        `yaml:"name"`
	Command     []string      `yaml:"command"`
	Timeout     time.Duration `yaml:"timeout"`
	Description string        `yaml:"description"`
}

type RemoteConfig struct {
	TasksEnabled bool   `yaml:"tasks_enabled"`
	ShellEnabled bool   `yaml:"shell_enabled"`
	AuditPath    string `yaml:"audit_path"`
}

type HardwareConfig struct {
	SmartDevices []string `yaml:"smart_devices"`
}

type PowerConfig struct {
	PreventSleep bool `yaml:"prevent_sleep"`
}

type BufferConfig struct {
	Path           string `yaml:"path"`
	MaxEvents      int    `yaml:"max_events"`
	MirrorToStdout bool   `yaml:"mirror_to_stdout"`
}

func Default() Config {
	return Config{
		Agent: AgentConfig{
			Name:     "homelytics-agent",
			Interval: 10 * time.Second,
		},
		Cloud: CloudConfig{
			Transport:   "none",
			ReplayBatch: 50,
			ReplayEvery: 15 * time.Second,
		},
		Network: NetworkConfig{
			PublicIPURL: "https://api.ipify.org",
		},
		Remote: RemoteConfig{
			TasksEnabled: true,
			ShellEnabled: false,
			AuditPath:    "./homelytics-audit.jsonl",
		},
		Power: PowerConfig{
			PreventSleep: false,
		},
		Buffer: BufferConfig{
			Path:           "./homelytics-buffer.jsonl",
			MaxEvents:      1000,
			MirrorToStdout: true,
		},
	}
}

func Load(path string) (Config, error) {
	cfg := Default()
	if path == "" {
		cfg.applyDefaults()
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	cfg.applyDefaults()
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c *Config) applyDefaults() {
	if c.Agent.Name == "" {
		c.Agent.Name = "homelytics-agent"
	}
	if c.Agent.Interval <= 0 {
		c.Agent.Interval = 10 * time.Second
	}
	if c.Cloud.Transport == "" {
		c.Cloud.Transport = "none"
	}
	if c.Cloud.ReplayBatch <= 0 {
		c.Cloud.ReplayBatch = 50
	}
	if c.Cloud.ReplayEvery <= 0 {
		c.Cloud.ReplayEvery = 15 * time.Second
	}
	if c.Remote.AuditPath == "" {
		c.Remote.AuditPath = "./homelytics-audit.jsonl"
	}
	if c.Network.PublicIPURL == "" {
		c.Network.PublicIPURL = "https://api.ipify.org"
	}
	if c.Buffer.Path == "" {
		c.Buffer.Path = "./homelytics-buffer.jsonl"
	}
	if c.Buffer.MaxEvents <= 0 {
		c.Buffer.MaxEvents = 1000
	}
	for i := range c.Network.PortChecks {
		if c.Network.PortChecks[i].Timeout <= 0 {
			c.Network.PortChecks[i].Timeout = 2 * time.Second
		}
	}
	for i := range c.Network.SpeedTests {
		if c.Network.SpeedTests[i].Timeout <= 0 {
			c.Network.SpeedTests[i].Timeout = 8 * time.Second
		}
		if c.Network.SpeedTests[i].MaxBytes <= 0 {
			c.Network.SpeedTests[i].MaxBytes = 5 * 1024 * 1024
		}
	}
	for i := range c.Processes {
		if c.Processes[i].GracePeriod <= 0 {
			c.Processes[i].GracePeriod = 5 * time.Second
		}
	}
	for i := range c.LogStreams {
		if c.LogStreams[i].MaxBytes <= 0 {
			c.LogStreams[i].MaxBytes = 16 * 1024
		}
	}
	for i := range c.Tasks {
		if c.Tasks[i].Timeout <= 0 {
			c.Tasks[i].Timeout = 60 * time.Second
		}
	}
}

func (c Config) Validate() error {
	if c.Remote.ShellEnabled {
		return errors.New("remote shell is not implemented yet; keep remote.shell_enabled=false")
	}
	if c.Cloud.Transport != "none" && c.Cloud.Transport != "http" {
		return fmt.Errorf("cloud transport %q is not supported", c.Cloud.Transport)
	}
	if c.Cloud.Transport == "http" && c.Cloud.Endpoint == "" {
		return errors.New("cloud endpoint is required for http transport")
	}
	for _, proc := range c.Processes {
		if proc.Name == "" {
			return errors.New("process entry is missing name")
		}
		if proc.Match == "" && proc.Service == "" {
			return fmt.Errorf("process %q needs match or service", proc.Name)
		}
		if proc.Restart && len(proc.RestartCommand) == 0 && proc.Service == "" {
			return fmt.Errorf("process %q enables restart but has no service or restart_command", proc.Name)
		}
	}
	for _, check := range c.Network.PortChecks {
		if check.Name == "" || check.Address == "" {
			return errors.New("port check needs name and address")
		}
	}
	for _, test := range c.Network.SpeedTests {
		if test.Name == "" || test.URL == "" {
			return errors.New("speed test needs name and url")
		}
	}
	for _, check := range c.Network.DNSChecks {
		if check.Name == "" || check.Domain == "" {
			return errors.New("dns check needs name and domain")
		}
	}
	for _, stream := range c.LogStreams {
		if stream.Name == "" || stream.Path == "" {
			return errors.New("log stream needs name and path")
		}
	}
	for _, task := range c.Tasks {
		if task.Name == "" || len(task.Command) == 0 {
			return errors.New("task needs name and command")
		}
	}
	return nil
}
