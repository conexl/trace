package collectors

import (
	"context"
	"errors"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type ServiceManager interface {
	Status(ctx context.Context, name string) (string, bool, error)
	Restart(ctx context.Context, name string) error
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

func (systemdManager) Status(ctx context.Context, name string) (string, bool, error) {
	cmd := exec.CommandContext(ctx, "systemctl", "is-active", name)
	out, err := cmd.CombinedOutput()
	status := strings.TrimSpace(string(out))
	return status, status == "active", err
}

func (systemdManager) Restart(ctx context.Context, name string) error {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	return exec.CommandContext(ctx, "systemctl", "restart", name).Run()
}

type launchdManager struct{}

func (launchdManager) Status(ctx context.Context, name string) (string, bool, error) {
	cmd := exec.CommandContext(ctx, "launchctl", "print", name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return strings.TrimSpace(string(out)), false, err
	}
	return "loaded", true, nil
}

func (launchdManager) Restart(ctx context.Context, name string) error {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	if err := exec.CommandContext(ctx, "launchctl", "kickstart", "-k", name).Run(); err != nil {
		return err
	}
	return nil
}

type noopServiceManager struct{}

func (noopServiceManager) Status(context.Context, string) (string, bool, error) {
	return "unsupported", false, errors.New("no supported service manager found")
}

func (noopServiceManager) Restart(context.Context, string) error {
	return errors.New("no supported service manager found")
}

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
