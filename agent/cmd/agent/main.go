package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
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
	"agent/internal/configclient"
	"agent/internal/logger"
	"agent/internal/pairing"
	"agent/internal/services"
	"agent/internal/tasksclient"
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
	pairAgent := flag.Bool("pair", false, "claim backend pairing credentials and save mTLS PEM files")
	pairDir := flag.String("pair-dir", "", "directory for pairing PEM files")
	pairCAFile := flag.String("pair-ca-file", "", "override saved CA PEM path")
	pairCertFile := flag.String("pair-cert-file", "", "override saved agent certificate PEM path")
	pairKeyFile := flag.String("pair-key-file", "", "override saved agent private key PEM path")
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

	if *pairAgent {
		client, err := pairing.NewClient(cfg.Cloud)
		if err != nil {
			fmt.Fprintf(os.Stderr, "pairing client error: %v\n", err)
			os.Exit(1)
		}
		hostname, _ := os.Hostname()
		resp, err := client.Claim(ctx, pairing.Request{AgentName: cfg.Agent.Name, Hostname: hostname})
		if err != nil {
			fmt.Fprintf(os.Stderr, "pairing error: %v\n", err)
			os.Exit(1)
		}
		saved, err := pairing.SaveCredentials(resp, pairing.SaveOptions{Dir: *pairDir, CAFile: firstNonEmpty(*pairCAFile, cfg.Cloud.MTLS.CAFile), CertFile: firstNonEmpty(*pairCertFile, cfg.Cloud.MTLS.CertFile), KeyFile: firstNonEmpty(*pairKeyFile, cfg.Cloud.MTLS.KeyFile)})
		if err != nil {
			fmt.Fprintf(os.Stderr, "save pairing credentials error: %v\n", err)
			os.Exit(1)
		}
		writeJSON(map[string]any{
			"agent_id":   resp.AgentID,
			"expires_at": resp.ExpiresAt,
			"mtls":       saved,
		})
		return
	}

	if *selfUpdate {
		result, err := updater.New().ApplyOptions(ctx, updater.Options{
			URL:              cfg.Update.URL,
			ExpectedSHA256:   cfg.Update.SHA256,
			SignatureURL:     cfg.Update.SignatureURL,
			Ed25519PublicKey: cfg.Update.Ed25519PublicKey,
		}, *updateTarget)
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
	taskClient, err := buildTaskClient(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "task client error: %v\n", err)
		os.Exit(1)
	}

	serviceManager := collectors.NewServiceManager()
	configClient, err := buildConfigClient(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config client error: %v\n", err)
		os.Exit(1)
	}
	updateClient := buildUpdater(cfg)
	agent := services.NewAgent(
		cfg,
		*configPath,
		collectors.NewSystemCollector(),
		collectors.NewNetworkCollector(),
		collectors.NewProcessCollector(serviceManager),
		collectors.NewLogCollector(),
		collectors.NewHardwareCollector(),
		serviceManager,
		buffer,
		transportClient,
		taskClient,
		configClient,
		updateClient,
		runner,
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

func buildTaskClient(cfg config.Config) (*tasksclient.Client, error) {
	if cfg.Cloud.Transport == "none" || !cfg.Remote.TasksEnabled {
		return nil, nil
	}
	return tasksclient.New(cfg.Cloud)
}

func buildConfigClient(cfg config.Config) (*configclient.Client, error) {
	if cfg.Cloud.Transport == "none" {
		return nil, nil
	}
	return configclient.New(cfg.Cloud, cfg.Agent.Name)
}

func buildUpdater(cfg config.Config) *updater.Updater {
	tlsConfig, err := buildUpdaterTLSConfig(cfg.Cloud.MTLS)
	if err != nil {
		slog.Warn("updater tls config failed", "error", err)
		return updater.New()
	}
	if tlsConfig != nil {
		return updater.NewWithTLS(tlsConfig)
	}
	return updater.New()
}

func buildUpdaterTLSConfig(cfg config.MTLS) (*tls.Config, error) {
	if cfg.CAFile == "" && cfg.CertFile == "" && cfg.KeyFile == "" {
		return nil, nil
	}
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
	if cfg.CAFile != "" {
		caPEM, err := os.ReadFile(cfg.CAFile)
		if err != nil {
			return nil, fmt.Errorf("read mtls ca: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(caPEM) {
			return nil, fmt.Errorf("parse mtls ca: no certificates found")
		}
		tlsConfig.RootCAs = pool
	}
	if cfg.CertFile != "" || cfg.KeyFile != "" {
		if cfg.CertFile == "" || cfg.KeyFile == "" {
			return nil, fmt.Errorf("mtls cert_file and key_file must be set together")
		}
		cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("load mtls key pair: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}
	return tlsConfig, nil
}

func writeJSON(value any) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(value)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
