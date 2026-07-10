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

- System telemetry: host info, uptime, CPU, per-core CPU, memory, swap, disks, temperatures, SMART health, Linux power profile/governor, and macOS Apple Silicon thermal pressure/CPU limit hints when available.
- Network audit: public IP, DNS A/AAAA records compared to public IP, TCP port checks, listening TCP ports, lightweight download speed tests, and Rx/Tx counters per interface.
- Process/service monitoring: process matching by name/cmdline plus native `systemd` on Linux and `launchd` on macOS where available, including service exit status where the OS exposes it.
- Watchdog events: critical missing processes create events and can execute a configured restart policy with max restart windows and cooldowns.
- Log tailing: bounded reads from configured files, so large logs do not explode memory.
- Power guard: optional sleep inhibition via `systemd-inhibit` on Linux or `caffeinate` on macOS.
- Offline buffer: snapshots are appended to durable JSONL, replayed in batches, acked only after successful upload, and corrupt lines are quarantined instead of blocking future replay.
- Remote tasks: built-in safe tasks (`service-action`, `dns-recheck`, `diagnostics`) plus preconfigured allowlisted tasks only, with JSONL audit events and a disabled-by-default shell policy.

## Remote Execution Safety

The current agent executes only built-in safe tasks or named tasks from the YAML allowlist. It does not invoke a shell, rejects common shell interpreters as task executables, runs with a minimal environment, supports an optional absolute `working_dir`, and caps stdout/stderr per task. Dashboard service actions are accepted only for `processes:` entries with `remote_control: true` and a concrete `service`. The `diagnostics` task returns a bounded JSON bundle from existing collectors instead of running arbitrary commands. Each accepted or rejected run is appended to `remote.audit_path`. Agent-side PTY shell primitives are implemented, but `remote.shell_enabled` stays rejected by config validation until mTLS identity and cloud-side authorization exist. This keeps the dangerous path present for integration work without making it accidentally reachable in demos.

## Agent Transport

Set `cloud.transport: http` and `cloud.endpoint` to send buffered snapshots to `POST /v1/agent/snapshots`. The default `none` transport keeps everything local for demos. When `cloud.mtls.ca_file`, `cert_file`, and `key_file` are configured, the HTTP client uses those credentials for mutual TLS. The transport seam is intentionally small so the HTTP client can be replaced with generated gRPC without touching collectors.

## Self Update

Configure `update.url` and, for production, `update.signature_url` plus `update.ed25519_public_key`, then run:

```bash
homelytics-agent -config /etc/homelytics/agent.yaml -self-update
```

The updater downloads to a temporary file, verifies SHA256 when configured, verifies an Ed25519 signature when `ed25519_public_key` is configured, chmods it executable, and atomically replaces the target binary. Under `systemd` or `launchd`, the supervisor can restart the daemon after the replacement. Example service files live in `deploy/`.

## Next Milestone

- Add a real gRPC sink when the backend transport moves beyond HTTP.
- Add PTY streaming for remote shell behind cloud-side authorization and a stricter audit workflow.
- Persist richer watchdog restart history across agent restarts.

## Pairing

Fast SaaS install flow:

```bash
curl -fsSL https://trace.solen.one/install.sh | sudo env TRACE_PAIRING_CODE=PASTE_CODE TRACE_AGENT_NAME=home-mini sh
```

The installer downloads a published agent binary, writes `/etc/homelytics/agent.yaml`,
claims mTLS credentials, removes the one-time token from the config, and starts the
agent through `systemd` on Linux or `launchd` on macOS. Override
`TRACE_AGENT_URL` when testing unpublished builds.

Claim backend-issued mTLS credentials with a one-time pairing token:

```bash
go run ./cmd/agent -config ./config.example.yaml -pair -pair-dir ./certs
```

The command writes `ca.pem`, `agent.pem`, and `agent-key.pem` with `0600` permissions and prints the paths as JSON. Set `cloud.mtls.*` to those paths before running the daemon over HTTPS/mTLS.

## Remote Task Polling

When `cloud.transport: http` and `remote.tasks_enabled: true`, the agent polls the backend every `remote.poll_every` for tasks targeting `agent.name`. Built-in safe tasks are handled by the agent; custom tasks must be declared in the local YAML `tasks:` allowlist. Every attempt is written to the audit JSONL.
