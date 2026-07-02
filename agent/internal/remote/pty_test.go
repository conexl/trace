package remote

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"agent/internal/config"
)

func TestStartShellRejectsDisabledPolicy(t *testing.T) {
	_, err := StartShell(context.Background(), config.RemoteConfig{ShellEnabled: false}, ShellOptions{Command: []string{"/bin/sh"}})
	if !errors.Is(err, ErrShellDisabled) {
		t.Fatalf("err = %v", err)
	}
}

func TestStartShellRunsCommandInPTY(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	session, err := StartShell(ctx, config.RemoteConfig{ShellEnabled: true}, ShellOptions{Command: []string{"/bin/sh"}})
	if err != nil {
		t.Fatalf("StartShell() error = %v", err)
	}
	defer session.Close()
	if _, err := session.Write([]byte("printf ready; exit\n")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	buf := make([]byte, 256)
	var output strings.Builder
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		n, err := session.Read(buf)
		if n > 0 {
			output.Write(buf[:n])
			if strings.Contains(output.String(), "ready") {
				return
			}
		}
		if err != nil {
			break
		}
	}
	t.Fatalf("session output = %q", output.String())
}
