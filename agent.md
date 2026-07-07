# Agent status

`agent` - локальный демон Trace/Homelytics для домашнего сервера. Он собирает состояние узла, следит за сервисами, буферизует телеметрию при обрыве сети, принимает безопасные задачи от backend и применяет desired config.

Последняя проверка по коду: `agent` проходит `go test ./...`.

## Команды запуска

```bash
cd agent
go run ./cmd/agent -config ./config.example.yaml -once
```

Полезные CLI-режимы:

```bash
go run ./cmd/agent -config ./config.example.yaml -list-tasks
go run ./cmd/agent -config ./config.example.yaml -run-task disk-usage
go run ./cmd/agent -config ./config.example.yaml -pair -pair-dir ./certs
go run ./cmd/agent -config ./config.example.yaml -self-update
```

## Готово

### Сбор данных

- Host snapshot: hostname, OS, platform, kernel, version, uptime.
- System metrics: CPU percent, per-core CPU, memory, swap, disks.
- Hardware snapshot: temperature sensors, SMART health, power profile data.
- macOS/M-series hints: Apple Silicon chip, thermal level, CPU speed limit, scheduler limit, battery, если ОС отдает эти данные.
- Linux power hints: power profile и CPU governor, если доступны.
- Agent health: возраст конфига, размер локального буфера, статус последней выгрузки.
- Applied config revision в каждом snapshot.

### Network

- Public IP lookup через `network.public_ip_url`.
- DNS A/AAAA checks с проверкой against public IP.
- TCP reachability checks для заданных адресов.
- Listening TCP ports с pid/process, если ОС отдает эти данные.
- Rx/Tx counters по интерфейсам.
- Lightweight download speed checks с лимитом байт и timeout.
- Backend task `dns-recheck`, которая повторно проверяет указанные домены.

### Process and service control

- Process matching по name/cmdline.
- Native service status через `systemd` на Linux и `launchd` на macOS.
- Service discovery через `systemctl list-unit-files` или `launchctl list`.
- Service actions: `start`, `stop`, `restart`.
- Remote service actions разрешены только для процессов с `remote_control: true`.
- Process snapshot содержит service name, running status, OS status, last exit code, pid, CPU, RSS и errors.

### Watchdog

- Critical process detection.
- Restart policy на уровне процесса: `restart`, `restart_command`, `service`, `max_restarts`, `restart_window`, `restart_cooldown`, `grace_period`.
- Watchdog events уходят в snapshots.
- Restart events могут создавать backend alerts.
- Exit code попадает в snapshot, если service manager его отдает.

### Logs

- Incremental log tailing для заданных файлов.
- Bounded read size через `log_streams[].max_bytes`.
- Offset tracking в каждом log chunk.
- Большие log-файлы не читаются в память целиком.

### Offline buffer

- Durable JSONL buffer по пути `buffer.path`.
- Replay batch через `cloud.replay_batch` и `cloud.replay_every`.
- Ack выполняется только после успешной выгрузки.
- Битые строки попадают в quarantine и не блокируют replay.
- Есть mirror to stdout для локального demo.

### Transport and pairing

- HTTP transport в backend `POST /v1/agent/snapshots`.
- Token auth через `cloud.token`.
- Optional mTLS через `cloud.mtls.ca_file`, `cert_file`, `key_file`.
- Pairing CLI вызывает backend pairing endpoint и сохраняет CA, cert, key PEM files.
- TLS minimum version: TLS 1.2.

### Desired config

- Agent получает desired config через `GET /v1/agent/config`, если включен HTTP transport.
- Desired config может менять agent interval, logging, watchdog, performance, network checks, processes, log streams, remote policy, update policy, hardware, power и buffer settings.
- Agent пишет merged YAML config на диск с правами `0600`.
- После изменения config agent завершает процесс, чтобы `systemd` или `launchd` перезапустили его с новым config.

### Remote tasks

- Agent poll-ит backend tasks через `GET /v1/agent/tasks`.
- Agent отправляет результат в `/v1/agent/tasks/{id}/result`.
- Выполняются только allowlisted YAML tasks.
- Shell executables запрещены: `sh`, `bash`, `zsh`, `fish`, `pwsh`, `powershell`, `cmd`, `cmd.exe`.
- Runner использует minimal env и запрещает override для `PATH`, `LD_PRELOAD`, `DYLD_INSERT_LIBRARIES`.
- `working_dir` должен быть absolute и clean.
- stdout/stderr ограничены по размеру.
- Каждая принятая или отклоненная task пишет JSONL audit event в `remote.audit_path`.

### Power

- Optional sleep prevention через `systemd-inhibit` на Linux.
- Optional sleep prevention через `caffeinate` на macOS.
- Agent пишет `prevent_sleep` в power snapshot.

