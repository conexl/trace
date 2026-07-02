package services

import (
	"context"
	"log/slog"
	"time"

	"agent/internal/collectors"
	"agent/internal/config"
	"agent/internal/logger"
	"agent/internal/power"
)

type Agent struct {
	cfg       config.Config
	system    *collectors.SystemCollector
	network   *collectors.NetworkCollector
	processes *collectors.ProcessCollector
	logs      *collectors.LogCollector
	sink      logger.Sink
}

func NewAgent(
	cfg config.Config,
	system *collectors.SystemCollector,
	network *collectors.NetworkCollector,
	processes *collectors.ProcessCollector,
	logs *collectors.LogCollector,
	sink logger.Sink,
) *Agent {
	return &Agent{cfg: cfg, system: system, network: network, processes: processes, logs: logs, sink: sink}
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
	if a.cfg.Agent.Once {
		return nil
	}

	ticker := time.NewTicker(a.cfg.Agent.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := a.collectAndPublish(ctx); err != nil {
				slog.Warn("collection cycle failed", "error", err)
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
		Processes: processes,
		Logs:      a.logs.Collect(ctx, a.cfg.LogStreams),
		Events:    events,
		Collected: time.Now(),
	}
	return a.sink.PublishSnapshot(ctx, snapshot)
}
