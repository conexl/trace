package collectors

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"sort"
	"strings"
	"time"

	"agent/internal/config"

	gnet "github.com/shirou/gopsutil/v3/net"
)

type NetworkCollector struct {
	client *http.Client
}

func NewNetworkCollector() *NetworkCollector {
	return &NetworkCollector{client: &http.Client{Timeout: 4 * time.Second}}
}

func (c *NetworkCollector) Collect(ctx context.Context, cfg config.NetworkConfig) NetworkSnapshot {
	publicIP, publicErr := c.publicIP(ctx, cfg.PublicIPURL)
	if publicErr != nil {
		publicIP = ""
	}

	return NetworkSnapshot{
		PublicIP:  publicIP,
		DNS:       c.checkDNS(ctx, cfg.DNSChecks, publicIP),
		Ports:     c.checkPorts(ctx, cfg.PortChecks),
		Traffic:   c.traffic(ctx),
		Listening: collectListeningPorts(ctx),
		Speed:     c.speedTests(ctx, cfg.SpeedTests),
	}
}

func (c *NetworkCollector) PublicIP(ctx context.Context, endpoint string) (string, error) {
	return c.publicIP(ctx, endpoint)
}

func (c *NetworkCollector) publicIP(ctx context.Context, endpoint string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("public ip endpoint returned %s", resp.Status)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 256))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(body)), nil
}

func (c *NetworkCollector) CheckDNS(ctx context.Context, domains []string, publicIP string) []DNSResult {
	checks := make([]config.DNSCheck, 0, len(domains))
	for _, d := range domains {
		checks = append(checks, config.DNSCheck{Name: d, Domain: d})
	}
	return c.checkDNS(ctx, checks, publicIP)
}

func (c *NetworkCollector) checkDNS(ctx context.Context, checks []config.DNSCheck, publicIP string) []DNSResult {
	results := make([]DNSResult, 0, len(checks))
	for _, check := range checks {
		records, err := net.DefaultResolver.LookupHost(ctx, check.Domain)
		result := DNSResult{Name: check.Name, Domain: check.Domain, Records: records}
		if err != nil {
			result.Error = err.Error()
		} else {
			sort.Strings(result.Records)
			result.Matches = publicIP != "" && contains(result.Records, publicIP)
		}
		results = append(results, result)
	}
	return results
}

func (c *NetworkCollector) checkPorts(ctx context.Context, checks []config.PortCheck) []PortResult {
	results := make([]PortResult, 0, len(checks))
	for _, check := range checks {
		started := time.Now()
		dialer := net.Dialer{Timeout: check.Timeout}
		conn, err := dialer.DialContext(ctx, "tcp", check.Address)
		result := PortResult{Name: check.Name, Address: check.Address, Latency: time.Since(started)}
		if err != nil {
			result.Error = err.Error()
		} else {
			result.Reachable = true
			_ = conn.Close()
		}
		results = append(results, result)
	}
	return results
}

func (c *NetworkCollector) traffic(ctx context.Context) []TrafficCounter {
	counters, err := gnet.IOCountersWithContext(ctx, true)
	if err != nil {
		return nil
	}
	traffic := make([]TrafficCounter, 0, len(counters))
	for _, counter := range counters {
		traffic = append(traffic, TrafficCounter{
			Interface: counter.Name,
			BytesSent: counter.BytesSent,
			BytesRecv: counter.BytesRecv,
		})
	}
	return traffic
}

func contains(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func (c *NetworkCollector) speedTests(ctx context.Context, tests []config.SpeedTest) []SpeedResult {
	results := make([]SpeedResult, 0, len(tests))
	for _, test := range tests {
		result := SpeedResult{Name: test.Name, URL: test.URL}
		testCtx, cancel := context.WithTimeout(ctx, test.Timeout)
		started := time.Now()
		req, err := http.NewRequestWithContext(testCtx, http.MethodGet, test.URL, nil)
		if err != nil {
			result.Error = err.Error()
			cancel()
			results = append(results, result)
			continue
		}
		resp, err := c.client.Do(req)
		if err != nil {
			result.Error = err.Error()
			cancel()
			results = append(results, result)
			continue
		}
		read, err := io.Copy(io.Discard, io.LimitReader(resp.Body, test.MaxBytes))
		_ = resp.Body.Close()
		cancel()
		result.Duration = time.Since(started)
		result.BytesRead = read
		if err != nil {
			result.Error = err.Error()
		} else if resp.StatusCode >= 300 {
			result.Error = fmt.Sprintf("speed test returned %s", resp.Status)
		}
		if result.Duration > 0 {
			result.Mbps = float64(read*8) / result.Duration.Seconds() / 1_000_000
		}
		results = append(results, result)
	}
	return results
}
