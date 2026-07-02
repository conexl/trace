package collectors

import (
	"context"
	"os/exec"
	"strings"
	"time"

	"agent/internal/config"

	"github.com/shirou/gopsutil/v3/process"
)

type ProcessCollector struct {
	services ServiceManager
}

func NewProcessCollector(services ServiceManager) *ProcessCollector {
	return &ProcessCollector{services: services}
}

func (c *ProcessCollector) Collect(ctx context.Context, configs []config.ProcessConfig) ([]ProcessSnapshot, []Event) {
	processes, _ := process.ProcessesWithContext(ctx)
	results := make([]ProcessSnapshot, 0, len(configs))
	events := make([]Event, 0)

	for _, cfg := range configs {
		result := ProcessSnapshot{Name: cfg.Name, Match: cfg.Match, Service: cfg.Service}
		if cfg.Service != "" {
			status, running, err := c.services.Status(ctx, cfg.Service)
			result.Status = status
			result.Running = running
			if err != nil && !running {
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
				Message:   "critical process is not running",
				Timestamp: time.Now(),
			})
			if cfg.Restart {
				events = append(events, c.restart(ctx, cfg))
			}
		}

		results = append(results, result)
	}

	return results, events
}

func (c *ProcessCollector) restart(ctx context.Context, cfg config.ProcessConfig) Event {
	restartCtx, cancel := context.WithTimeout(ctx, cfg.GracePeriod+20*time.Second)
	defer cancel()

	var err error
	if len(cfg.RestartCommand) > 0 {
		cmd := exec.CommandContext(restartCtx, cfg.RestartCommand[0], cfg.RestartCommand[1:]...)
		err = cmd.Run()
	} else {
		err = c.services.Restart(restartCtx, cfg.Service)
	}
	if err != nil {
		return Event{
			Type:      "process.restart_failed",
			Severity:  "critical",
			Subject:   cfg.Name,
			Message:   err.Error(),
			Timestamp: time.Now(),
		}
	}
	return Event{
		Type:      "process.restarted",
		Severity:  "warning",
		Subject:   cfg.Name,
		Message:   "restart policy executed",
		Timestamp: time.Now(),
	}
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
