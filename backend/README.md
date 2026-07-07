# Homelytics Backend

Production-shaped backend for the Homelytics hackathon product. It uses Fx for dependency injection and lifecycle management, MongoDB as the production document store, Redis for hot runtime presence, and in-memory fallbacks when external services are not configured for fast local demos/tests.

## Architecture

- `config`: environment-driven configuration.
- `store`: deep storage seam. Mongo adapter in production, memory adapter for tests/dev fallback.
- `presence`: Redis-backed agent presence with memory fallback.
- `ingest`: accepts agent snapshot batches and converts them into current server state.
- `alerts`: evaluates snapshot-derived incidents, persists them in MongoDB when configured, and sends best-effort notifications.
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

## Authentication and Registration

The backend supports email/password users with session tokens. The first registered user becomes `owner` and can access admin endpoints. Subsequent registrations are controlled by `HOMELYTICS_REGISTRATION_DISABLED` and the optional `HOMELYTICS_ADMIN_TOKEN`.

Register the first user (owner):

```bash
curl -X POST http://localhost:8080/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"owner@example.com","password":"password123"}'
```

Login:

```bash
curl -X POST http://localhost:8080/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"owner@example.com","password":"password123"}'
```

Create an additional admin user with the admin token:

```bash
curl -X POST http://localhost:8080/v1/auth/register \
  -H 'Authorization: Bearer dev-admin-token' \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@example.com","password":"password123"}'
```

Login and register endpoints are rate-limited by client IP (defaults: 10/min for login, 5/hour for register).

## MongoDB

Set Mongo env vars to use Mongo instead of the in-memory fallback:

```bash
HOMELYTICS_MONGO_URI='mongodb://localhost:27017' \
HOMELYTICS_MONGO_DATABASE='homelytics' \
HOMELYTICS_INGEST_TOKENS=dev-agent-token \
HOMELYTICS_ADMIN_TOKEN=dev-admin-token \
go run ./cmd/backend
```

The backend stores the current server state in the `servers` collection with a unique index on `summary.id`. Alerts are stored in the `alerts` collection and sorted by `created_at`.

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
- `HOMELYTICS_TRUST_FORWARDED_HEADERS`, default `false`. Set to `true` only when the backend runs behind a trusted reverse proxy so rate limiting can use `X-Forwarded-For`/`X-Real-Ip`.
- `HOMELYTICS_ENV`, default `development`
- `HOMELYTICS_INGEST_TOKENS`, comma-separated bearer tokens for agent ingest
- `HOMELYTICS_ADMIN_TOKEN`, optional bearer token for bootstrapping admin users and read APIs
- `HOMELYTICS_REGISTRATION_DISABLED`, default `false`. When `true`, only the first user can register; additional users require an `AdminToken`.
- `HOMELYTICS_BOOTSTRAP_ADMIN_EMAIL`, optional email that is automatically granted `owner` role on first registration.
- `HOMELYTICS_LOGIN_RATE_LIMIT`, default `10`
- `HOMELYTICS_LOGIN_RATE_WINDOW`, default `1m`
- `HOMELYTICS_REGISTER_RATE_LIMIT`, default `5`
- `HOMELYTICS_REGISTER_RATE_WINDOW`, default `1h`
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
- `AI_API_KEY`, required for AI Incident Analyst (DeepSeek by default)
- `AI_BASE_URL`, default `https://api.deepseek.com`
- `AI_MODEL`, default `deepseek-chat`

## AI Incident Analyst

The backend includes an AI-powered incident analysis feature that uses DeepSeek by default:

```bash
# Configure DeepSeek (default)
export AI_API_KEY=sk-xxxxxxxxxxxxxxxx
export AI_BASE_URL=https://api.deepseek.com  # optional, default
export AI_MODEL=deepseek-chat              # optional, default

# Or use OpenAI
export AI_API_KEY=sk-xxxxxxxxxxxxxxxx
export AI_BASE_URL=https://api.openai.com/v1
export AI_MODEL=gpt-4o-mini
```

API endpoint:

- `POST /v1/incidents/{id}/analyze` - returns structured AI analysis

Response includes:
- `summary`: 1-2 sentence incident summary
- `root_cause`: most likely cause
- `severity`: critical/warning/info
- `suggestions`: array of actionable remediation steps
- `confidence`: 0.0-1.0 confidence score

## API

