package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"agent/internal/collectors"
	"agent/internal/commands"
	"agent/internal/config"
	"agent/internal/configclient"
	"agent/internal/logger"
	"agent/internal/power"
	"agent/internal/tasksclient"
	"agent/internal/transport"
	"agent/internal/updater"
)

type Agent struct {
	cfg            config.Config
	configPath     string
	system         *collectors.SystemCollector
	network        *collectors.NetworkCollector
	processes      *collectors.ProcessCollector
	logs           *collectors.LogCollector
	hardware       *collectors.HardwareCollector
	serviceManager collectors.ServiceManager
	buffer         logger.BufferedSink
	transport      transport.Client
	tasks          *tasksclient.Client
	configClient   ConfigClient
	updater        updaterClient
	runner         *commands.Runner
	startTime      time.Time
	lastConfigFetch time.Time
	lastUploadSuccess bool
	lastServiceDiscovery time.Time
	cachedServices       []string
}

type ConfigClient interface {
	Fetch(ctx context.Context) (configclient.DesiredConfig, error)
}

type updaterClient interface {
	ApplyOptions(ctx context.Context, opts updater.Options, target string) (updater.Result, error)
	CheckOptions(ctx context.Context, opts updater.Options) (updater.Result, error)
}

func NewAgent(
	cfg config.Config,
	configPath string,
	system *collectors.SystemCollector,
	network *collectors.NetworkCollector,
	processes *collectors.ProcessCollector,
	logs *collectors.LogCollector,
	hardware *collectors.HardwareCollector,
	serviceManager collectors.ServiceManager,
	buffer logger.BufferedSink,
	transport transport.Client,
	tasks *tasksclient.Client,
	configClient ConfigClient,
	updaterSvc updaterClient,
	runner *commands.Runner,
) *Agent {
	return &Agent{
		cfg: cfg, configPath: configPath, system: system, network: network, processes: processes, logs: logs, hardware: hardware, serviceManager: serviceManager, buffer: buffer, transport: transport, tasks: tasks, configClient: configClient, updater: updaterSvc, runner: runner,
		startTime: time.Now(),
		lastConfigFetch: time.Now(),
		lastUploadSuccess: true,
	}
}

var osExit = os.Exit

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
	configTicker := time.NewTicker(30 * time.Second)
	defer configTicker.Stop()
	updateTicker := time.NewTicker(5 * time.Minute)
	defer updateTicker.Stop()
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
		case <-configTicker.C:
			if err := a.pollConfig(ctx); err != nil {
				slog.Warn("config poll failed", "error", err)
			}
		case <-updateTicker.C:
			if done, err := a.checkUpdate(ctx); err != nil {
				slog.Warn("update check failed", "error", err)
			} else if done {
				return nil
			}
		}
	}
}

func (a *Agent) pollConfig(ctx context.Context) error {
	if a.configClient == nil || a.configPath == "" {
		return nil
	}
	desired, err := a.configClient.Fetch(ctx)
	if err != nil {
		return err
	}
	a.lastConfigFetch = time.Now()
	changed, err := a.applyDesiredConfig(a.configPath, desired)
	if err != nil {
		return err
	}
	if changed {
		slog.Info("agent config changed, exiting to allow restart with new config")
		os.Exit(0)
	}
	return nil
}

