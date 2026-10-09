package system

import (
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// ThermalStatus holds thermal and CPU power state
type ThermalStatus struct {
	CPUSpeedLimit     int     // 100 = 100% full speed, <100 = throttled
	CPUSchedulerLimit int     // 100 = 100%
	AvailableCPUs     int     // Number of active cores available
	IsThrottled       bool    // True if Speed Limit < 100
	PowerSource       string  // "AC Power" or "Battery Power"
	BatteryPct        int     // 0-100
	IsCharging        bool    // True if charging
	LoadAvg1          float64 // 1-minute load average
	LoadAvg5          float64 // 5-minute load average
	LoadAvg15         float64 // 15-minute load average
}

var (
	speedLimitRegex      = regexp.MustCompile(`CPU_Speed_Limit\s*=\s*(\d+)`)
	schedulerLimitRegex  = regexp.MustCompile(`CPU_Scheduler_Limit\s*=\s*(\d+)`)
	availableCPUsRegex   = regexp.MustCompile(`CPU_Available_CPUs\s*=\s*(\d+)`)
	batteryPctRegex      = regexp.MustCompile(`(\d+)%`)
)

// GetThermalStatus returns current thermal throttling and power info
func GetThermalStatus() ThermalStatus {
	status := ThermalStatus{
		CPUSpeedLimit:     100,
		CPUSchedulerLimit: 100,
		AvailableCPUs:     8,
		PowerSource:       "Unknown",
	}

	// 1. pmset -g therm
	if out, err := exec.Command("pmset", "-g", "therm").Output(); err == nil {
		text := string(out)
		if m := speedLimitRegex.FindStringSubmatch(text); len(m) > 1 {
			if v, err := strconv.Atoi(m[1]); err == nil {
				status.CPUSpeedLimit = v
			}
		}
		if m := schedulerLimitRegex.FindStringSubmatch(text); len(m) > 1 {
			if v, err := strconv.Atoi(m[1]); err == nil {
				status.CPUSchedulerLimit = v
			}
		}
		if m := availableCPUsRegex.FindStringSubmatch(text); len(m) > 1 {
			if v, err := strconv.Atoi(m[1]); err == nil {
				status.AvailableCPUs = v
			}
		}
		if status.CPUSpeedLimit < 100 {
			status.IsThrottled = true
		}
	}

	// 2. pmset -g batt
	if out, err := exec.Command("pmset", "-g", "batt").Output(); err == nil {
		text := string(out)
		if strings.Contains(text, "AC Power") {
			status.PowerSource = "AC Power"
		} else if strings.Contains(text, "Battery Power") {
			status.PowerSource = "Battery"
		}

		if m := batteryPctRegex.FindStringSubmatch(text); len(m) > 1 {
			if v, err := strconv.Atoi(m[1]); err == nil {
				status.BatteryPct = v
			}
		}
		if strings.Contains(text, "charging") || strings.Contains(text, "charged") {
			status.IsCharging = true
		}
	}

	// 3. Load averages via sysctl -n vm.loadavg
	if out, err := exec.Command("sysctl", "-n", "vm.loadavg").Output(); err == nil {
		text := strings.Trim(strings.TrimSpace(string(out)), "{ }")
		parts := strings.Fields(text)
		if len(parts) >= 3 {
			if v, err := strconv.ParseFloat(parts[0], 64); err == nil {
				status.LoadAvg1 = v
			}
			if v, err := strconv.ParseFloat(parts[1], 64); err == nil {
				status.LoadAvg5 = v
			}
			if v, err := strconv.ParseFloat(parts[2], 64); err == nil {
				status.LoadAvg15 = v
			}
		}
	}

	return status
}
