package collectors

import (
	"context"
	"errors"
	"testing"
	"time"

	"agent/internal/config"
)

func TestProcessCollectorRestartsCriticalServiceOnceThenSuppresses(t *testing.T) {
	manager := &fakeServiceManager{status: ServiceStatus{Status: "failed", Running: false, ExitCode: 7}}
	collector := NewProcessCollector(manager)
	now := time.Date(2026, 7, 3, 12, 0, 0, 0, time.UTC)
	collector.clock = func() time.Time { return now }
	cfg := []config.ProcessConfig{{
		Name:            "nginx",
		Service:         "nginx",
		Critical:        true,
		Restart:         true,
		GracePeriod:     time.Second,
		MaxRestarts:     1,
		RestartWindow:   time.Minute,
		RestartCooldown: 5 * time.Minute,
	}}

	processes, events := collector.Collect(context.Background(), cfg)
	if len(processes) != 1 || processes[0].LastExitCode != 7 {
		t.Fatalf("processes = %#v", processes)
	}
	if manager.restarts != 1 {
		t.Fatalf("restarts = %d", manager.restarts)
	}
	if len(events) != 2 || events[0].Type != "process.down" || events[1].Type != "process.restarted" {
		t.Fatalf("events = %#v", events)
	}
	if events[0].ExitCode != 7 || events[1].ExitCode != 7 {
		t.Fatalf("exit codes = %#v", events)
	}

	now = now.Add(10 * time.Second)
	_, events = collector.Collect(context.Background(), cfg)
	if manager.restarts != 1 {
		t.Fatalf("unexpected restart count = %d", manager.restarts)
	}
	if len(events) != 2 || events[1].Type != "process.restart_suppressed" {
		t.Fatalf("events = %#v", events)
	}
}

func TestProcessCollectorReportsRestartFailure(t *testing.T) {
	manager := &fakeServiceManager{status: ServiceStatus{Status: "failed", Running: false}, restartErr: errors.New("permission denied")}
	collector := NewProcessCollector(manager)
	collector.clock = func() time.Time { return time.Date(2026, 7, 3, 12, 0, 0, 0, time.UTC) }
	cfg := []config.ProcessConfig{{Name: "api", Service: "api", Critical: true, Restart: true, GracePeriod: time.Second, MaxRestarts: 3, RestartWindow: time.Minute, RestartCooldown: time.Minute}}

	_, events := collector.Collect(context.Background(), cfg)
	if len(events) != 2 || events[1].Type != "process.restart_failed" || events[1].Severity != "critical" {
		t.Fatalf("events = %#v", events)
	}
	if events[1].Action != "service restart api" {
		t.Fatalf("action = %q", events[1].Action)
	}
}

type fakeServiceManager struct {
	status     ServiceStatus
	restartErr error
	restarts   int
	starts     int
	stops      int
}

func (m *fakeServiceManager) Status(context.Context, string) (ServiceStatus, error) {
	return m.status, nil
}

func (m *fakeServiceManager) Start(context.Context, string) error {
	m.starts++
	return nil
}

func (m *fakeServiceManager) Stop(context.Context, string) error {
	m.stops++
	return nil
}

func (m *fakeServiceManager) Restart(context.Context, string) error {
	m.restarts++
	return m.restartErr
}
