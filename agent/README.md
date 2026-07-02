# Homelytics Agent

Local Go daemon for a home-server monitoring micro-SaaS. The current milestone focuses on the agent: collecting host telemetry, checking DNS/ports, watching critical processes, tailing logs, and buffering snapshots locally for later cloud upload.

## Quick Start

```bash
go run ./cmd/agent -config ./config.example.yaml -once
```

The agent writes JSONL snapshots to `homelytics-buffer.jsonl` and mirrors them to stdout by default.

List or run allowlisted tasks locally:

```bash
go run ./cmd/agent -config ./config.example.yaml -list-tasks
go run ./cmd/agent -config ./config.example.yaml -run-task disk-usage
```

## MVP Capabilities

- System telemetry: host info, uptime, CPU, per-core CPU, memory, swap, disks, temperatures, SMART health, and power profile/governor when available.
- Network audit: public IP, DNS A/AAAA records compared to public IP, TCP port checks, listening TCP ports, lightweight download speed tests, and Rx/Tx counters per interface.
- Process/service monitoring: process matching by name/cmdline plus native `systemd` on Linux and `launchd` on macOS where available.
- Watchdog events: critical missing processes create events and can execute a configured restart policy.
- Log tailing: bounded reads from configured files, so large logs do not explode memory.
- Power guard: optional sleep inhibition via `systemd-inhibit` on Linux or `caffeinate` on macOS.
- Offline buffer: snapshots are appended to JSONL, replayed in batches, and acked only after successful upload.
- Remote tasks: safe command runner for preconfigured tasks only, with JSONL audit events and a disabled-by-default shell policy.

## Remote Execution Safety

The current agent executes only named tasks from the YAML allowlist. It does not invoke a shell, so shell metacharacters are passed as plain argv values. Each accepted or rejected run is appended to `remote.audit_path`. Interactive PTY shell is intentionally rejected by config validation until mTLS identity and cloud-side authorization exist.

## Agent Transport

Set `cloud.transport: http` and `cloud.endpoint` to send buffered snapshots to `POST /v1/agent/snapshots`. The default `none` transport keeps everything local for demos. When `cloud.mtls.ca_file`, `cert_file`, and `key_file` are configured, the HTTP client uses those credentials for mutual TLS. The transport seam is intentionally small so the HTTP client can be replaced with generated gRPC without touching collectors.

## Next Milestone

- Generate Go code from `proto/agent.proto` and add a real mTLS gRPC sink.
- Add pairing flow for one-time token enrollment.
- Add PTY streaming for remote shell behind an explicit allowlist and audit log.
- Add Telegram alert worker on the backend side.
