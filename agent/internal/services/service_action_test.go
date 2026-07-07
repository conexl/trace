package services

import (
	"context"
	"testing"

	"agent/internal/collectors"
	"agent/internal/config"
	"agent/internal/tasksclient"
)

func TestRunServiceActionRequiresRemoteControl(t *testing.T) {
	agent := &Agent{
		cfg:            config.Config{Processes: []config.ProcessConfig{{Name: "nginx", Service: "nginx", RemoteControl: false}}},
		serviceManager: &fakeServiceManager{},
	}
	result, err := agent.runServiceAction(context.Background(), tasksclient.TaskPayload{Service: "nginx", Action: "restart"})
	if err == nil || result.ExitCode == 0 {
		t.Fatalf("result=%#v err=%v", result, err)
	}
}

func TestRunServiceActionRestartsAllowedService(t *testing.T) {
	manager := &fakeServiceManager{}
	agent := &Agent{
		cfg:            config.Config{Processes: []config.ProcessConfig{{Name: "nginx", Service: "nginx", RemoteControl: true}}},
		serviceManager: manager,
	}
	result, err := agent.runServiceAction(context.Background(), tasksclient.TaskPayload{Service: "nginx", Action: "restart"})
	if err != nil {
		t.Fatalf("runServiceAction() error = %v", err)
	}
	if result.ExitCode != 0 || manager.restarts != 1 {
		t.Fatalf("result=%#v restarts=%d", result, manager.restarts)
	}
}

type fakeServiceManager struct {
	restarts int
	starts   int
	stops    int
	err      error
}

func (m *fakeServiceManager) Status(context.Context, string) (collectors.ServiceStatus, error) {
	return collectors.ServiceStatus{Status: "active", Running: true}, nil
}

func (m *fakeServiceManager) Start(context.Context, string) error {
	m.starts++
	return m.err
}

func (m *fakeServiceManager) Stop(context.Context, string) error {
	m.stops++
	return m.err
}

func (m *fakeServiceManager) Restart(context.Context, string) error {
	m.restarts++
	if m.err != nil {
		return m.err
	}
	return nil
}

func (m *fakeServiceManager) ListServices(context.Context) ([]string, error) {
	return []string{"nginx"}, nil
}
