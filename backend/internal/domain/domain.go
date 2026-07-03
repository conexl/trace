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
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Hostname     string    `json:"hostname"`
	Platform     string    `json:"platform"`
	PublicIP     string    `json:"public_ip"`
	Status       string    `json:"status"`
	LastSeen     time.Time `json:"last_seen"`
	CPUPercent   float64   `json:"cpu_percent"`
	MemoryUsed   float64   `json:"memory_used_percent"`
	ProcessCount int       `json:"process_count"`
	EventCount   int       `json:"event_count"`
}

type ServerState struct {
	Summary  ServerSummary `json:"summary"`
	Snapshot AgentSnapshot `json:"snapshot"`
	Events   []AgentEvent  `json:"events"`
}
