package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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
	Update     UpdateConfig    `yaml:"update"`
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
	Name            string        `yaml:"name"`
	Match           string        `yaml:"match"`
	Service         string        `yaml:"service"`
	Critical        bool          `yaml:"critical"`
	Restart         bool          `yaml:"restart"`
	RemoteControl   bool          `yaml:"remote_control"`
	RestartCommand  []string      `yaml:"restart_command"`
	GracePeriod     time.Duration `yaml:"grace_period"`
	MaxRestarts     int           `yaml:"max_restarts"`
	RestartWindow   time.Duration `yaml:"restart_window"`
	RestartCooldown time.Duration `yaml:"restart_cooldown"`
}

type LogStream struct {
	Name     string `yaml:"name"`
	Path     string `yaml:"path"`
	MaxBytes int64  `yaml:"max_bytes"`
}

type TaskConfig struct {
	Name           string            `yaml:"name"`
	Command        []string          `yaml:"command"`
	Timeout        time.Duration     `yaml:"timeout"`
	Description    string            `yaml:"description"`
	WorkingDir     string            `yaml:"working_dir"`
	Env            map[string]string `yaml:"env"`
	MaxOutputBytes int64             `yaml:"max_output_bytes"`
}

type RemoteConfig struct {
	TasksEnabled bool          `yaml:"tasks_enabled"`
	ShellEnabled bool          `yaml:"shell_enabled"`
	AuditPath    string        `yaml:"audit_path"`
	PollEvery    time.Duration `yaml:"poll_every"`
}

type UpdateConfig struct {
	URL              string `yaml:"url"`
	SHA256           string `yaml:"sha256"`
	SignatureURL     string `yaml:"signature_url"`
	Ed25519PublicKey string `yaml:"ed25519_public_key"`
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
			PollEvery:    15 * time.Second,
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
	if c.Remote.PollEvery <= 0 {
		c.Remote.PollEvery = 15 * time.Second
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
		if c.Processes[i].MaxRestarts == 0 {
			c.Processes[i].MaxRestarts = 3
		}
		if c.Processes[i].RestartWindow == 0 {
			c.Processes[i].RestartWindow = 5 * time.Minute
		}
		if c.Processes[i].RestartCooldown == 0 {
			c.Processes[i].RestartCooldown = time.Minute
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
		if c.Tasks[i].MaxOutputBytes == 0 {
			c.Tasks[i].MaxOutputBytes = 64 * 1024
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
		if proc.MaxRestarts < 0 {
			return fmt.Errorf("process %q has negative max_restarts", proc.Name)
		}
		if proc.RestartWindow < 0 {
			return fmt.Errorf("process %q has negative restart_window", proc.Name)
		}
		if proc.RestartCooldown < 0 {
			return fmt.Errorf("process %q has negative restart_cooldown", proc.Name)
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
		if err := validateTaskSandbox(task); err != nil {
			return err
		}
	}
	return nil
}

var taskEnvNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func validateTaskSandbox(task TaskConfig) error {
	command := filepath.Base(task.Command[0])
	switch strings.ToLower(command) {
	case "sh", "bash", "zsh", "fish", "pwsh", "powershell", "cmd", "cmd.exe":
		return fmt.Errorf("task %q uses forbidden shell command %q", task.Name, command)
	}
	if task.WorkingDir != "" {
		cleaned := filepath.Clean(task.WorkingDir)
		if !filepath.IsAbs(cleaned) {
			return fmt.Errorf("task %q working_dir must be absolute", task.Name)
		}
		if cleaned != task.WorkingDir {
			return fmt.Errorf("task %q working_dir must be clean", task.Name)
		}
	}
	for key := range task.Env {
		if !taskEnvNamePattern.MatchString(key) {
			return fmt.Errorf("task %q has invalid env name %q", task.Name, key)
		}
		switch key {
		case "PATH", "LD_PRELOAD", "DYLD_INSERT_LIBRARIES":
			return fmt.Errorf("task %q cannot override env %q", task.Name, key)
		}
	}
	if task.MaxOutputBytes < 0 {
		return fmt.Errorf("task %q has negative max_output_bytes", task.Name)
	}
	return nil
}
