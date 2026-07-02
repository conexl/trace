package remote

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"agent/internal/config"

	"github.com/creack/pty"
)

var ErrShellDisabled = errors.New("remote shell is disabled by policy")

type ShellSession struct {
	cmd  *exec.Cmd
	file *os.File
	once sync.Once
}

type ShellOptions struct {
	Command []string
	Env     []string
	Dir     string
}

func StartShell(ctx context.Context, cfg config.RemoteConfig, opts ShellOptions) (*ShellSession, error) {
	if !cfg.ShellEnabled {
		return nil, ErrShellDisabled
	}
	command := opts.Command
	if len(command) == 0 {
		command = []string{defaultShell()}
	}
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	cmd.Env = append(os.Environ(), opts.Env...)
	cmd.Dir = opts.Dir
	file, err := pty.Start(cmd)
	if err != nil {
		return nil, fmt.Errorf("start pty: %w", err)
	}
	session := &ShellSession{cmd: cmd, file: file}
	go func() {
		<-ctx.Done()
		_ = session.Close()
	}()
	return session, nil
}

func (s *ShellSession) Read(p []byte) (int, error) {
	if s == nil || s.file == nil {
		return 0, io.ErrClosedPipe
	}
	return s.file.Read(p)
}

func (s *ShellSession) Write(p []byte) (int, error) {
	if s == nil || s.file == nil {
		return 0, io.ErrClosedPipe
	}
	return s.file.Write(p)
}

func (s *ShellSession) Close() error {
	var err error
	s.once.Do(func() {
		if s.file != nil {
			err = s.file.Close()
		}
		if s.cmd != nil && s.cmd.Process != nil {
			_ = s.cmd.Process.Kill()
			_, _ = s.cmd.Process.Wait()
		}
	})
	return err
}

func defaultShell() string {
	if shell := os.Getenv("SHELL"); shell != "" {
		return shell
	}
	return "/bin/sh"
}
