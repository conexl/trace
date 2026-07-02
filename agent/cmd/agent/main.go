package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"agent/internal/collectors"
	"agent/internal/commands"
	"agent/internal/config"
	"agent/internal/logger"
	"agent/internal/services"
	"agent/internal/transport"
	"agent/internal/updater"
)

func main() {
	configPath := flag.String("config", "", "path to YAML config")
	once := flag.Bool("once", false, "collect one snapshot and exit")
	listTasks := flag.Bool("list-tasks", false, "list configured remote tasks and exit")
	runTask := flag.String("run-task", "", "run one configured task and exit")
	selfUpdate := flag.Bool("self-update", false, "download and atomically replace the agent binary")
	updateTarget := flag.String("update-target", "", "override self-update target path")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}
	if *once {
		cfg.Agent.Once = true
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if *selfUpdate {
		result, err := updater.New().Apply(ctx, cfg.Update.URL, cfg.Update.SHA256, *updateTarget)
		writeJSON(result)
		if err != nil {
			fmt.Fprintf(os.Stderr, "update error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	runner := commands.NewRunner(cfg)
	if *listTasks {
		writeJSON(runner.List())
		return
	}
	if *runTask != "" {
		result, err := runner.Run(ctx, *runTask)
		writeJSON(result)
		if err != nil {
			fmt.Fprintf(os.Stderr, "task error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	buffer, err := logger.NewJSONLBuffer(cfg.Buffer)
	if err != nil {
		fmt.Fprintf(os.Stderr, "buffer error: %v\n", err)
		os.Exit(1)
	}
	defer buffer.Close()

	transportClient, err := buildTransport(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "transport error: %v\n", err)
		os.Exit(1)
	}

	serviceManager := collectors.NewServiceManager()
	agent := services.NewAgent(
		cfg,
		collectors.NewSystemCollector(),
		collectors.NewNetworkCollector(),
		collectors.NewProcessCollector(serviceManager),
		collectors.NewLogCollector(),
		collectors.NewHardwareCollector(),
		buffer,
		transportClient,
	)

	slog.Info("homelytics agent started", "name", cfg.Agent.Name, "interval", cfg.Agent.Interval)
	if err := agent.Run(ctx); err != nil && ctx.Err() == nil {
		fmt.Fprintf(os.Stderr, "agent error: %v\n", err)
		os.Exit(1)
	}
}

func buildTransport(cfg config.Config) (transport.Client, error) {
	switch cfg.Cloud.Transport {
	case "", "none":
		return transport.NopClient{}, nil
	case "http":
		return transport.NewHTTPClient(cfg.Cloud)
	default:
		return nil, fmt.Errorf("unsupported cloud transport %q", cfg.Cloud.Transport)
	}
}

func writeJSON(value any) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(value)
}
