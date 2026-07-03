//go:build linux

package commands

import (
	"os/exec"
	"syscall"
)

func applyProcessSandbox(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pdeathsig: syscall.SIGKILL}
}
