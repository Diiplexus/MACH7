package system

import (
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// MemoryStatus holds memory usage, pressure and swap info
type MemoryStatus struct {
	TotalBytes      uint64
	UsedBytes       uint64
	FreeBytes       uint64
	WiredBytes      uint64
	ActiveBytes     uint64
	InactiveBytes   uint64
	CompressedBytes uint64
	PressurePct     float64 // 0 - 100%
	SwapTotalMB     float64
	SwapUsedMB      float64
	SwapFreeMB      float64
}

var (
	swapRegex = regexp.MustCompile(`total\s*=\s*([0-9.]+)([MGK])\s+used\s*=\s*([0-9.]+)([MGK])\s+free\s*=\s*([0-9.]+)([MGK])`)
)

// GetMemoryStatus parses vm_stat and vm.swapusage
func GetMemoryStatus(totalRAM uint64) MemoryStatus {
	status := MemoryStatus{
		TotalBytes: totalRAM,
	}

	out, err := exec.Command("vm_stat").Output()
	if err == nil {
		lines := strings.Split(string(out), "\n")
		pageSize := uint64(4096)

		var freePages, activePages, inactivePages, wiredPages, speculativePages, compressedPages uint64

		for _, line := range lines {
			parts := strings.Split(line, ":")
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			valStr := strings.Trim(strings.TrimSpace(parts[1]), ".")
			val, _ := strconv.ParseUint(valStr, 10, 64)

			switch key {
			case "Pages free":
				freePages = val
			case "Pages active":
				activePages = val
			case "Pages inactive":
				inactivePages = val
			case "Pages speculative":
				speculativePages = val
			case "Pages wired down":
				wiredPages = val
			case "Pages occupied by compressor":
				compressedPages = val
			}
		}

		status.FreeBytes = (freePages + speculativePages) * pageSize
		status.ActiveBytes = activePages * pageSize
		status.InactiveBytes = inactivePages * pageSize
		status.WiredBytes = wiredPages * pageSize
		status.CompressedBytes = compressedPages * pageSize

		status.UsedBytes = status.ActiveBytes + status.WiredBytes + status.CompressedBytes
		if totalRAM > 0 {
			status.PressurePct = (float64(status.UsedBytes) / float64(totalRAM)) * 100.0
			if status.PressurePct > 100 {
				status.PressurePct = 100
			}
		}
	}

	// Swap usage
	if swapOut, err := exec.Command("sysctl", "-n", "vm.swapusage").Output(); err == nil {
		if m := swapRegex.FindStringSubmatch(string(swapOut)); len(m) >= 7 {
			status.SwapTotalMB = parseUnitToMB(m[1], m[2])
			status.SwapUsedMB = parseUnitToMB(m[3], m[4])
			status.SwapFreeMB = parseUnitToMB(m[5], m[6])
		}
	}

	return status
}

func parseUnitToMB(valStr, unit string) float64 {
	val, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return 0
	}
	switch strings.ToUpper(unit) {
	case "G":
		return val * 1024
	case "M":
		return val
	case "K":
		return val / 1024
	default:
		return val
	}
}