- `GET /healthz`
- `POST /v1/auth/register`
- `POST /v1/auth/login`
- `POST /v1/pairing/claim`
- `POST /v1/agent/snapshots`
- `GET /v1/agent/tasks`
- `GET /v1/agent/config`
- `POST /v1/agent/tasks/{task_id}/result`
- `GET /v1/alerts`
- `GET /v1/incidents`
- `GET /v1/incidents/metrics`
- `GET /v1/incidents/{id}`
- `POST /v1/incidents/{id}/restart`
- `POST /v1/incidents/{id}/disable-watchdog`
- `POST /v1/incidents/{id}/analyze`
- `GET /v1/servers`
- `GET /v1/servers/{id}`
- `GET /v1/servers/{id}/config`
- `POST /v1/servers/{id}/config`
- `POST /v1/servers/{server_id}/tasks`
- `POST /v1/servers/{server_id}/service-actions`
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

## Production Security

In `HOMELYTICS_ENV=production`, the backend refuses to start unless either mTLS (`HOMELYTICS_TLS_REQUIRE_CLIENT_CERT=true` and `HOMELYTICS_TLS_CLIENT_CA_FILE`) or ingest tokens (`HOMELYTICS_INGEST_TOKENS`) are configured. This prevents accidental open ingest endpoints in production.

## Remote Tasks

Queue an allowlisted task for an agent/server:

```bash
curl -X POST http://localhost:8080/v1/servers/homelytics-devbox/tasks \
  -H 'Authorization: Bearer dev-admin-token' \
  -H 'Content-Type: application/json' \
  -d '{"task_name":"disk-usage"}'
```

Queue an immediate DNS recheck for selected domains:

```bash
curl -X POST http://localhost:8080/v1/servers/homelytics-devbox/tasks \
  -H 'Authorization: Bearer dev-admin-token' \
  -H 'Content-Type: application/json' \
  -d '{"task_name":"dns-recheck","domains":["example.com","example.org"]}'
```

Agents poll `GET /v1/agent/tasks?agent_id=<agent-name>` and report results to `POST /v1/agent/tasks/{task_id}/result`. The agent still executes only locally allowlisted YAML tasks.

Queue a typed service action for a service that the latest agent snapshot marked as `remote_control: true`:

```bash
curl -X POST http://localhost:8080/v1/servers/homelytics-devbox/service-actions \
  -H 'Authorization: Bearer dev-admin-token' \
  -H 'Content-Type: application/json' \
  -d '{"service":"nginx","action":"restart"}'
```

## Agent Config Polling

Agents poll `GET /v1/agent/config?agent_id=<agent-name>` and apply the returned desired configuration to their local YAML. The backend stores the desired config per server under `GET /v1/servers/{id}/config` and accepts updates through `POST /v1/servers/{id}/config`. UI changes to watchdog processes, DNS checks, service policies, and update settings are persisted there and pushed to agents on the next poll.

## Alerts

Snapshot ingest evaluates alerts for:

- critical agent events such as `process.down`
- DNS/public IP mismatches
- failed DNS checks
- unreachable configured ports

Recent alerts are persisted in MongoDB when `HOMELYTICS_MONGO_URI` is configured, otherwise they use an in-memory fallback. They are available through:

```bash
curl -H 'Authorization: Bearer dev-admin-token' http://localhost:8080/v1/alerts
```

Optional legacy Telegram alert notifications from the backend process:

```bash
HOMELYTICS_TELEGRAM_BOT_TOKEN='123:abc' \
HOMELYTICS_TELEGRAM_CHAT_ID='123456' \
go run ./cmd/backend
```

## Incident Metrics

Incident reliability metrics are calculated server-side from the incident store:

```bash
curl -H 'Authorization: Bearer dev-admin-token' \
  'http://localhost:8080/v1/incidents/metrics?window=7d'
```

The response includes total/open/resolved counts, critical/warning counts, MTTR in seconds, incident frequency per day, and the same breakdown per service. Add `server_id=<id>` to scope the calculation to one server.

## Standalone Incident Notifications

Incident notifications can run as a separate process on another server. The backend publishes `incident.*` events to Redis channel `events`; the worker subscribes to that channel and sends Telegram messages. This keeps Telegram outages away from the core ingest/API path.

```bash
HOMELYTICS_REDIS_ADDR=localhost:6379 \
HOMELYTICS_NOTIFICATIONS_TELEGRAM_BOT_TOKEN='123:abc' \
HOMELYTICS_NOTIFICATIONS_TELEGRAM_CHAT_ID='123456' \
go run ./cmd/notifications
```

The worker currently sends notifications for:

- `incident.created`
- `incident.resolved`
- `incident.action`
