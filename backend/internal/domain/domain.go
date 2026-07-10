package domain

import (
	"encoding/json"
	"time"
)

type SnapshotEnvelope struct {
	Snapshots []AgentSnapshot `json:"snapshots"`
}

type AgentSnapshot struct {
	AgentName string          `json:"agent_name"`
	Host      HostSnapshot    `json:"host"`
	System    SystemSnapshot  `json:"system"`
	Network   NetworkSnapshot `json:"network"`
	Hardware  json.RawMessage `json:"hardware,omitempty"`
	Processes []Process       `json:"processes"`
	Logs      []LogChunk      `json:"logs"`
	Events    []AgentEvent    `json:"events"`
	Collected time.Time       `json:"collected_at"`
	Raw       json.RawMessage `json:"-"`
}

type HostSnapshot struct {
	Hostname string        `json:"hostname"`
	OS       string        `json:"os"`
	Platform string        `json:"platform"`
	Kernel   string        `json:"kernel"`
	Version  string        `json:"version,omitempty"`
	Uptime   time.Duration `json:"uptime"`
}

type SystemSnapshot struct {
	CPUPercent float64   `json:"cpu_percent"`
	PerCPU     []float64 `json:"per_cpu_percent"`
	Memory     Memory    `json:"memory"`
	Disks      []Disk    `json:"disks"`
}

type Memory struct {
	Total       uint64  `json:"total"`
	Available   uint64  `json:"available"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"used_percent"`
	SwapTotal   uint64  `json:"swap_total"`
	SwapUsed    uint64  `json:"swap_used"`
	SwapPercent float64 `json:"swap_percent"`
}

type Disk struct {
	Mountpoint  string  `json:"mountpoint"`
	Filesystem  string  `json:"filesystem"`
	Total       uint64  `json:"total"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"used_percent"`
}

type NetworkSnapshot struct {
	PublicIP  string          `json:"public_ip"`
	DNS       json.RawMessage `json:"dns,omitempty"`
	Ports     json.RawMessage `json:"ports,omitempty"`
	Traffic   json.RawMessage `json:"traffic,omitempty"`
	Listening json.RawMessage `json:"listening_ports,omitempty"`
	Speed     json.RawMessage `json:"speed_tests,omitempty"`
}

type Process struct {
	Name          string  `json:"name"`
	Match         string  `json:"match"`
	Service       string  `json:"service"`
	RemoteControl bool    `json:"remote_control,omitempty"`
	Running       bool    `json:"running"`
	PID           int32   `json:"pid,omitempty"`
	Status        string  `json:"status,omitempty"`
	LastExitCode  int     `json:"last_exit_code,omitempty"`
	CPUPercent    float64 `json:"cpu_percent,omitempty"`
	MemoryRSS     uint64  `json:"memory_rss,omitempty"`
	Error         string  `json:"error,omitempty"`
}

type LogChunk struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	Data      string    `json:"data"`
	Offset    int64     `json:"offset"`
	Truncated bool      `json:"truncated"`
	Collected time.Time `json:"collected_at"`
	Error     string    `json:"error,omitempty"`
}

