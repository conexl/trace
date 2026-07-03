package collectors

import "time"

type Snapshot struct {
	AgentName string            `json:"agent_name"`
	Host      HostSnapshot      `json:"host"`
	System    SystemSnapshot    `json:"system"`
	Network   NetworkSnapshot   `json:"network"`
	Hardware  HardwareSnapshot  `json:"hardware"`
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

type HardwareSnapshot struct {
	Temperatures []TemperatureSensor `json:"temperatures"`
	SMART        []SMARTDevice       `json:"smart"`
	Power        PowerSnapshot       `json:"power"`
}

type TemperatureSensor struct {
	Name        string  `json:"name"`
	Temperature float64 `json:"temperature_celsius"`
}

type SMARTDevice struct {
	Device  string `json:"device"`
	Healthy bool   `json:"healthy"`
	Summary string `json:"summary"`
	Error   string `json:"error,omitempty"`
}

type PowerSnapshot struct {
	Profile        string `json:"profile,omitempty"`
	Governor       string `json:"governor,omitempty"`
	Architecture   string `json:"architecture,omitempty"`
	Chip           string `json:"chip,omitempty"`
	ThermalLevel   string `json:"thermal_level,omitempty"`
	CPUSpeedLimit  string `json:"cpu_speed_limit,omitempty"`
	SchedulerLimit string `json:"scheduler_limit,omitempty"`
}

type NetworkSnapshot struct {
	PublicIP  string           `json:"public_ip"`
	DNS       []DNSResult      `json:"dns"`
	Ports     []PortResult     `json:"ports"`
	Traffic   []TrafficCounter `json:"traffic"`
	Listening []ListeningPort  `json:"listening_ports"`
	Speed     []SpeedResult    `json:"speed_tests"`
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

type ListeningPort struct {
	Protocol string `json:"protocol"`
	Address  string `json:"address"`
	Port     uint16 `json:"port"`
	PID      int    `json:"pid,omitempty"`
	Process  string `json:"process,omitempty"`
}

type SpeedResult struct {
	Name      string        `json:"name"`
	URL       string        `json:"url"`
	BytesRead int64         `json:"bytes_read"`
	Duration  time.Duration `json:"duration"`
	Mbps      float64       `json:"mbps"`
	Error     string        `json:"error,omitempty"`
}

type ProcessSnapshot struct {
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

type Event struct {
	Type      string    `json:"type"`
	Severity  string    `json:"severity"`
	Message   string    `json:"message"`
	Subject   string    `json:"subject,omitempty"`
	Action    string    `json:"action,omitempty"`
	ExitCode  int       `json:"exit_code,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}