### Self-update

- Manual self-update через `-self-update`.
- Periodic update checks в daemon mode.
- Policies: `manual`, `check`, `auto`.
- Download во временный файл.
- SHA256 verification, если задан `update.sha256`.
- Ed25519 signature verification, если заданы `signature_url` и `ed25519_public_key`.
- Atomic binary replace через `os.Rename`.
- Policy `check` отправляет event `update.available` без замены бинарника.
- Policy `auto` заменяет бинарник и завершает процесс, чтобы supervisor перезапустил agent.

## Готово

### AI Incident Analyst

- Backend имеет AI client для OpenAI-совместимых API: `/home/conexl/Code/Trace/backend/internal/ai/client.go`.
- **DeepSeek используется по умолчанию** (model: `deepseek-chat`).
- Endpoint `POST /v1/incidents/{id}/analyze` возвращает structured JSON analysis.
- Frontend показывает AI Analysis block в IncidentDrawer с кнопкой "AI Analyze".
- Analysis содержит: summary, root_cause, severity, suggestions, confidence.
- Fallback если AI_API_KEY не настроен: сообщение "Configure AI_API_KEY to enable".
- Поддержка OpenAI, Anthropic через AI_BASE_URL и AI_MODEL env vars.

### Incident Autopilot

- Backend создает incidents из agent events (`process.down`, `process.restart_failed`, `process.restart_suppressed`).
- Incident содержит timeline с событиями: crash, restart attempt, action execution.
- Frontend показывает incident drawer с timeline и actions.
- MVP actions: `Restart` (через `service-action` task) и `Disable Watchdog` (через desired config).
- Actions `Run Diagnostics` и `Rollback Config` показываются как "Coming soon".
- Все actions логируются в audit log.
- Incident автоматически закрывается при `process.up` event.
- Actions защищены `requireAdmin` middleware.
- Backend тесты проходят: `go test ./...`.
- Frontend билдится: `npm run build`.

### Remote shell

- PTY primitive есть в `agent/internal/remote`.
- Config validation запрещает `remote.shell_enabled=true`.
- Backend authorization, session audit и streaming protocol для shell пока не подключены.
- В MVP remote shell должен оставаться выключенным.

### gRPC

- `agent/proto/agent.proto` есть.
- Generated gRPC client/server transport не подключен к runtime.
- Рабочий production path сейчас: HTTP плюс optional mTLS.

### Service history

- Agent отправляет watchdog/service events в snapshots.
- Agent не хранит длинную restart timeline между собственными рестартами.
- Backend владеет incident history и audit timeline.

### Platform depth

- Linux и macOS service managers работают через native tools.
- macOS ARM/M-series данные собираются best-effort через доступные OS commands.
- Windows support не реализован.
- SMART зависит от наличия `smartctl` и прав процесса.

## Не реализовано

- Full remote PTY shell через backend/frontend.
- Task `diagnostics`.
- Config rollback с локальной историей.
- First-class task `disable-watchdog`.
- gRPC transport в runtime.
- eBPF traffic analysis.
- External port checks снаружи LAN.
- Wake scheduling через `power.sleep_at` и `power.wake_at`.
- Per-task container/jail sandboxing.
- User switching для tasks за пределами текущего process-level hook.
- Agent-side incident model. Backend должен создавать incidents из snapshots, alerts, audit logs и task results.

## Backend/frontend contract

### Snapshot upload

Agent отправляет buffered snapshots в:

```text
POST /v1/agent/snapshots
```

Snapshot содержит host, system, network, hardware, processes, logs, events, config revision, available services, agent health и collection time.

### Task polling

Agent poll-ит:

```text
GET /v1/agent/tasks?agent_id=<agent-name>&limit=1
```

Built-in task names:

- `service-action`
- `dns-recheck`

Также работают allowlisted task names из YAML, например `disk-usage`.

### Task result

Agent отправляет результат в:

```text
POST /v1/agent/tasks/{task_id}/result
```

Result содержит exit code, stdout, stderr, duration, start time и optional error.

### Desired config

Agent получает config через:

```text
GET /v1/agent/config?agent_id=<agent-name>
```

Ответ мапится в локальный YAML config. Если merged config изменился, agent пишет его на диск и выходит для supervisor restart.

## MVP recommendation

Для incident MVP активными должны быть:

- `Restart`: active, работает через `service-action`.
- `Disable Watchdog`: active только если backend меняет desired config. Не отправлять fake agent task.

Показывать disabled или `Coming soon`:

- `Run Diagnostics`: planned task name `diagnostics`.
- `Rollback Config`: нужен config history, diff, confirmation и audit trail.
