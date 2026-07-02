package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"agent/internal/collectors"
	"agent/internal/config"
	"agent/internal/logger"
	"agent/internal/services"
)

func main() {
	configPath := flag.String("config", "", "path to YAML config")
	once := flag.Bool("once", false, "collect one snapshot and exit")
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

	sink, err := logger.NewJSONLBuffer(cfg.Buffer)
	if err != nil {
		fmt.Fprintf(os.Stderr, "buffer error: %v\n", err)
		os.Exit(1)
	}
	defer sink.Close()

	serviceManager := collectors.NewServiceManager()
	agent := services.NewAgent(
		cfg,
		collectors.NewSystemCollector(),
		collectors.NewNetworkCollector(),
		collectors.NewProcessCollector(serviceManager),
		collectors.NewLogCollector(),
		sink,
	)

	slog.Info("homelytics agent started", "name", cfg.Agent.Name, "interval", cfg.Agent.Interval)
	if err := agent.Run(ctx); err != nil && ctx.Err() == nil {
		fmt.Fprintf(os.Stderr, "agent error: %v\n", err)
		os.Exit(1)
	}
}