type AgentEvent struct {
	Type      string    `json:"type"`
	Severity  string    `json:"severity"`
	Message   string    `json:"message"`
	Subject   string    `json:"subject,omitempty"`
	Action    string    `json:"action,omitempty"`
	ExitCode  int       `json:"exit_code,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type ServerSummary struct {
	ID                    string    `json:"id"`
	Name                  string    `json:"name"`
	Hostname              string    `json:"hostname"`
	Platform              string    `json:"platform"`
	PublicIP              string    `json:"public_ip"`
	Version               string    `json:"version,omitempty"`
	Status                string    `json:"status"`
	LastSeen              time.Time `json:"last_seen"`
	CPUPercent            float64   `json:"cpu_percent"`
	MemoryUsed            float64   `json:"memory_used_percent"`
	ProcessCount          int       `json:"process_count"`
	EventCount            int       `json:"event_count"`
	AppliedConfigRevision int64     `json:"applied_config_revision"`
	DesiredConfigRevision int64     `json:"desired_config_revision"`
}

type ServerState struct {
	Summary  ServerSummary `json:"summary"`
	Snapshot AgentSnapshot `json:"snapshot"`
	Events   []AgentEvent  `json:"events"`
}

const (
	RoleMember = "member"
	// Legacy roles are still accepted for existing local databases, but new SaaS
	// accounts are created as members because the MVP has no admin panel.
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleViewer = "viewer"
)

const (
	PlanFree = "free"
	PlanPlus = "plus"
)

type PlanLimits struct {
	MaxServers     int `json:"max_servers"`
	RetentionHours int `json:"retention_hours"`
}

type PlanFeatures struct {
	RemoteTasks           bool `json:"remote_tasks"`
	ServiceActions        bool `json:"service_actions"`
	AIIncidentAnalysis    bool `json:"ai_incident_analysis"`
	TelegramNotifications bool `json:"telegram_notifications"`
	ConfigManagement      bool `json:"config_management"`
	AuditLog              bool `json:"audit_log"`
}

type Subscription struct {
	Plan     string       `json:"plan" bson:"plan"`
	Status   string       `json:"status" bson:"status"`
	Limits   PlanLimits   `json:"limits" bson:"-"`
	Features PlanFeatures `json:"features" bson:"-"`
}

type User struct {
	Email        string    `json:"email" bson:"_id"`
	PasswordHash string    `json:"-" bson:"password_hash"`
	Role         string    `json:"role" bson:"role"`
	Plan         string    `json:"plan" bson:"plan"`
	Verified     bool      `json:"verified" bson:"verified"`
	CreatedAt    time.Time `json:"created_at" bson:"created_at"`
}

func NormalizePlan(plan string) string {
	switch plan {
	case PlanPlus:
		return PlanPlus
	default:
		return PlanFree
	}
}

func EntitlementsForPlan(plan string) Subscription {
	switch NormalizePlan(plan) {
	case PlanPlus:
		return Subscription{
			Plan:   PlanPlus,
			Status: "active",
			Limits: PlanLimits{
				MaxServers:     10,
				RetentionHours: 24 * 30,
			},
			Features: PlanFeatures{
				RemoteTasks:           true,
				ServiceActions:        true,
				AIIncidentAnalysis:    true,
				TelegramNotifications: true,
				ConfigManagement:      true,
				AuditLog:              true,
			},
		}
	default:
		return Subscription{
			Plan:   PlanFree,
			Status: "active",
			Limits: PlanLimits{
				MaxServers:     1,
				RetentionHours: 24,
			},
			Features: PlanFeatures{
				RemoteTasks:           false,
				ServiceActions:        false,
				AIIncidentAnalysis:    false,
				TelegramNotifications: false,
				ConfigManagement:      false,
				AuditLog:              false,
			},
		}
	}
}

type AgentDesiredConfig struct {
	Agent       AgentConfig       `json:"agent" bson:"agent"`
	Logging     LoggingConfig     `json:"logging" bson:"logging"`
	Watchdog    WatchdogConfig    `json:"watchdog" bson:"watchdog"`
	Performance PerformanceConfig `json:"performance" bson:"performance"`
	Network     NetworkConfig     `json:"network" bson:"network"`
	Processes   []ProcessConfig   `json:"processes" bson:"processes"`
	LogStreams  []LogStream       `json:"log_streams" bson:"log_streams"`
	Remote      RemoteConfig      `json:"remote" bson:"remote"`
	Update      UpdateConfig      `json:"update" bson:"update"`
	Hardware    HardwareConfig    `json:"hardware" bson:"hardware"`
	Power       PowerConfig       `json:"power" bson:"power"`
	Buffer      BufferConfig      `json:"buffer" bson:"buffer"`
	Revision    int64             `json:"revision" bson:"revision"`
	UpdatedAt   time.Time         `json:"updated_at" bson:"updated_at"`
}

type AuditLog struct {
	ID        string    `json:"id" bson:"_id,omitempty"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
	UserEmail string    `json:"user_email" bson:"user_email"`
	Action    string    `json:"action" bson:"action"`
	Target    string    `json:"target" bson:"target"` // e.g. server ID
	Details   string    `json:"details,omitempty" bson:"details,omitempty"`
}

