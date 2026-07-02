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
