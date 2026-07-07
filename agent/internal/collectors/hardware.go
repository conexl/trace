package collectors

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
	power := PowerSnapshot{Profile: profile, Governor: governor, Architecture: runtime.GOARCH}
	if runtime.GOOS == "darwin" {
		power = mergePower(power, collectDarwinPower(context.Background()))
	}
	power.PreventSleep = checkPreventSleep()
	return power
}

func checkPreventSleep() bool {
	if runtime.GOOS == "darwin" {
		cmd := exec.Command("pmset", "-g", "assertions")
		out, err := cmd.Output()
		if err != nil {
			return false
		}
		return strings.Contains(string(out), "PreventUserIdleSystemSleep") || strings.Contains(string(out), "PreventSystemSleep")
	}
	if runtime.GOOS == "linux" {
		cmd := exec.Command("systemd-inhibit", "--list", "--no-pager")
		out, err := cmd.Output()
		if err != nil {
			return false
		}
		return strings.Contains(string(out), "sleep")
	}
	return false
}

func collectDarwinPower(ctx context.Context) PowerSnapshot {
	power := PowerSnapshot{Architecture: runtime.GOARCH}
	if chip := runTrimmed(ctx, 2*time.Second, "sysctl", "-n", "machdep.cpu.brand_string"); chip != "" {
		power.Chip = chip
	}
	if arm64 := runTrimmed(ctx, 2*time.Second, "sysctl", "-n", "hw.optional.arm64"); arm64 == "1" && power.Chip == "" {
		power.Chip = "Apple Silicon"
	}
	if out := runTrimmed(ctx, 2*time.Second, "pmset", "-g", "therm"); out != "" {
		power = mergePower(power, parsePMSetTherm(out))
	}
	if out := runTrimmed(ctx, 2*time.Second, "pmset", "-g", "batt"); out != "" {
		power.Battery = out
	}
	return power
}

func parsePMSetTherm(output string) PowerSnapshot {
	var power PowerSnapshot
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		switch key {
		case "CPU_Speed_Limit":
			power.CPUSpeedLimit = value
		case "Scheduler_Limit":
			power.SchedulerLimit = value
		case "Thermal_Level":
			power.ThermalLevel = value
		}
	}
	return power
}

func mergePower(base PowerSnapshot, extra PowerSnapshot) PowerSnapshot {
	if base.Profile == "" {
		base.Profile = extra.Profile
	}
	if base.Governor == "" {
		base.Governor = extra.Governor
	}
	if base.Architecture == "" {
		base.Architecture = extra.Architecture
	}
	if base.Chip == "" {
		base.Chip = extra.Chip
	}
	if base.ThermalLevel == "" {
		base.ThermalLevel = extra.ThermalLevel
	}
	if base.CPUSpeedLimit == "" {
		base.CPUSpeedLimit = extra.CPUSpeedLimit
	}
	if base.SchedulerLimit == "" {
		base.SchedulerLimit = extra.SchedulerLimit
	}
	if base.Battery == "" {
		base.Battery = extra.Battery
	}
	return base
}

func runTrimmed(ctx context.Context, timeout time.Duration, name string, args ...string) string {
	cmdCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	out, err := exec.CommandContext(cmdCtx, name, args...).CombinedOutput()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
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
