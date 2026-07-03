package collectors

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"agent/internal/config"

	"github.com/shirou/gopsutil/v3/process"
)

type ProcessCollector struct {
	services ServiceManager
	clock    func() time.Time
	mu       sync.Mutex
	watchdog map[string]watchdogState
}

type watchdogState struct {
	Attempts      []time.Time
	CooldownUntil time.Time
}

func NewProcessCollector(services ServiceManager) *ProcessCollector {
	return &ProcessCollector{services: services, clock: time.Now, watchdog: make(map[string]watchdogState)}
}

func (c *ProcessCollector) Collect(ctx context.Context, configs []config.ProcessConfig) ([]ProcessSnapshot, []Event) {
	processes, _ := process.ProcessesWithContext(ctx)
	results := make([]ProcessSnapshot, 0, len(configs))
	events := make([]Event, 0)

	for _, cfg := range configs {
		result := ProcessSnapshot{Name: cfg.Name, Match: cfg.Match, Service: cfg.Service, RemoteControl: cfg.RemoteControl}
		if cfg.Service != "" {
			status, err := c.services.Status(ctx, cfg.Service)
			result.Status = status.Status
			result.Running = status.Running
			result.LastExitCode = status.ExitCode
			if err != nil && !status.Running {
				result.Error = err.Error()
			}
		}

		if cfg.Match != "" {
			if proc := findProcess(ctx, processes, cfg.Match); proc != nil {
				result.Running = true
				result.PID = proc.Pid
				result.Status = "running"
				if cpuPercent, err := proc.CPUPercentWithContext(ctx); err == nil {
					result.CPUPercent = cpuPercent
				}
				if mem, err := proc.MemoryInfoWithContext(ctx); err == nil && mem != nil {
					result.MemoryRSS = mem.RSS
				}
			}
		}

		if cfg.Critical && !result.Running {
			events = append(events, Event{
				Type:      "process.down",
				Severity:  "critical",
				Subject:   cfg.Name,
				ExitCode:  result.LastExitCode,
				Message:   "critical process is not running",
				Timestamp: c.clock(),
			})
			if cfg.Restart {
				events = append(events, c.restart(ctx, cfg, result.LastExitCode))
			}
		}

		results = append(results, result)
	}

	return results, events
}

func (c *ProcessCollector) restart(ctx context.Context, cfg config.ProcessConfig, exitCode int) Event {
	now := c.clock()
	if suppressed, until := c.recordRestartAttempt(cfg, now); suppressed {
		return Event{
			Type:      "process.restart_suppressed",
			Severity:  "critical",
			Subject:   cfg.Name,
			Action:    restartAction(cfg),
			ExitCode:  exitCode,
			Message:   fmt.Sprintf("restart suppressed until %s", until.Format(time.RFC3339)),
			Timestamp: now,
		}
	}

	restartCtx, cancel := context.WithTimeout(ctx, cfg.GracePeriod+20*time.Second)
	defer cancel()

	var err error
	if len(cfg.RestartCommand) > 0 {
		cmd := exec.CommandContext(restartCtx, cfg.RestartCommand[0], cfg.RestartCommand[1:]...)
		out, runErr := cmd.CombinedOutput()
		if runErr != nil {
			err = fmt.Errorf("restart command failed: %w: %s", runErr, strings.TrimSpace(string(out)))
		}
	} else {
		err = c.services.Restart(restartCtx, cfg.Service)
	}
	if err != nil {
		return Event{
			Type:      "process.restart_failed",
			Severity:  "critical",
			Subject:   cfg.Name,
			Action:    restartAction(cfg),
			ExitCode:  firstNonZero(exitCode, exitCodeFromError(err)),
			Message:   err.Error(),
			Timestamp: now,
		}
	}
	return Event{
		Type:      "process.restarted",
		Severity:  "warning",
		Subject:   cfg.Name,
		Action:    restartAction(cfg),
		ExitCode:  exitCode,
		Message:   "restart policy executed",
		Timestamp: now,
	}
}

func (c *ProcessCollector) recordRestartAttempt(cfg config.ProcessConfig, now time.Time) (bool, time.Time) {
	key := cfg.Name
	c.mu.Lock()
	defer c.mu.Unlock()
	state := c.watchdog[key]
	if !state.CooldownUntil.IsZero() && now.Before(state.CooldownUntil) {
		return true, state.CooldownUntil
	}
	windowStart := now.Add(-cfg.RestartWindow)
	attempts := state.Attempts[:0]
	for _, attempt := range state.Attempts {
		if attempt.After(windowStart) {
			attempts = append(attempts, attempt)
		}
	}
	if len(attempts) >= cfg.MaxRestarts {
		state.Attempts = attempts
		state.CooldownUntil = now.Add(cfg.RestartCooldown)
		c.watchdog[key] = state
		return true, state.CooldownUntil
	}
	state.Attempts = append(attempts, now)
	state.CooldownUntil = time.Time{}
	c.watchdog[key] = state
	return false, time.Time{}
}

func restartAction(cfg config.ProcessConfig) string {
	if len(cfg.RestartCommand) > 0 {
		return strings.Join(cfg.RestartCommand, " ")
	}
	if cfg.Service != "" {
		return "service restart " + cfg.Service
	}
	return "restart"
}

func exitCodeFromError(err error) int {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return 0
}

func firstNonZero(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}

func findProcess(ctx context.Context, processes []*process.Process, match string) *process.Process {
	needle := strings.ToLower(match)
	for _, proc := range processes {
		name, _ := proc.NameWithContext(ctx)
		cmdline, _ := proc.CmdlineWithContext(ctx)
		if strings.Contains(strings.ToLower(name), needle) || strings.Contains(strings.ToLower(cmdline), needle) {
			return proc
		}
	}
	return nil
}
