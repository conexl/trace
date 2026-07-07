//go:build linux

package commands

import (
	"os/exec"
	"os/user"
	"strconv"
	"syscall"
)

func applyProcessSandbox(cmd *exec.Cmd, userName string) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pdeathsig: syscall.SIGKILL}
	if userName != "" {
		if u, err := user.Lookup(userName); err == nil {
			uid, _ := strconv.Atoi(u.Uid)
			gid, _ := strconv.Atoi(u.Gid)
			cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}
		}
	}
}
