package services

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"agent/internal/collectors"
	"agent/internal/config"
	"agent/internal/tasksclient"
)

func TestRunDiagnosticsReturnsSafeJSONBundle(t *testing.T) {
	manager := &fakeServiceManager{}
	agent := &Agent{
		cfg: config.Config{
			Agent:     config.AgentConfig{Name: "devbox", Revision: 7},
			Network:   config.NetworkConfig{PublicIPURL: ""},
			Processes: []config.ProcessConfig{{Name: "nginx", Service: "nginx", RemoteControl: true}},
		},
		system:         collectors.NewSystemCollector(),
		network:        collectors.NewNetworkCollector(),
		processes:      collectors.NewProcessCollector(manager),
		hardware:       collectors.NewHardwareCollector(),
		serviceManager: manager,
		startTime:      time.Now().Add(-time.Minute),
	}

	result, err := agent.runDiagnostics(context.Background(), tasksclient.TaskPayload{
		Service:    "nginx",
		IncidentID: "incident-1",
	})
	if err != nil {
		t.Fatalf("runDiagnostics() error = %v", err)
	}
	if result.ExitCode != 0 || result.Stdout == "" {
		t.Fatalf("result = %#v", result)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(result.Stdout), &payload); err != nil {
		t.Fatalf("diagnostics JSON invalid: %v", err)
	}
	if payload["incident_id"] != "incident-1" || payload["service"] != "nginx" {
		t.Fatalf("payload = %#v", payload)
	}
}
