package services

import (
	"context"
	"log/slog"
	"time"

	"agent/internal/collectors"
	"agent/internal/config"
	"agent/internal/logger"
	"agent/internal/power"
	"agent/internal/transport"
)

type Agent struct {
	cfg       config.Config
	system    *collectors.SystemCollector
	network   *collectors.NetworkCollector
	processes *collectors.ProcessCollector
	logs      *collectors.LogCollector
	hardware  *collectors.HardwareCollector
	buffer    logger.BufferedSink
	transport transport.Client
}

func NewAgent(
	cfg config.Config,
	system *collectors.SystemCollector,
	network *collectors.NetworkCollector,
	processes *collectors.ProcessCollector,
	logs *collectors.LogCollector,
	hardware *collectors.HardwareCollector,
	buffer logger.BufferedSink,
	transport transport.Client,
) *Agent {
	return &Agent{cfg: cfg, system: system, network: network, processes: processes, logs: logs, hardware: hardware, buffer: buffer, transport: transport}
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
