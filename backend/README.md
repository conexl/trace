# Homelytics Backend

Production-shaped backend for the Homelytics hackathon product. It uses Fx for dependency injection and lifecycle management, MongoDB as the production document store, and an in-memory store when Mongo is not configured for fast local demos/tests.

## Architecture

- `config`: environment-driven configuration.
- `store`: deep storage seam. Mongo adapter in production, memory adapter for tests/dev fallback.
- `ingest`: accepts agent snapshot batches and converts them into current server state.
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

## Environment

- `HOMELYTICS_HTTP_ADDR`, default `:8080`
- `HOMELYTICS_INGEST_TOKENS`, comma-separated bearer tokens for agent ingest
- `HOMELYTICS_ADMIN_TOKEN`, optional bearer token for read APIs
- `HOMELYTICS_OFFLINE_AFTER`, default `3m`
- `HOMELYTICS_MAX_EVENTS`, default `200`
- `HOMELYTICS_MONGO_URI`, empty means in-memory fallback
- `HOMELYTICS_MONGO_DATABASE`, default `homelytics`
- `HOMELYTICS_ENV`, default `development`

## API

- `GET /healthz`
- `POST /v1/agent/snapshots`
- `GET /v1/servers`
- `GET /v1/servers/{id}`
