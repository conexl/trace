package collectors

import "time"

type Snapshot struct {
	AgentName string            `json:"agent_name"`
	Host      HostSnapshot      `json:"host"`
	System    SystemSnapshot    `json:"system"`
	Network   NetworkSnapshot   `json:"network"`
	Processes []ProcessSnapshot `json:"processes"`
	Logs      []LogChunk        `json:"logs"`
	Events    []Event           `json:"events"`
	Collected time.Time         `json:"collected_at"`
}

type HostSnapshot struct {
	Hostname string        `json:"hostname"`
	OS       string        `json:"os"`
	Platform string        `json:"platform"`
	Kernel   string        `json:"kernel"`
	Uptime   time.Duration `json:"uptime"`
}

type SystemSnapshot struct {
	CPUPercent float64        `json:"cpu_percent"`
	PerCPU     []float64      `json:"per_cpu_percent"`
	Memory     MemorySnapshot `json:"memory"`
	Disks      []DiskSnapshot `json:"disks"`
}

type MemorySnapshot struct {
	Total       uint64  `json:"total"`
	Available   uint64  `json:"available"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"used_percent"`
	SwapTotal   uint64  `json:"swap_total"`
	SwapUsed    uint64  `json:"swap_used"`
	SwapPercent float64 `json:"swap_percent"`
}

type DiskSnapshot struct {
	Mountpoint  string  `json:"mountpoint"`
	Filesystem  string  `json:"filesystem"`
	Total       uint64  `json:"total"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"used_percent"`
}

type NetworkSnapshot struct {
	PublicIP string           `json:"public_ip"`
	DNS      []DNSResult      `json:"dns"`
	Ports    []PortResult     `json:"ports"`
	Traffic  []TrafficCounter `json:"traffic"`
}

type DNSResult struct {
	Name    string   `json:"name"`
	Domain  string   `json:"domain"`
	Records []string `json:"records"`
	Matches bool     `json:"matches_public_ip"`
	Error   string   `json:"error,omitempty"`
}

type PortResult struct {
	Name      string        `json:"name"`
	Address   string        `json:"address"`
	Reachable bool          `json:"reachable"`
	Latency   time.Duration `json:"latency"`
	Error     string        `json:"error,omitempty"`
}

type TrafficCounter struct {
	Interface string `json:"interface"`
	BytesSent uint64 `json:"bytes_sent"`
	BytesRecv uint64 `json:"bytes_recv"`
}

type ProcessSnapshot struct {
	Name       string  `json:"name"`
	Match      string  `json:"match"`
	Service    string  `json:"service"`
	Running    bool    `json:"running"`
	PID        int32   `json:"pid,omitempty"`
	Status     string  `json:"status,omitempty"`
	CPUPercent float64 `json:"cpu_percent,omitempty"`
	MemoryRSS  uint64  `json:"memory_rss,omitempty"`
	Error      string  `json:"error,omitempty"`
}

type LogChunk struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	Data      string    `json:"data"`
	Truncated bool      `json:"truncated"`
	Collected time.Time `json:"collected_at"`
	Error     string    `json:"error,omitempty"`
}

type Event struct {
	Type      string    `json:"type"`
	Severity  string    `json:"severity"`
	Message   string    `json:"message"`
	Subject   string    `json:"subject,omitempty"`
	ExitCode  int       `json:"exit_code,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}
