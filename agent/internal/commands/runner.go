package commands

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"

	"agent/internal/config"
)

type Runner struct {
	tasks map[string]config.TaskConfig
}

type Result struct {
	Name      string        `json:"name"`
	ExitCode  int           `json:"exit_code"`
	Stdout    string        `json:"stdout"`
	Stderr    string        `json:"stderr"`
	Duration  time.Duration `json:"duration"`
	StartedAt time.Time     `json:"started_at"`
}

func NewRunner(tasks []config.TaskConfig) *Runner {
	byName := make(map[string]config.TaskConfig, len(tasks))
	for _, task := range tasks {
		byName[task.Name] = task
	}
	return &Runner{tasks: byName}
}

func (r *Runner) Run(ctx context.Context, name string) (Result, error) {
	task, ok := r.tasks[name]
	if !ok {
		return Result{}, fmt.Errorf("task %q is not configured", name)
	}
	if len(task.Command) == 0 {
		return Result{}, errors.New("task command is empty")
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
	if err != nil {
		return result, err
	}
	return result, nil
}
