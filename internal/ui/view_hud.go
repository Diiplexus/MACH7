package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"mach7/internal/profiles"
)

func renderHUD(m Model) string {
	var s strings.Builder

	// Top summary cards: Specs + Thermals
	specsBox := renderSpecsCard(m)
	thermBox := renderThermalsCard(m)
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, specsBox, "  ", thermBox)

	s.WriteString(topRow)
	s.WriteString("\n\n")

	// Real-Time Activity Card (Instant CPU Usage, User/Sys split, Disk I/O)
	realtimeBox := renderRealtimeCard(m)
	s.WriteString(realtimeBox)
	s.WriteString("\n\n")

	// Memory & Swap Card
	memBox := renderMemoryCard(m)
	s.WriteString(memBox)
	s.WriteString("\n\n")

	// Active Profile & Benchmarks Status Card
	statusBox := renderStatusCard(m)
	s.WriteString(statusBox)

	return s.String()
}

func renderSpecsCard(m Model) string {
	b := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorSubtle).
		Padding(0, 1).
		Width(46)

	title := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("💻 HARDWARE OVERVIEW")
	hw := m.hwInfo

	content := fmt.Sprintf("%s\n\n"+
		"• Model:      %s\n"+
		"• Processor:  %s\n"+
		"• Cores:      %d Cores / %d Threads\n"+
		"• Graphics:   %s\n"+
		"• Memory:     %.1f GB LPDDR4X\n"+
		"• OS:         %s %s (%s)",
		title,
		hw.ModelName,
		hw.CPUBrand,
		hw.CPUCores, hw.CPULogical,
		hw.GPUName,
		hw.TotalRAMGB,
		hw.OSProduct, hw.OSVersion, hw.OSBuild,
	)

	return b.Render(content)
}

func renderThermalsCard(m Model) string {
	b := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorSubtle).
		Padding(0, 1).
		Width(46)

	title := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("🌡️ THERMALS & POWER")
	th := m.thermStatus

	throttleColor := ColorSuccess
	throttleText := "100% (No Throttling)"
	if th.IsThrottled {
		throttleColor = ColorDanger
		throttleText = fmt.Sprintf("%d%% (THROTTLED!)", th.CPUSpeedLimit)
	}

	powerIcon := "⚡"
	if th.PowerSource == "Battery" {
		powerIcon = "🔋"
	}

	content := fmt.Sprintf("%s\n\n"+
		"• CPU Speed Limit:    %s\n"+
		"• Available CPUs:     %d Cores\n"+
		"• Power Source:       %s %s (%d%%)\n"+
		"• CPU Load (1m/5m):   %.2f / %.2f\n"+
		"• Long Load (15m):    %.2f\n"+
		"• Thermal Envelope:   28W Shared Package",
		title,
		lipgloss.NewStyle().Foreground(throttleColor).Bold(true).Render(throttleText),
		th.AvailableCPUs,
		powerIcon, th.PowerSource, th.BatteryPct,
		th.LoadAvg1, th.LoadAvg5,
		th.LoadAvg15,
	)

	return b.Render(content)
}

func renderRealtimeCard(m Model) string {
	b := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorSubtle).
		Padding(0, 1).
		Width(94)

	title := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("⚡ REAL-TIME CPU & DISK I/O ACTIVITY")
	cpu := m.cpuMetrics

	cpuBar := renderProgressBar(float64(cpu.TotalUsagePct), 35)

	content := fmt.Sprintf("%s\n\n"+
		"Real-Time CPU:  [%s] %3d%% Total  (User: %d%% | System: %d%% | Idle: %d%%)\n"+
		"Live NVMe I/O:  Throughput: %.2f MB/s  •  Transactions: %d TPS\n"+
		"💡 Real-time stats refreshed every 2 seconds via macOS Mach kernel telemetry.",
		title,
		cpuBar, cpu.TotalUsagePct, cpu.UserPct, cpu.SystemPct, cpu.IdlePct,
		cpu.DiskMBs, cpu.DiskTPS,
	)

	return b.Render(content)
}

func renderMemoryCard(m Model) string {
	b := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorSubtle).
		Padding(0, 1).
		Width(94)

	title := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("🧠 MEMORY PRESSURE & SWAP USAGE")
	mem := m.memStatus

	usedGB := float64(mem.UsedBytes) / (1024 * 1024 * 1024)
	totalGB := m.hwInfo.TotalRAMGB
	pressureBar := renderProgressBar(mem.PressurePct, 35)

	swapText := "No active swap"
	if mem.SwapTotalMB > 0 {
		swapText = fmt.Sprintf("%.1f MB used / %.1f MB total (%.1f MB free)",
			mem.SwapUsedMB, mem.SwapTotalMB, mem.SwapFreeMB)
	}

	content := fmt.Sprintf("%s\n\n"+
		"Memory Pressure: [%s] %.1f%%  (%.1f GB used / %.1f GB total)\n"+
		"• Active: %.1f GB   • Wired: %.1f GB   • Inactive: %.1f GB   • Compressed: %.1f GB\n"+
		"• NVMe Swap: %s\n"+
		"💡 Press [P] to purge inactive memory cache and relieve swap.",
		title,
		pressureBar, mem.PressurePct, usedGB, totalGB,
		float64(mem.ActiveBytes)/(1024*1024*1024),
		float64(mem.WiredBytes)/(1024*1024*1024),
		float64(mem.InactiveBytes)/(1024*1024*1024),
		float64(mem.CompressedBytes)/(1024*1024*1024),
		swapText,
	)

	return b.Render(content)
}

func renderStatusCard(m Model) string {
	b := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorSubtle).
		Padding(0, 1).
		Width(94)

	title := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Render("🎯 ACTIVE PROFILE & PERFORMANCE SUMMARY")

	// Detect currently applied profile
	activeID := profiles.DetectActiveProfile(m.registry)
	profileName := "Custom Mix"
	for _, p := range m.profiles {
		if p.ID == activeID {
			profileName = fmt.Sprintf("%s %s (%s)", p.Icon, p.Name, p.Tagline)
			break
		}
	}

	activeCount := 0
	for _, tw := range m.registry.Tweaks {
		if tw.CheckStatusFunc() {
			activeCount++
		}
	}

	benchSummary := "No benchmark run yet. Press [B] to run a live performance benchmark!"
	if m.lastBench != nil {
		benchSummary = fmt.Sprintf("Latest Benchmark Score: %d  (CPU: %d | Mem: %d | UI Latency Index: %d)",
			m.lastBench.OverallScore, m.lastBench.CPUScore, m.lastBench.MemScore, m.lastBench.UILatencyScore)
	}

	content := fmt.Sprintf("%s\n\n"+
		"• Applied Profile:   %s\n"+
		"• Active Tweaks:     %d / %d Enabled\n"+
		"• Performance Test:  %s\n"+
		"💡 Press [B] to run benchmark test • Press [2] to switch profile.",
		title,
		lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Render(profileName),
		activeCount, len(m.registry.Tweaks),
		benchSummary,
	)

	return b.Render(content)
}

func renderProgressBar(pct float64, width int) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}

	filled := int((pct / 100.0) * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled

	fillColor := ColorSuccess
	if pct > 75 {
		fillColor = ColorDanger
	} else if pct > 55 {
		fillColor = ColorWarning
	}

	bar := lipgloss.NewStyle().Foreground(fillColor).Render(strings.Repeat("█", filled)) +
		lipgloss.NewStyle().Foreground(ColorSubtle).Render(strings.Repeat("░", empty))
	return bar
}