type Metric struct {
	ServerID  string    `json:"server_id" bson:"server_id"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
	CPU       float64   `json:"cpu" bson:"cpu"`
	Memory    float64   `json:"memory" bson:"memory"`
	NetIn     uint64    `json:"net_in,omitempty" bson:"net_in,omitempty"`
	NetOut    uint64    `json:"net_out,omitempty" bson:"net_out,omitempty"`
}

type AgentConfig struct {
	Name     string        `json:"name" bson:"name"`
	Interval time.Duration `json:"interval" bson:"interval"`
}

type LoggingConfig struct {
	Level string `json:"level" bson:"level"`
}

type WatchdogConfig struct {
	PollingSeconds int `json:"polling_seconds" bson:"polling_seconds"`
	TimeoutSeconds int `json:"timeout_seconds" bson:"timeout_seconds"`
}

type PerformanceConfig struct {
	Mode     string `json:"mode" bson:"mode"`
	FanCurve string `json:"fan_curve" bson:"fan_curve"`
}

type NetworkConfig struct {
	PublicIPURL string      `json:"public_ip_url" bson:"public_ip_url"`
	DNSChecks   []DNSCheck  `json:"dns_checks" bson:"dns_checks"`
	PortChecks  []PortCheck `json:"port_checks" bson:"port_checks"`
	SpeedTests  []SpeedTest `json:"speed_tests" bson:"speed_tests"`
}

type DNSCheck struct {
	Name     string `json:"name" bson:"name"`
	Domain   string `json:"domain" bson:"domain"`
	Group    string `json:"group,omitempty" bson:"group,omitempty"`
	Critical bool   `json:"critical,omitempty" bson:"critical,omitempty"`
}

type PortCheck struct {
	Name    string        `json:"name" bson:"name"`
	Address string        `json:"address" bson:"address"`
	Timeout time.Duration `json:"timeout" bson:"timeout"`
}

type SpeedTest struct {
	Name     string        `json:"name" bson:"name"`
	URL      string        `json:"url" bson:"url"`
	MaxBytes int64         `json:"max_bytes" bson:"max_bytes"`
	Timeout  time.Duration `json:"timeout" bson:"timeout"`
}

type ProcessConfig struct {
	Name            string        `json:"name" bson:"name"`
	Match           string        `json:"match" bson:"match"`
	Service         string        `json:"service" bson:"service"`
	Critical        bool          `json:"critical" bson:"critical"`
	Restart         bool          `json:"restart" bson:"restart"`
	RemoteControl   bool          `json:"remote_control" bson:"remote_control"`
	RestartCommand  []string      `json:"restart_command" bson:"restart_command"`
	GracePeriod     time.Duration `json:"grace_period" bson:"grace_period"`
	MaxRestarts     int           `json:"max_restarts" bson:"max_restarts"`
	RestartWindow   time.Duration `json:"restart_window" bson:"restart_window"`
	RestartCooldown time.Duration `json:"restart_cooldown" bson:"restart_cooldown"`
	CPUThreshold    int           `json:"cpu_threshold" bson:"cpu_threshold"`
	MemoryThreshold int           `json:"memory_threshold" bson:"memory_threshold"`
}

type LogStream struct {
	Name     string `json:"name" bson:"name"`
	Path     string `json:"path" bson:"path"`
	MaxBytes int64  `json:"max_bytes" bson:"max_bytes"`
}

type RemoteConfig struct {
	TasksEnabled bool          `json:"tasks_enabled" bson:"tasks_enabled"`
	ShellEnabled bool          `json:"shell_enabled" bson:"shell_enabled"`
	AuditPath    string        `json:"audit_path" bson:"audit_path"`
	PollEvery    time.Duration `json:"poll_every" bson:"poll_every"`
}

type UpdateConfig struct {
	Policy           string `json:"policy" bson:"policy"`
	URL              string `json:"url" bson:"url"`
	SHA256           string `json:"sha256" bson:"sha256"`
	SignatureURL     string `json:"signature_url" bson:"signature_url"`
	Ed25519PublicKey string `json:"ed25519_public_key" bson:"ed25519_public_key"`
}

type HardwareConfig struct {
	SmartDevices []string `json:"smart_devices" bson:"smart_devices"`
}

type PowerConfig struct {
	PreventSleep bool   `json:"prevent_sleep" bson:"prevent_sleep"`
	SleepAt      string `json:"sleep_at,omitempty" bson:"sleep_at,omitempty"`
	WakeAt       string `json:"wake_at,omitempty" bson:"wake_at,omitempty"`
}

type BufferConfig struct {
	Path           string `json:"path" bson:"path"`
	MaxEvents      int    `json:"max_events" bson:"max_events"`
	MirrorToStdout bool   `json:"mirror_to_stdout" bson:"mirror_to_stdout"`
}
