package system

import (
	"os/exec"
	"strconv"
	"strings"
)

// CPUMetrics holds real-time instantaneous CPU and Disk I/O stats
type CPUMetrics struct {
	UserPct       int     // % CPU user
	SystemPct     int     // % CPU system
	IdlePct       int     // % CPU idle
	TotalUsagePct int     // % CPU total active (User + System)
	DiskMBs       float64 // Disk throughput in MB/s
	DiskTPS       int     // Disk transactions per second
}

// GetLiveCPUMetrics queries macOS iostat for real-time CPU & I/O usage
func GetLiveCPUMetrics() CPUMetrics {
	metrics := CPUMetrics{
		IdlePct: 100,
	}

	out, err := exec.Command("iostat", "-c", "1").Output()
	if err != nil {
		return metrics
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 3 {
		return metrics
	}

	// Last line has the data: KB/t  tps  MB/s  us sy id   1m   5m   15m
	dataLine := lines[len(lines)-1]
	fields := strings.Fields(dataLine)
	if len(fields) >= 6 {
		// fields[1] = tps
		if tps, err := strconv.Atoi(fields[1]); err == nil {
			metrics.DiskTPS = tps
		}
		// fields[2] = MB/s
		if mbs, err := strconv.ParseFloat(fields[2], 64); err == nil {
			metrics.DiskMBs = mbs
		}
		// fields[3] = us (user)
		if us, err := strconv.Atoi(fields[3]); err == nil {
			metrics.UserPct = us
		}
		// fields[4] = sy (system)
		if sy, err := strconv.Atoi(fields[4]); err == nil {
			metrics.SystemPct = sy
		}
		// fields[5] = id (idle)
		if id, err := strconv.Atoi(fields[5]); err == nil {
			metrics.IdlePct = id
		}

		metrics.TotalUsagePct = metrics.UserPct + metrics.SystemPct
		if metrics.TotalUsagePct > 100 {
			metrics.TotalUsagePct = 100
		}
	}

	return metrics
}
