package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"agent/internal/collectors"
	"agent/internal/commands"
	"agent/internal/config"
	"agent/internal/logger"
	"agent/internal/power"
	"agent/internal/tasksclient"
	"agent/internal/transport"
)

type Agent struct {
	cfg            config.Config
	system         *collectors.SystemCollector
	network        *collectors.NetworkCollector
	processes      *collectors.ProcessCollector
	logs           *collectors.LogCollector
	hardware       *collectors.HardwareCollector
	serviceManager collectors.ServiceManager
	buffer         logger.BufferedSink
	transport      transport.Client
	tasks          *tasksclient.Client
	runner         *commands.Runner
}

func NewAgent(
	cfg config.Config,
	system *collectors.SystemCollector,
	network *collectors.NetworkCollector,
	processes *collectors.ProcessCollector,
	logs *collectors.LogCollector,
	hardware *collectors.HardwareCollector,
	serviceManager collectors.ServiceManager,
	buffer logger.BufferedSink,
	transport transport.Client,
	tasks *tasksclient.Client,
	runner *commands.Runner,
) *Agent {
	return &Agent{cfg: cfg, system: system, network: network, processes: processes, logs: logs, hardware: hardware, serviceManager: serviceManager, buffer: buffer, transport: transport, tasks: tasks, runner: runner}
}

func (a *Agent) Run(ctx context.Context) error {
	inhibitor, err := power.Start(ctx, a.cfg.Power.PreventSleep)
	if err != nil {
		slog.Warn("power inhibitor unavailable", "error", err)
	} else {
		defer inhibitor.Stop()
	}

	if err := a.collectAndPublish(ctx); err != nil {
		return err
	}
	if err := a.flushBuffered(ctx); err != nil {
		slog.Warn("buffer replay failed", "error", err)
	}
	if a.cfg.Agent.Once {
		return nil
	}

	collectTicker := time.NewTicker(a.cfg.Agent.Interval)
	defer collectTicker.Stop()
	replayTicker := time.NewTicker(a.cfg.Cloud.ReplayEvery)
	defer replayTicker.Stop()
	taskTicker := time.NewTicker(a.cfg.Remote.PollEvery)
	defer taskTicker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-collectTicker.C:
			if err := a.collectAndPublish(ctx); err != nil {
				slog.Warn("collection cycle failed", "error", err)
			}
		case <-replayTicker.C:
			if err := a.flushBuffered(ctx); err != nil {
				slog.Warn("buffer replay failed", "error", err)
			}
		case <-taskTicker.C:
			if err := a.pollAndRunTasks(ctx); err != nil {
				slog.Warn("task polling failed", "error", err)
			}
		}
	}
}

func (a *Agent) collectAndPublish(ctx context.Context) error {
	host, system, err := a.system.Collect(ctx)
	events := make([]collectors.Event, 0)
	if err != nil {
		events = append(events, collectors.Event{
			Type:      "system.collect_failed",
			Severity:  "warning",
			Message:   err.Error(),
			Timestamp: time.Now(),
		})
	}

	processes, processEvents := a.processes.Collect(ctx, a.cfg.Processes)
	events = append(events, processEvents...)

	snapshot := collectors.Snapshot{
		AgentName: a.cfg.Agent.Name,
		Host:      host,
		System:    system,
		Network:   a.network.Collect(ctx, a.cfg.Network),
		Hardware:  a.hardware.Collect(ctx, a.cfg.Hardware),
		Processes: processes,
		Logs:      a.logs.Collect(ctx, a.cfg.LogStreams),
		Events:    events,
		Collected: time.Now(),
	}
	return a.buffer.PublishSnapshot(ctx, snapshot)
}

func (a *Agent) flushBuffered(ctx context.Context) error {
	batch, err := a.buffer.ReadBatch(a.cfg.Cloud.ReplayBatch)
	if err != nil || len(batch) == 0 {
		return err
	}
	if err := a.transport.SendSnapshots(ctx, batch); err != nil {
		return err
	}
	return a.buffer.Ack(len(batch))
}

func (a *Agent) pollAndRunTasks(ctx context.Context) error {
	if a.tasks == nil || a.runner == nil || !a.cfg.Remote.TasksEnabled || a.cfg.Cloud.Transport == "none" {
		return nil
	}
	tasks, err := a.tasks.Poll(ctx, a.cfg.Agent.Name, 1)
	if err != nil {
		return err
	}
	for _, task := range tasks {
		result, runErr := a.runTask(ctx, task)
		if err := a.tasks.Complete(ctx, task.ID, resultWithError(result, runErr)); err != nil {
			return err
		}
	}
	return nil
}

func (a *Agent) runTask(ctx context.Context, task tasksclient.Task) (tasksclient.TaskResult, error) {
	if task.Name == "service-action" {
		return a.runServiceAction(ctx, task.Payload)
	}
	result, runErr := a.runner.Run(ctx, task.Name)
	return tasksclient.FromCommandResult(result, runErr), runErr
}

func (a *Agent) runServiceAction(ctx context.Context, payload tasksclient.TaskPayload) (tasksclient.TaskResult, error) {
	started := time.Now()
	err := a.validateServiceAction(payload)
	if err == nil {
		switch payload.Action {
		case "start":
			err = a.serviceManager.Start(ctx, payload.Service)
		case "stop":
			err = a.serviceManager.Stop(ctx, payload.Service)
		case "restart":
			err = a.serviceManager.Restart(ctx, payload.Service)
		default:
			err = fmt.Errorf("unsupported service action %q", payload.Action)
		}
	}
	result := tasksclient.TaskResult{
		ExitCode:   0,
		Stdout:     fmt.Sprintf("%s %s", payload.Action, payload.Service),
		DurationMS: time.Since(started).Milliseconds(),
		StartedAt:  started,
	}
	if err != nil {
		result.ExitCode = 1
		result.Error = err.Error()
	}
	return result, err
}

func (a *Agent) validateServiceAction(payload tasksclient.TaskPayload) error {
	if payload.Service == "" {
		return fmt.Errorf("service is required")
	}
	if payload.Action != "start" && payload.Action != "stop" && payload.Action != "restart" {
		return fmt.Errorf("unsupported service action %q", payload.Action)
	}
	for _, proc := range a.cfg.Processes {
		if proc.Service == payload.Service && proc.RemoteControl {
			return nil
		}
	}
	return fmt.Errorf("service %q is not remote-controllable", payload.Service)
}

func resultWithError(result tasksclient.TaskResult, err error) tasksclient.TaskResult {
	if err != nil && result.Error == "" {
		result.Error = err.Error()
	}
	return result
}
