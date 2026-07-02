package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"agent/internal/config"
)

type Runner struct {
	tasks        map[string]config.TaskConfig
	tasksEnabled bool
	auditPath    string
	mu           sync.Mutex
}

type Result struct {
	Name      string        `json:"name"`
	ExitCode  int           `json:"exit_code"`
	Stdout    string        `json:"stdout"`
	Stderr    string        `json:"stderr"`
	Duration  time.Duration `json:"duration"`
	StartedAt time.Time     `json:"started_at"`
}

type TaskInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Command     []string `json:"command"`
}

type AuditEvent struct {
	Task      string    `json:"task"`
	Allowed   bool      `json:"allowed"`
	ExitCode  int       `json:"exit_code,omitempty"`
	Error     string    `json:"error,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

func NewRunner(cfg config.Config) *Runner {
	byName := make(map[string]config.TaskConfig, len(cfg.Tasks))
	for _, task := range cfg.Tasks {
		byName[task.Name] = task
	}
	return &Runner{tasks: byName, tasksEnabled: cfg.Remote.TasksEnabled, auditPath: cfg.Remote.AuditPath}
}

func (r *Runner) List() []TaskInfo {
	infos := make([]TaskInfo, 0, len(r.tasks))
	for _, task := range r.tasks {
		infos = append(infos, TaskInfo{Name: task.Name, Description: task.Description, Command: append([]string(nil), task.Command...)})
	}
	return infos
}

func (r *Runner) Run(ctx context.Context, name string) (Result, error) {
	if !r.tasksEnabled {
		err := errors.New("remote tasks are disabled by policy")
		r.audit(AuditEvent{Task: name, Allowed: false, Error: err.Error(), Timestamp: time.Now()})
		return Result{}, err
	}
	task, ok := r.tasks[name]
	if !ok {
		err := fmt.Errorf("task %q is not configured", name)
		r.audit(AuditEvent{Task: name, Allowed: false, Error: err.Error(), Timestamp: time.Now()})
		return Result{}, err
	}
	if len(task.Command) == 0 {
		err := errors.New("task command is empty")
		r.audit(AuditEvent{Task: name, Allowed: false, Error: err.Error(), Timestamp: time.Now()})
		return Result{}, err
	}

	started := time.Now()
	runCtx, cancel := context.WithTimeout(ctx, task.Timeout)
	defer cancel()

	cmd := exec.CommandContext(runCtx, task.Command[0], task.Command[1:]...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	result := Result{
		Name:      name,
		ExitCode:  0,
		Stdout:    stdout.String(),
		Stderr:    stderr.String(),
		Duration:  time.Since(started),
		StartedAt: started,
	}
	if cmd.ProcessState != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
	}
	audit := AuditEvent{Task: name, Allowed: true, ExitCode: result.ExitCode, Timestamp: time.Now()}
	if err != nil {
		audit.Error = err.Error()
		r.audit(audit)
		return result, err
	}
	r.audit(audit)
	return result, nil
}

func (r *Runner) audit(event AuditEvent) {
	if r.auditPath == "" {
		return
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	file, err := os.OpenFile(r.auditPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer file.Close()
	_, _ = file.Write(append(payload, '\n'))
}
