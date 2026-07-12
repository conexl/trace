#!/bin/sh
set -eu

TRACE_ENDPOINT="${TRACE_ENDPOINT:-https://trace.solen.one}"
TRACE_PAIRING_CODE="${TRACE_PAIRING_CODE:-}"
TRACE_AGENT_NAME="${TRACE_AGENT_NAME:-}"
TRACE_AGENT_URL="${TRACE_AGENT_URL:-}"
TRACE_PREFIX="${TRACE_PREFIX:-/etc/homelytics}"
TRACE_STATE_DIR="${TRACE_STATE_DIR:-/var/lib/homelytics}"
TRACE_LOG_DIR="${TRACE_LOG_DIR:-/var/log/homelytics}"
TRACE_BIN="${TRACE_BIN:-/usr/local/bin/homelytics-agent}"

need() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "trace installer: missing required command: $1" >&2
    exit 1
  }
}

if [ "$(id -u)" != "0" ]; then
  echo "trace installer: please run with sudo/root" >&2
  echo "example: curl -fsSL ${TRACE_ENDPOINT}/install.sh | sudo env TRACE_PAIRING_CODE=XXXX sh" >&2
  exit 1
fi

if [ -z "$TRACE_PAIRING_CODE" ]; then
  echo "trace installer: TRACE_PAIRING_CODE is required" >&2
  exit 1
fi

need uname
need chmod
need mkdir
need cat

if command -v curl >/dev/null 2>&1; then
  fetch="curl -fsSL"
elif command -v wget >/dev/null 2>&1; then
  fetch="wget -qO-"
else
  echo "trace installer: curl or wget is required" >&2
  exit 1
fi

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"
case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *) echo "trace installer: unsupported architecture: $arch" >&2; exit 1 ;;
esac
case "$os" in
  linux|darwin) ;;
  *) echo "trace installer: unsupported OS: $os" >&2; exit 1 ;;
esac

if [ -z "$TRACE_AGENT_URL" ]; then
  TRACE_AGENT_URL="${TRACE_ENDPOINT}/downloads/homelytics-agent-${os}-${arch}"
fi

if [ -z "$TRACE_AGENT_NAME" ]; then
  host="$(hostname 2>/dev/null || echo home-server)"
  TRACE_AGENT_NAME="trace-${host}"
fi
case "$TRACE_AGENT_NAME" in
  *[!A-Za-z0-9._:-]*)
    echo "trace installer: TRACE_AGENT_NAME may only contain letters, numbers, dot, underscore, colon and dash" >&2
    exit 1
    ;;
esac

cert_dir="${TRACE_STATE_DIR}/certs"
config_file="${TRACE_PREFIX}/agent.yaml"
audit_file="${TRACE_LOG_DIR}/audit.jsonl"
buffer_file="${TRACE_STATE_DIR}/buffer.jsonl"

mkdir -p "$TRACE_PREFIX" "$TRACE_STATE_DIR" "$TRACE_LOG_DIR" "$cert_dir"
chmod 700 "$TRACE_STATE_DIR" "$cert_dir"

echo "trace installer: downloading agent binary from ${TRACE_AGENT_URL}"
tmp_bin="$(mktemp)"
# shellcheck disable=SC2086
$fetch "$TRACE_AGENT_URL" > "$tmp_bin"
install -m 0755 "$tmp_bin" "$TRACE_BIN"
rm -f "$tmp_bin"

cat > "$config_file" <<YAML
agent:
  name: ${TRACE_AGENT_NAME}
  interval: 10s
cloud:
  transport: http
  endpoint: ${TRACE_ENDPOINT}
  token: ${TRACE_PAIRING_CODE}
  replay_batch: 50
  replay_every: 15s
  mtls:
    ca_file: ${cert_dir}/ca.pem
    cert_file: ${cert_dir}/agent.pem
    key_file: ${cert_dir}/agent-key.pem
    server_ca_file: ""
logging:
  level: INFO
watchdog:
  polling_seconds: 10
  timeout_seconds: 30
performance:
  mode: balanced
  fan_curve: auto
network:
  public_ip_url: https://api.ipify.org
  dns_checks: []
  port_checks: []
  speed_tests: []
processes: []
log_streams: []
tasks:
  - name: diagnostics
    description: Safe host diagnostics bundle
    command: ["diagnostics"]
    timeout: 30s
    max_output_bytes: 65536
remote:
  tasks_enabled: true
  shell_enabled: false
  audit_path: ${audit_file}
  poll_every: 15s
update:
  policy: check
  url: ""
  sha256: ""
  signature_url: ""
  ed25519_public_key: ""
hardware:
  smart_devices: []
power:
  prevent_sleep: false
buffer:
  path: ${buffer_file}
  max_events: 1000
  mirror_to_stdout: false
YAML
chmod 600 "$config_file"

echo "trace installer: claiming pairing credentials"
"$TRACE_BIN" -config "$config_file" -pair -pair-dir "$cert_dir" >/tmp/trace-pairing.json
chmod 600 "$cert_dir"/*.pem

# Pairing token is single-use; keep config mTLS-only afterwards.
if command -v sed >/dev/null 2>&1; then
  sed -i.bak 's/^  token: .*/  token: ""/' "$config_file" && rm -f "${config_file}.bak"
fi

if [ "$os" = "linux" ]; then
  service_file="/etc/systemd/system/homelytics-agent.service"
  cat > "$service_file" <<UNIT
[Unit]
Description=Trace Homelytics Agent
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=${TRACE_BIN} -config ${config_file}
Restart=always
RestartSec=5s
User=root
NoNewPrivileges=true
ProtectSystem=full
ProtectHome=read-only
ReadWritePaths=${TRACE_STATE_DIR} ${TRACE_LOG_DIR}

[Install]
WantedBy=multi-user.target
UNIT
  systemctl daemon-reload
  systemctl enable --now homelytics-agent.service
elif [ "$os" = "darwin" ]; then
  plist="/Library/LaunchDaemons/com.homelytics.agent.plist"
  cat > "$plist" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>com.homelytics.agent</string>
  <key>ProgramArguments</key>
  <array>
    <string>${TRACE_BIN}</string>
    <string>-config</string>
    <string>${config_file}</string>
  </array>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <true/>
  <key>StandardOutPath</key>
  <string>${TRACE_LOG_DIR}/agent.out.log</string>
  <key>StandardErrorPath</key>
  <string>${TRACE_LOG_DIR}/agent.err.log</string>
</dict>
</plist>
PLIST
  chmod 644 "$plist"
  launchctl bootout system "$plist" >/dev/null 2>&1 || true
  launchctl bootstrap system "$plist"
fi

agent_id="$(sed -n 's/.*"agent_id"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' /tmp/trace-pairing.json | head -1)"
rm -f /tmp/trace-pairing.json

echo "trace installer: installed and started"
echo "trace installer: agent=${TRACE_AGENT_NAME} id=${agent_id:-unknown}"
