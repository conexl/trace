# Homelytics Backend

Production-shaped backend for the Homelytics hackathon product. It uses Fx for dependency injection and lifecycle management, MongoDB as the production document store, Redis for hot runtime presence, and in-memory fallbacks when external services are not configured for fast local demos/tests.

## Architecture

- `config`: environment-driven configuration.
- `store`: deep storage seam. Mongo adapter in production, memory adapter for tests/dev fallback.
- `presence`: Redis-backed agent presence with memory fallback.
- `ingest`: accepts agent snapshot batches and converts them into current server state.
- `alerts`: evaluates snapshot-derived incidents and sends best-effort notifications.
- `httpapi`: HTTP routes, auth middleware, request limits, security headers, and graceful server lifecycle.
- `app`: Fx composition root.

## Run Locally

```bash
cd backend
HOMELYTICS_INGEST_TOKENS=dev-agent-token \
HOMELYTICS_ADMIN_TOKEN=dev-admin-token \
go run ./cmd/backend
```

Then point the agent at it:

```yaml
cloud:
  transport: http
  endpoint: http://localhost:8080
  token: dev-agent-token
```

Run one agent snapshot:

```bash
cd ../agent
go run ./cmd/agent -config ./config.example.yaml -once
```

Read backend state:

```bash
curl -H 'Authorization: Bearer dev-admin-token' http://localhost:8080/v1/servers
curl -H 'Authorization: Bearer dev-admin-token' http://localhost:8080/v1/servers/homelytics-devbox
```

## MongoDB

Set Mongo env vars to use Mongo instead of the in-memory fallback:

```bash
HOMELYTICS_MONGO_URI='mongodb://localhost:27017' \
HOMELYTICS_MONGO_DATABASE='homelytics' \
HOMELYTICS_INGEST_TOKENS=dev-agent-token \
HOMELYTICS_ADMIN_TOKEN=dev-admin-token \
go run ./cmd/backend
```

The backend stores the current server state in the `servers` collection with a unique index on `summary.id`.

## Redis Presence

Set Redis env vars to track hot agent presence outside Mongo. Snapshot ingest and agent task polling both refresh the presence TTL, so frontend online/offline badges do not have to wait for the next full metrics snapshot.

```bash
HOMELYTICS_REDIS_ADDR='localhost:6379' \
HOMELYTICS_REDIS_KEY_PREFIX='homelytics' \
go run ./cmd/backend
```

If Redis is not configured, the backend uses an in-memory fallback and still derives status from `last_seen`.

## Environment

- `HOMELYTICS_HTTP_ADDR`, default `:8080`
- `HOMELYTICS_CORS_ALLOWED_ORIGINS`, comma-separated frontend origins, for example `http://localhost:5173`
- `HOMELYTICS_INGEST_TOKENS`, comma-separated bearer tokens for agent ingest
- `HOMELYTICS_ADMIN_TOKEN`, optional bearer token for read APIs
- `HOMELYTICS_OFFLINE_AFTER`, default `3m`
- `HOMELYTICS_MAX_EVENTS`, default `200`
- `HOMELYTICS_MONGO_URI`, empty means in-memory fallback
- `HOMELYTICS_MONGO_DATABASE`, default `homelytics`
- `HOMELYTICS_REDIS_ADDR`, empty means in-memory presence fallback
- `HOMELYTICS_REDIS_PASSWORD`, optional
- `HOMELYTICS_REDIS_DB`, default `0`
- `HOMELYTICS_REDIS_KEY_PREFIX`, default `homelytics`
- `HOMELYTICS_ALERT_MEMORY_LIMIT`, default `200`
- `HOMELYTICS_TELEGRAM_BOT_TOKEN`, optional
- `HOMELYTICS_TELEGRAM_CHAT_ID`, optional
- `HOMELYTICS_ENV`, default `development`

## API

- `GET /healthz`
- `POST /v1/pairing/claim`
- `POST /v1/agent/snapshots`
- `GET /v1/agent/tasks`
- `POST /v1/agent/tasks/{task_id}/result`
- `GET /v1/alerts`
- `GET /v1/servers`
- `GET /v1/servers/{id}`
- `POST /v1/servers/{server_id}/tasks`
- `GET /v1/tasks/{task_id}`

## Pairing and mTLS

Pairing issues an agent client certificate from the backend pairing CA. In development, if no CA files are configured, the backend creates an ephemeral in-memory CA for quick demos. In production, configure persistent CA files.

```bash
HOMELYTICS_PAIRING_TOKENS=pair-once \
HOMELYTICS_PAIRING_CA_CERT_FILE=/etc/homelytics/ca.pem \
HOMELYTICS_PAIRING_CA_KEY_FILE=/etc/homelytics/ca-key.pem \
HOMELYTICS_PAIRING_CERT_TTL=720h \
go run ./cmd/backend
```

Claim credentials once:

```bash
curl -X POST http://localhost:8080/v1/pairing/claim \
  -H 'Authorization: Bearer pair-once' \
  -H 'Content-Type: application/json' \
  -d '{"agent_name":"home-mini","hostname":"mac-mini"}'
```

Enable backend TLS and require agent client certificates for ingest. Pairing can still run before the client certificate exists; mTLS is enforced at the ingest route:

```bash
HOMELYTICS_TLS_ENABLED=true \
HOMELYTICS_TLS_CERT_FILE=/etc/homelytics/server.pem \
HOMELYTICS_TLS_KEY_FILE=/etc/homelytics/server-key.pem \
HOMELYTICS_TLS_CLIENT_CA_FILE=/etc/homelytics/ca.pem \
HOMELYTICS_TLS_REQUIRE_CLIENT_CERT=true \
go run ./cmd/backend
```

When `HOMELYTICS_TLS_REQUIRE_CLIENT_CERT=true`, `POST /v1/agent/snapshots` requires a verified client certificate. Bearer ingest tokens remain available when that flag is false for local development and migration.

## Remote Tasks

Queue an allowlisted task for an agent/server:

```bash
curl -X POST http://localhost:8080/v1/servers/homelytics-devbox/tasks \
  -H 'Authorization: Bearer dev-admin-token' \
  -H 'Content-Type: application/json' \
  -d '{"task_name":"disk-usage"}'
```

Agents poll `GET /v1/agent/tasks?agent_id=<agent-name>` and report results to `POST /v1/agent/tasks/{task_id}/result`. The agent still executes only locally allowlisted YAML tasks.

## Alerts

Snapshot ingest evaluates alerts for:

- critical agent events such as `process.down`
- DNS/public IP mismatches
- failed DNS checks
- unreachable configured ports

Recent alerts are available through:

```bash
curl -H 'Authorization: Bearer dev-admin-token' http://localhost:8080/v1/alerts
```

Optional Telegram notifications:

```bash
HOMELYTICS_TELEGRAM_BOT_TOKEN='123:abc' \
HOMELYTICS_TELEGRAM_CHAT_ID='123456' \
go run ./cmd/backend
```