func (a *Agent) checkUpdate(ctx context.Context) (bool, error) {
	policy := a.cfg.Update.Policy
	if policy == "" {
		policy = "check"
	}
	if policy == "manual" || strings.TrimSpace(a.cfg.Update.URL) == "" {
		return false, nil
	}
	if a.updater == nil {
		return false, nil
	}

	currentSHA, err := updater.CurrentExecutableSHA256()
	if err != nil {
		return false, err
	}

	opts := updater.Options{
		URL:              a.cfg.Update.URL,
		ExpectedSHA256:   a.cfg.Update.SHA256,
		SignatureURL:     a.cfg.Update.SignatureURL,
		Ed25519PublicKey: a.cfg.Update.Ed25519PublicKey,
	}

	if policy == "auto" {
		result, err := a.updater.ApplyOptions(ctx, opts, "")
		if err != nil {
			return false, err
		}
		if !result.Updated {
			return false, nil
		}
		if strings.EqualFold(result.SHA256, currentSHA) {
			slog.Info("update matches current executable", "sha256", result.SHA256)
			return false, nil
		}
		slog.Info("agent updated, exiting to restart", "sha256", result.SHA256, "signature_verified", result.SignatureVerified)
		osExit(0)
		return false, nil
	}

	result, err := a.updater.CheckOptions(ctx, opts)
	if err != nil {
		return false, err
	}
	if strings.EqualFold(result.SHA256, currentSHA) {
		return false, nil
	}
	slog.Info("update available", "sha256", result.SHA256, "signature_verified", result.SignatureVerified)
	if a.buffer != nil {
		_ = a.buffer.PublishSnapshot(ctx, collectors.Snapshot{
			AgentName: a.cfg.Agent.Name,
			Events: []collectors.Event{{
				Type:      "update.available",
				Severity:  "info",
				Message:   fmt.Sprintf("Update available: sha256=%s signature_verified=%v", result.SHA256, result.SignatureVerified),
				Timestamp: time.Now(),
			}},
			Collected: time.Now(),
		})
	}
	return false, nil
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

	var availableServices []string
	if time.Since(a.lastServiceDiscovery) > 5*time.Minute {
		if svcs, err := a.serviceManager.ListServices(ctx); err == nil {
			a.cachedServices = svcs
			a.lastServiceDiscovery = time.Now()
		}
	}
	availableServices = a.cachedServices

	snapshot := collectors.Snapshot{
		AgentName: a.cfg.Agent.Name,
		Host:      host,
		System:    system,
		Network:   a.network.Collect(ctx, a.cfg.Network),
		Hardware:  a.hardware.Collect(ctx, a.cfg.Hardware),
		Processes: processes,
		Logs:      a.logs.Collect(ctx, a.cfg.LogStreams),
		Events:    events,
		AppliedConfigRevision: a.cfg.Agent.Revision,
		AvailableServices:     availableServices,
		Health: collectors.AgentHealth{
			ConfigAgeSeconds:    time.Since(a.cfg.Agent.LastLoaded).Seconds(),
			BufferedEventsCount: a.buffer.Count(),
			LastUploadSuccess:   a.lastUploadSuccess,
		},
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
		a.lastUploadSuccess = false
		return err
	}
	a.lastUploadSuccess = true
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
	if task.Name == "dns-recheck" {
		return a.runDNSRecheck(ctx, task.Payload)
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

func (a *Agent) runDNSRecheck(ctx context.Context, payload tasksclient.TaskPayload) (tasksclient.TaskResult, error) {
	started := time.Now()
	result := tasksclient.TaskResult{ExitCode: 0, StartedAt: started}
	if len(payload.Domains) == 0 {
		result.ExitCode = 1
		result.Error = "no domains provided"
		return result, fmt.Errorf("no domains provided")
	}

	publicIP := ""
	if a.network != nil && a.cfg.Network.PublicIPURL != "" {
		ip, err := a.network.PublicIP(ctx, a.cfg.Network.PublicIPURL)
		if err == nil {
			publicIP = ip
		}
	}

	var checks []collectors.DNSResult
	if a.network != nil {
		checks = a.network.CheckDNS(ctx, payload.Domains, publicIP)
	} else {
		result.ExitCode = 1
		result.Error = "network collector unavailable"
		return result, fmt.Errorf("network collector unavailable")
	}

	data, err := json.Marshal(checks)
	if err != nil {
		result.ExitCode = 1
		result.Error = err.Error()
		return result, err
	}
	result.Stdout = string(data)
	result.DurationMS = time.Since(started).Milliseconds()
	for _, c := range checks {
		if c.Error != "" {
			result.ExitCode = 1
			result.Error = fmt.Sprintf("%s failed: %s", c.Domain, c.Error)
			break
		}
	}
	return result, nil
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
