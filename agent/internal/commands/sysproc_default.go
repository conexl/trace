//go:build !linux

package commands

import "os/exec"

func applyProcessSandbox(*exec.Cmd) {}
