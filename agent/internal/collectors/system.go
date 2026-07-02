package collectors

import (
	"context"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

type SystemCollector struct{}

func NewSystemCollector() *SystemCollector {
	return &SystemCollector{}
}

func (c *SystemCollector) Collect(ctx context.Context) (HostSnapshot, SystemSnapshot, error) {
	hostInfo, err := host.InfoWithContext(ctx)
	if err != nil {
		return HostSnapshot{}, SystemSnapshot{}, err
	}
	cpuTotal, _ := cpu.PercentWithContext(ctx, 200*time.Millisecond, false)
	perCPU, _ := cpu.PercentWithContext(ctx, 200*time.Millisecond, true)
	vm, _ := mem.VirtualMemoryWithContext(ctx)
	swap, _ := mem.SwapMemoryWithContext(ctx)
	parts, _ := disk.PartitionsWithContext(ctx, false)

	disks := make([]DiskSnapshot, 0, len(parts))
	for _, part := range parts {
		usage, err := disk.UsageWithContext(ctx, part.Mountpoint)
		if err != nil {
			continue
		}
		disks = append(disks, DiskSnapshot{
			Mountpoint:  usage.Path,
			Filesystem:  part.Fstype,
			Total:       usage.Total,
			Free:        usage.Free,
			UsedPercent: usage.UsedPercent,
		})
	}

	var totalCPU float64
	if len(cpuTotal) > 0 {
		totalCPU = cpuTotal[0]
	}

	return HostSnapshot{
			Hostname: hostInfo.Hostname,
			OS:       hostInfo.OS,
			Platform: hostInfo.Platform,
			Kernel:   hostInfo.KernelVersion,
			Uptime:   time.Duration(hostInfo.Uptime) * time.Second,
		}, SystemSnapshot{
			CPUPercent: totalCPU,
			PerCPU:     perCPU,
			Memory: MemorySnapshot{
				Total:       vm.Total,
				Available:   vm.Available,
				Used:        vm.Used,
				UsedPercent: vm.UsedPercent,
				SwapTotal:   swap.Total,
				SwapUsed:    swap.Used,
				SwapPercent: swap.UsedPercent,
			},
			Disks: disks,
		}, nil
}
