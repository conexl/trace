# Homelytics Frontend

Premium dark dashboard for the Homelytics home-server monitoring stack.

## Stack

- Vite + React 19 + TypeScript
- Tailwind CSS 3
- Radix UI primitives (Dialog, Tooltip, Slider)
- Recharts
- Framer Motion
- React Router

## Development

Copy the example environment and adjust values:

```bash
cp .env.example .env
```

The default Vite dev server proxies `/v1` and `/healthz` to `http://localhost:8080`.

```bash
pnpm install
pnpm dev
```

## Build

```bash
pnpm build
```

Static output is written to `dist/`.

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `VITE_API_BASE_URL` | Backend base URL (used in production builds) | `''` |
| `VITE_ADMIN_TOKEN` | Bearer token for admin read APIs | `''` |

## Live Backend Coverage

- Server list/detail, alerts and pairing are wired to the backend `/v1` API.
- Service start/stop/restart uses `POST /v1/servers/:id/service-actions`.
- Live service actions are enabled only when the agent reports the process with `remote_control: true`.
- Watchdog policy editing, add/remove service, DNS record management and agent settings are persisted through `GET/POST /v1/servers/:id/config` and pushed to agents on their next config poll.
- DNS recheck queues a `dns-recheck` server task for the agent.
- Add-service discovery from the agent (listing available systemd services) is not implemented yet; in live mode the user can type a process/service name manually.
