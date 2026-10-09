package system

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// HardwareInfo holds machine hardware and OS specifications
type HardwareInfo struct {
	ModelName      string
	ModelID        string
	CPUBrand       string
	CPUCores       int
	CPULogical     int
	TotalRAMBytes  uint64
	TotalRAMGB     float64
	GPUName        string
	OSProduct      string
	OSVersion      string
	OSBuild        string
	Arch           string
	IsIntelMac     bool
}

// GetHardwareInfo queries macOS sysctl and system_profiler for machine specs
func GetHardwareInfo() HardwareInfo {
	info := HardwareInfo{
		Arch:       runtime.GOARCH,
		CPULogical: runtime.NumCPU(),
	}

	// CPU Brand
	if out, err := exec.Command("sysctl", "-n", "machdep.cpu.brand_string").Output(); err == nil {
		info.CPUBrand = strings.TrimSpace(string(out))
		if strings.Contains(strings.ToLower(info.CPUBrand), "intel") {
			info.IsIntelMac = true
		}
	}

	// Physical cores
	if out, err := exec.Command("sysctl", "-n", "hw.physicalcpu").Output(); err == nil {
		if cores, err := strconv.Atoi(strings.TrimSpace(string(out))); err == nil {
			info.CPUCores = cores
		}
	}

	// RAM
	if out, err := exec.Command("sysctl", "-n", "hw.memsize").Output(); err == nil {
		if mem, err := strconv.ParseUint(strings.TrimSpace(string(out)), 10, 64); err == nil {
			info.TotalRAMBytes = mem
			info.TotalRAMGB = float64(mem) / (1024 * 1024 * 1024)
		}
	}

	// OS info
	if out, err := exec.Command("sw_vers", "-productName").Output(); err == nil {
		info.OSProduct = strings.TrimSpace(string(out))
	}
	if out, err := exec.Command("sw_vers", "-productVersion").Output(); err == nil {
		info.OSVersion = strings.TrimSpace(string(out))
	}
	if out, err := exec.Command("sw_vers", "-buildVersion").Output(); err == nil {
		info.OSBuild = strings.TrimSpace(string(out))
	}

	// Model Identifier
	if out, err := exec.Command("sysctl", "-n", "hw.model").Output(); err == nil {
		info.ModelID = strings.TrimSpace(string(out))
	}

	// Model Name friendly parsing
	if strings.HasPrefix(info.ModelID, "MacBookPro16,2") {
		info.ModelName = "MacBook Pro (13-inch, 2020, 4 TB3 Ports)"
		info.GPUName = "Intel Iris Plus Graphics (1.5GB Shared)"
	} else if strings.HasPrefix(info.ModelID, "MacBookPro") {
		info.ModelName = fmt.Sprintf("MacBook Pro (%s)", info.ModelID)
		info.GPUName = "Intel Iris / AMD Radeon Graphics"
	} else {
		info.ModelName = info.ModelID
		info.GPUName = "Integrated Graphics"
	}

	return info
}
