package collectors

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type ServiceStatus struct {
	Status   string
	Running  bool
	ExitCode int
	Output   string
}

type ServiceManager interface {
	Status(ctx context.Context, name string) (ServiceStatus, error)
	Start(ctx context.Context, name string) error
	Stop(ctx context.Context, name string) error
	Restart(ctx context.Context, name string) error
	ListServices(ctx context.Context) ([]string, error)
}

func NewServiceManager() ServiceManager {
	if runtime.GOOS == "linux" && commandExists("systemctl") {
		return systemdManager{}
	}
	if runtime.GOOS == "darwin" && commandExists("launchctl") {
		return launchdManager{}
	}
	return noopServiceManager{}
}

type systemdManager struct{}

func (systemdManager) Status(ctx context.Context, name string) (ServiceStatus, error) {
	cmd := exec.CommandContext(ctx, "systemctl", "show", name, "--property=ActiveState", "--property=SubState", "--property=ExecMainStatus", "--value")
	out, err := cmd.CombinedOutput()
	lines := splitNonEmptyLines(string(out))
	status := ServiceStatus{Output: strings.TrimSpace(string(out))}
	if len(lines) > 0 {
		status.Status = lines[0]
	}
	if len(lines) > 1 && lines[1] != status.Status {
		status.Status = status.Status + "/" + lines[1]
	}
	if len(lines) > 2 {
		if code, parseErr := strconv.Atoi(lines[2]); parseErr == nil {
			status.ExitCode = code
		}
	}
	status.Running = strings.HasPrefix(status.Status, "active")
	if status.Status == "" {
		fallback := exec.CommandContext(ctx, "systemctl", "is-active", name)
		fallbackOut, fallbackErr := fallback.CombinedOutput()
		status.Status = strings.TrimSpace(string(fallbackOut))
		status.Output = status.Status
		status.Running = status.Status == "active"
		if fallbackErr != nil {
			return status, fallbackErr
		}
	}
	return status, err
}

func (systemdManager) Start(ctx context.Context, name string) error {
	return runServiceCommand(ctx, "systemctl", "start", name)
}

func (systemdManager) Stop(ctx context.Context, name string) error {
	return runServiceCommand(ctx, "systemctl", "stop", name)
}

func (systemdManager) Restart(ctx context.Context, name string) error {
	return runServiceCommand(ctx, "systemctl", "restart", name)
}

func (systemdManager) ListServices(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "systemctl", "list-unit-files", "--type=service", "--no-legend", "--no-pager")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	lines := splitNonEmptyLines(string(out))
	services := make([]string, 0, len(lines))
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) > 0 {
			services = append(services, parts[0])
		}
	}
	return services, nil
}

type launchdManager struct{}

func (launchdManager) Status(ctx context.Context, name string) (ServiceStatus, error) {
	cmd := exec.CommandContext(ctx, "launchctl", "print", name)
	out, err := cmd.CombinedOutput()
	status := ServiceStatus{Status: "loaded", Running: err == nil, Output: strings.TrimSpace(string(out))}
	if err != nil {
		status.Status = "unloaded"
	}
	return status, err
}

func (launchdManager) Start(ctx context.Context, name string) error {
	return runServiceCommand(ctx, "launchctl", "kickstart", name)
}

func (launchdManager) Stop(ctx context.Context, name string) error {
	return runServiceCommand(ctx, "launchctl", "bootout", name)
}

func (launchdManager) Restart(ctx context.Context, name string) error {
	return runServiceCommand(ctx, "launchctl", "kickstart", "-k", name)
}

func (launchdManager) ListServices(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "launchctl", "list")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	lines := splitNonEmptyLines(string(out))
	services := make([]string, 0, len(lines))
	for i, line := range lines {
		if i == 0 { // Skip header
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 3 {
			services = append(services, parts[2])
		}
	}
	return services, nil
}

type noopServiceManager struct{}

func (noopServiceManager) Status(context.Context, string) (ServiceStatus, error) {
	return ServiceStatus{Status: "unsupported", Running: false}, errors.New("no supported service manager found")
}

func (noopServiceManager) Start(context.Context, string) error {
	return errors.New("no supported service manager found")
}

func (noopServiceManager) Stop(context.Context, string) error {
	return errors.New("no supported service manager found")
}

func (noopServiceManager) Restart(context.Context, string) error {
	return errors.New("no supported service manager found")
}

func (noopServiceManager) ListServices(context.Context) ([]string, error) {
	return nil, errors.New("no supported service manager found")
}

func runServiceCommand(ctx context.Context, name string, args ...string) error {
	cmdCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	out, err := exec.CommandContext(cmdCtx, name, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s failed: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return nil
}

func splitNonEmptyLines(value string) []string {
	raw := strings.Split(value, "\n")
	lines := make([]string, 0, len(raw))
	for _, line := range raw {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
