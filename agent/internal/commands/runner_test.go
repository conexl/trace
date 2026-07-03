package commands

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"agent/internal/config"
)

func TestRunnerRejectsUnknownTaskAndAudits(t *testing.T) {
	auditPath := filepath.Join(t.TempDir(), "audit.jsonl")
	runner := NewRunner(config.Config{Remote: config.RemoteConfig{TasksEnabled: true, AuditPath: auditPath}})

	if _, err := runner.Run(context.Background(), "missing"); err == nil {
		t.Fatal("Run() expected error")
	}
	data, err := os.ReadFile(auditPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "missing") || !strings.Contains(string(data), "false") {
		t.Fatalf("audit = %s", data)
	}
}

func TestRunnerRunsConfiguredTaskWithoutShell(t *testing.T) {
	auditPath := filepath.Join(t.TempDir(), "audit.jsonl")
	runner := NewRunner(config.Config{
		Remote: config.RemoteConfig{TasksEnabled: true, AuditPath: auditPath},
		Tasks:  []config.TaskConfig{{Name: "echo", Command: []string{"printf", "hello"}, Timeout: time.Second}},
	})

	result, err := runner.Run(context.Background(), "echo")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Stdout != "hello" || result.ExitCode != 0 {
		t.Fatalf("result = %#v", result)
	}
}

func TestRunnerHonorsDisabledPolicy(t *testing.T) {
	runner := NewRunner(config.Config{
		Remote: config.RemoteConfig{TasksEnabled: false, AuditPath: filepath.Join(t.TempDir(), "audit.jsonl")},
		Tasks:  []config.TaskConfig{{Name: "echo", Command: []string{"printf", "hello"}, Timeout: time.Second}},
	})
	if _, err := runner.Run(context.Background(), "echo"); err == nil {
		t.Fatal("Run() expected disabled policy error")
	}
}

func TestRunnerLimitsTaskOutput(t *testing.T) {
	runner := NewRunner(config.Config{
		Remote: config.RemoteConfig{TasksEnabled: true, AuditPath: filepath.Join(t.TempDir(), "audit.jsonl")},
		Tasks:  []config.TaskConfig{{Name: "printf", Command: []string{"printf", "abcdef"}, Timeout: time.Second, MaxOutputBytes: 3}},
	})
	result, err := runner.Run(context.Background(), "printf")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Stdout != "abc" || !result.OutputTruncated {
		t.Fatalf("result = %#v", result)
	}
}

func TestRunnerUsesWorkingDirAndSandboxEnv(t *testing.T) {
	dir := t.TempDir()
	runner := NewRunner(config.Config{
		Remote: config.RemoteConfig{TasksEnabled: true, AuditPath: filepath.Join(t.TempDir(), "audit.jsonl")},
		Tasks: []config.TaskConfig{{
			Name:       "pwd",
			Command:    []string{"pwd"},
			Timeout:    time.Second,
			WorkingDir: dir,
		}},
	})
	result, err := runner.Run(context.Background(), "pwd")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if strings.TrimSpace(result.Stdout) != dir {
		t.Fatalf("stdout = %q", result.Stdout)
	}
}

func TestRunnerUsesSandboxEnv(t *testing.T) {
	runner := NewRunner(config.Config{
		Remote: config.RemoteConfig{TasksEnabled: true, AuditPath: filepath.Join(t.TempDir(), "audit.jsonl")},
		Tasks: []config.TaskConfig{{
			Name:    "env",
			Command: []string{"env"},
			Timeout: time.Second,
			Env:     map[string]string{"CUSTOM": "ok"},
		}},
	})
	result, err := runner.Run(context.Background(), "env")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !strings.Contains(result.Stdout, "CUSTOM=ok") || strings.Contains(result.Stdout, "LD_PRELOAD=") {
		t.Fatalf("stdout = %q", result.Stdout)
	}
}

func TestRunnerRejectsShellAtRuntime(t *testing.T) {
	runner := NewRunner(config.Config{
		Remote: config.RemoteConfig{TasksEnabled: true, AuditPath: filepath.Join(t.TempDir(), "audit.jsonl")},
		Tasks:  []config.TaskConfig{{Name: "shell", Command: []string{"sh", "-c", "echo nope"}, Timeout: time.Second}},
	})
	if _, err := runner.Run(context.Background(), "shell"); err == nil {
		t.Fatal("Run() expected shell policy error")
	}
}
