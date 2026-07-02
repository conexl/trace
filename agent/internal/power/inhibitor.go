package power

import (
	"context"
	"errors"
	"os/exec"
	"runtime"
)

type Inhibitor struct {
	cmd *exec.Cmd
}

func Start(ctx context.Context, enabled bool) (*Inhibitor, error) {
	if !enabled {
		return &Inhibitor{}, nil
	}
	if runtime.GOOS == "linux" {
		if _, err := exec.LookPath("systemd-inhibit"); err == nil {
			cmd := exec.CommandContext(ctx, "systemd-inhibit", "--what=sleep", "--why=Homelytics agent keeps server awake", "sleep", "infinity")
			if err := cmd.Start(); err != nil {
				return nil, err
			}
			return &Inhibitor{cmd: cmd}, nil
		}
	}
	if runtime.GOOS == "darwin" {
		if _, err := exec.LookPath("caffeinate"); err == nil {
			cmd := exec.CommandContext(ctx, "caffeinate", "-dimsu")
			if err := cmd.Start(); err != nil {
				return nil, err
			}
			return &Inhibitor{cmd: cmd}, nil
		}
	}
	return nil, errors.New("prevent_sleep requested but no supported inhibitor command found")
}

func (i *Inhibitor) Stop() error {
	if i == nil || i.cmd == nil || i.cmd.Process == nil {
		return nil
	}
	return i.cmd.Process.Kill()
}
