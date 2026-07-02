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
