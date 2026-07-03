package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadAppliesDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("agent:\n  name: test-agent\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Agent.Name != "test-agent" {
		t.Fatalf("Agent.Name = %q", cfg.Agent.Name)
	}
	if cfg.Agent.Interval != 10*time.Second {
		t.Fatalf("Agent.Interval = %s", cfg.Agent.Interval)
	}
	if cfg.Network.PublicIPURL == "" {
		t.Fatal("Network.PublicIPURL was not defaulted")
	}
}

func TestValidateRejectsRestartWithoutCommandOrService(t *testing.T) {
	cfg := Default()
	cfg.Processes = []ProcessConfig{{Name: "api", Match: "api", Restart: true}}
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() expected error")
	}
}

func TestLoadDefaultsWatchdogPolicy(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("processes:\n  - name: api\n    match: api\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	proc := cfg.Processes[0]
	if proc.MaxRestarts != 3 || proc.RestartWindow != 5*time.Minute || proc.RestartCooldown != time.Minute {
		t.Fatalf("process defaults = %#v", proc)
	}
}

func TestValidateRejectsNegativeWatchdogPolicy(t *testing.T) {
	cfg := Default()
	cfg.Processes = []ProcessConfig{{Name: "api", Match: "api", MaxRestarts: -1}}
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() expected max_restarts error")
	}
}
