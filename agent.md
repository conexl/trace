# Homelytics Agent кратко

Go-агент для домашнего сервера. Собирает локальное состояние узла и буферизует snapshots для отправки в cloud.

## Уже умеет

- CPU, RAM, swap, disks, uptime, per-core load.
- Temperatures, SMART health best-effort, power profile/governor.
- Public IP, DNS checks, TCP port checks, listening TCP ports, Rx/Tx counters, speed checks.
- Process/service watchdog через process match и systemd/launchd status, с restart limits/cooldown.
- Incremental log tailing с offset.
- Durable offline JSONL buffer с replay/ack и quarantine битых строк.
- HTTP transport с optional mTLS.
- Allowlisted remote tasks с audit log.
- Policy-gated PTY primitive для будущего remote shell.
- Self-update: download, SHA256 verify, atomic replace.

## Запуск

```bash
cd agent
go run ./cmd/agent -config ./config.example.yaml -once
```

Для backend demo:

```yaml
cloud:
  transport: http
  endpoint: http://localhost:8080
```

## Ограничения

- Generated gRPC client еще не подключен.
- PTY shell есть как primitive, но запрещен config validation до mTLS/cloud authorization.
- Долговременная restart timeline пока не хранится между рестартами агента.

## Backend demo

Запусти `backend` с `HOMELYTICS_INGEST_TOKENS`, затем поставь тот же token в `agent/config.example.yaml` в `cloud.token` и `cloud.endpoint: http://localhost:8080`.

## Pairing/mTLS статус

Backend теперь умеет `/v1/pairing/claim`: one-time token -> client cert/key/CA PEM. Agent HTTP transport уже умеет читать `cloud.mtls.ca_file`, `cert_file`, `key_file`. Agent теперь имеет CLI `-pair` для вызова pairing endpoint и сохранения PEM-файлов в локальный config directory.
