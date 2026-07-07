package services

import (
	"context"
	"testing"

	"agent/internal/collectors"
	"agent/internal/config"
	"agent/internal/tasksclient"
)

func TestRunDNSRecheckResolvesDomains(t *testing.T) {
	agent := &Agent{
		cfg:     config.Config{Network: config.NetworkConfig{PublicIPURL: ""}},
		network: collectors.NewNetworkCollector(),
	}
	result, err := agent.runDNSRecheck(context.Background(), tasksclient.TaskPayload{Domains: []string{"localhost"}})
	if err != nil {
		t.Fatalf("runDNSRecheck() error = %v", err)
	}
	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", result.ExitCode, result.Error)
	}
	if result.Stdout == "" {
		t.Fatalf("expected stdout with dns results")
	}
}

func TestRunDNSRecheckRejectsEmptyDomains(t *testing.T) {
	agent := &Agent{
		cfg:     config.Config{Network: config.NetworkConfig{PublicIPURL: ""}},
		network: collectors.NewNetworkCollector(),
	}
	result, err := agent.runDNSRecheck(context.Background(), tasksclient.TaskPayload{})
	if err == nil {
		t.Fatalf("expected error for empty domains")
	}
	if result.ExitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", result.ExitCode)
	}
}
