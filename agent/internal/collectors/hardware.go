package collectors

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"agent/internal/config"
)

var sysRoot = "/sys"

type HardwareCollector struct{}

func NewHardwareCollector() *HardwareCollector {
	return &HardwareCollector{}
}

func (c *HardwareCollector) Collect(ctx context.Context, cfg config.HardwareConfig) HardwareSnapshot {
	return HardwareSnapshot{
		Temperatures: collectTemperatures(sysRoot),
		SMART:        collectSMART(ctx, cfg.SmartDevices),
		Power:        collectPower(sysRoot),
	}
}

func collectTemperatures(root string) []TemperatureSensor {
	zones, err := filepath.Glob(filepath.Join(root, "class/thermal/thermal_zone*"))
	if err != nil {
		return nil
	}
	sensors := make([]TemperatureSensor, 0, len(zones))
	for _, zone := range zones {
		name := strings.TrimSpace(readString(filepath.Join(zone, "type")))
		raw := strings.TrimSpace(readString(filepath.Join(zone, "temp")))
		milliCelsius, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			continue
		}
		if name == "" {
			name = filepath.Base(zone)
		}
		sensors = append(sensors, TemperatureSensor{Name: name, Temperature: milliCelsius / 1000})
	}
	return sensors
}

func collectPower(root string) PowerSnapshot {
	profile := strings.TrimSpace(readString(filepath.Join(root, "firmware/acpi/platform_profile")))
	governor := strings.TrimSpace(readString(filepath.Join(root, "devices/system/cpu/cpu0/cpufreq/scaling_governor")))
	return PowerSnapshot{Profile: profile, Governor: governor}
}

func collectSMART(ctx context.Context, devices []string) []SMARTDevice {
	if _, err := exec.LookPath("smartctl"); err != nil {
		return nil
	}
	if len(devices) == 0 {
		devices = scanSMARTDevices(ctx)
	}
	results := make([]SMARTDevice, 0, len(devices))
	for _, device := range devices {
		results = append(results, readSMARTDevice(ctx, device))
	}
	return results
}

func scanSMARTDevices(ctx context.Context) []string {
	cmdCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	out, err := exec.CommandContext(cmdCtx, "smartctl", "--scan").Output()
	if err != nil {
		return nil
	}
	var devices []string
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) > 0 {
			devices = append(devices, fields[0])
		}
	}
	return devices
}

func readSMARTDevice(ctx context.Context, device string) SMARTDevice {
	cmdCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	out, err := exec.CommandContext(cmdCtx, "smartctl", "-H", device).CombinedOutput()
	result := SMARTDevice{Device: device, Summary: strings.TrimSpace(string(out))}
	if err != nil {
		result.Error = err.Error()
		return result
	}
	healthy, parseErr := parseSMARTHealth(string(out))
	if parseErr != nil {
		result.Error = parseErr.Error()
		return result
	}
	result.Healthy = healthy
	return result
}

func parseSMARTHealth(output string) (bool, error) {
	lower := strings.ToLower(output)
	if strings.Contains(lower, "passed") || strings.Contains(lower, "ok") {
		return true, nil
	}
	if strings.Contains(lower, "failed") || strings.Contains(lower, "failing") {
		return false, nil
	}
	return false, errors.New("smart health status not found")
}

func readString(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

func formatSensor(sensor TemperatureSensor) string {
	return fmt.Sprintf("%s %.1fC", sensor.Name, sensor.Temperature)
}
