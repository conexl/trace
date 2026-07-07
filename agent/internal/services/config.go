package services

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"agent/internal/config"
	"agent/internal/configclient"

	"gopkg.in/yaml.v3"
)

func (a *Agent) applyDesiredConfig(path string, desired configclient.DesiredConfig) (bool, error) {
	if path == "" {
		return false, fmt.Errorf("config path is empty")
	}
	base := a.cfg
	if _, err := os.ReadFile(path); err == nil {
		if loaded, err := config.Load(path); err == nil {
			base = loaded
		} else {
			slog.Warn("failed to load existing config for merge, using runtime config", "error", err)
		}
	}
	cfg := desiredToConfig(base, desired)
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return false, fmt.Errorf("marshal config: %w", err)
	}
	current, err := os.ReadFile(path)
	if err == nil && string(current) == string(data) {
		return false, nil
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return false, fmt.Errorf("write config: %w", err)
	}
	slog.Info("agent config updated from server", "path", path)
	return true, nil
}

func desiredToConfig(base config.Config, desired configclient.DesiredConfig) config.Config {
	cfg := base
	if desired.Agent.Name != "" {
		cfg.Agent.Name = desired.Agent.Name
	}
	if desired.Agent.Interval > 0 {
		cfg.Agent.Interval = desired.Agent.Interval
	}
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "INFO"
	}
	if desired.Logging.Level != "" {
		cfg.Logging.Level = desired.Logging.Level
	}
	cfg.Watchdog = config.WatchdogConfig{
		PollingSeconds: desired.Watchdog.PollingSeconds,
		TimeoutSeconds: desired.Watchdog.TimeoutSeconds,
	}
	if cfg.Watchdog.PollingSeconds <= 0 {
		cfg.Watchdog.PollingSeconds = 10
	}
	if cfg.Watchdog.TimeoutSeconds <= 0 {
		cfg.Watchdog.TimeoutSeconds = 30
	}
	cfg.Performance = config.PerformanceConfig{
		Mode:     desired.Performance.Mode,
		FanCurve: desired.Performance.FanCurve,
	}
	if cfg.Performance.Mode == "" {
		cfg.Performance.Mode = "balanced"
	}
	if cfg.Performance.FanCurve == "" {
		cfg.Performance.FanCurve = "auto"
	}
	if desired.Network.PublicIPURL != "" {
		cfg.Network.PublicIPURL = desired.Network.PublicIPURL
	}
	cfg.Network.DNSChecks = make([]config.DNSCheck, len(desired.Network.DNSChecks))
	for i, c := range desired.Network.DNSChecks {
		cfg.Network.DNSChecks[i] = config.DNSCheck{Name: c.Name, Domain: c.Domain, Group: c.Group, Critical: c.Critical}
	}
	cfg.Network.PortChecks = make([]config.PortCheck, len(desired.Network.PortChecks))
	for i, c := range desired.Network.PortChecks {
		cfg.Network.PortChecks[i] = config.PortCheck{Name: c.Name, Address: c.Address, Timeout: c.Timeout}
	}
	cfg.Network.SpeedTests = make([]config.SpeedTest, len(desired.Network.SpeedTests))
	for i, c := range desired.Network.SpeedTests {
		cfg.Network.SpeedTests[i] = config.SpeedTest{Name: c.Name, URL: c.URL, MaxBytes: c.MaxBytes, Timeout: c.Timeout}
	}
	cfg.Processes = make([]config.ProcessConfig, len(desired.Processes))
	for i, p := range desired.Processes {
		cfg.Processes[i] = config.ProcessConfig{
			Name:            p.Name,
			Match:           p.Match,
			Service:         p.Service,
			Critical:        p.Critical,
			Restart:         p.Restart,
			RemoteControl:   p.RemoteControl,
			RestartCommand:  p.RestartCommand,
			GracePeriod:     p.GracePeriod,
			MaxRestarts:     p.MaxRestarts,
			RestartWindow:   p.RestartWindow,
			RestartCooldown: p.RestartCooldown,
			CPUThreshold:    p.CPUThreshold,
			MemoryThreshold: p.MemoryThreshold,
		}
	}
	cfg.LogStreams = make([]config.LogStream, len(desired.LogStreams))
	for i, l := range desired.LogStreams {
		cfg.LogStreams[i] = config.LogStream{Name: l.Name, Path: l.Path, MaxBytes: l.MaxBytes}
	}
	cfg.Remote = config.RemoteConfig{
		TasksEnabled: desired.Remote.TasksEnabled,
		ShellEnabled: desired.Remote.ShellEnabled,
		AuditPath:    desired.Remote.AuditPath,
		PollEvery:    desired.Remote.PollEvery,
	}
	if cfg.Remote.AuditPath == "" {
		cfg.Remote.AuditPath = cfg.Buffer.Path
	}
	if cfg.Remote.PollEvery <= 0 {
		cfg.Remote.PollEvery = 15 * time.Second
	}
	cfg.Update = config.UpdateConfig{
		Policy:           desired.Update.Policy,
		URL:              desired.Update.URL,
		SHA256:           desired.Update.SHA256,
		SignatureURL:     desired.Update.SignatureURL,
		Ed25519PublicKey: desired.Update.Ed25519PublicKey,
	}
	if cfg.Update.Policy == "" {
		cfg.Update.Policy = "check"
	}
	cfg.Hardware = config.HardwareConfig{SmartDevices: desired.Hardware.SmartDevices}
	cfg.Power = config.PowerConfig{
		PreventSleep: desired.Power.PreventSleep,
		SleepAt:      desired.Power.SleepAt,
		WakeAt:       desired.Power.WakeAt,
	}
	if desired.Buffer.Path != "" {
		cfg.Buffer.Path = desired.Buffer.Path
	}
	if desired.Buffer.MaxEvents > 0 {
		cfg.Buffer.MaxEvents = desired.Buffer.MaxEvents
	}
	cfg.Buffer.MirrorToStdout = desired.Buffer.MirrorToStdout
	return cfg
}
