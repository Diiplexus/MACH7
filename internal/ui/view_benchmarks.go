package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"mach7/internal/benchmark"
)

func renderBenchmarks(m Model) string {
	var s strings.Builder

	header := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).
		Render("📊 PERFORMANCE BENCHMARKS & BEFORE/AFTER TESTING")
	sub := lipgloss.NewStyle().Foreground(ColorDim).
		Render("Measure multi-thread CPU crunch, memory I/O, and perceived UI animation delay.")

	s.WriteString(header + "\n" + sub + "\n\n")

	if m.isBenchmarking {
		runningBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorWarning).
			Padding(1, 2).
			Width(94).
			Render(lipgloss.NewStyle().Bold(true).Foreground(ColorWarning).
				Render("⏳ Running benchmark suite across 8 threads... (Crunching SHA256 & memory I/O, please wait ~2.5s)"))
		s.WriteString(runningBox + "\n\n")
		return s.String()
	}

	if m.lastBench == nil {
		initialBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSubtle).
			Padding(1, 2).
			Width(94).
			Render(lipgloss.NewStyle().Foreground(ColorText).Render(
				"No benchmark has been run during this session.\n\n" +
					"Press [B] or [ENTER] to execute a live benchmark test!\n" +
					"• Test 1: Parallel CPU cryptographic crunch across all 8 cores/threads\n" +
					"• Test 2: Memory sequential & random stride allocation bandwidth\n" +
					"• Test 3: Perceived UI animation wait index & defaults roundtrip latency\n\n" +
					"💡 Workflow: Run test under stock defaults, press [S] to save as Baseline, then apply Maximum FPS and compare!"))
		s.WriteString(initialBox + "\n\n")
	} else {
		// Render current benchmark card
		currentCard := renderBenchmarkResultCard("⚡ CURRENT BENCHMARK RUN", *m.lastBench, ColorPrimary)
		s.WriteString(currentCard + "\n\n")

		// If baseline exists, render comparison card
		if m.baselineBench != nil {
			compCard := renderComparisonCard(*m.baselineBench, *m.lastBench)
			s.WriteString(compCard + "\n\n")
		} else {
			saveHint := lipgloss.NewStyle().Foreground(ColorWarning).
				Render("💡 Tip: Press [S] to save this run as your 'Before' Baseline, then switch profile and run [B] again to see the delta!")
			s.WriteString(saveHint + "\n\n")
		}
	}

	help := lipgloss.NewStyle().Foreground(ColorMuted).
		Render("Run Benchmark: [B / ENTER] • Save as Baseline: [S] • Clear Baseline: [C] • Navigate Tabs: [h/l or 1-6]")
	s.WriteString(help)

	return s.String()
}

func renderBenchmarkResultCard(titleStr string, res benchmark.Result, borderColor lipgloss.Color) string {
	b := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(94)

	title := lipgloss.NewStyle().Bold(true).Foreground(borderColor).Render(titleStr)
	timeStr := res.Timestamp.Format("15:04:05")

	throttleStatus := lipgloss.NewStyle().Foreground(ColorSuccess).Render("None (100% Speed)")
	if res.IsThrottled {
		throttleStatus = lipgloss.NewStyle().Foreground(ColorDanger).Bold(true).Render("THROTTLED DETECTED")
	}

	content := fmt.Sprintf("%s  (%s  •  Active Tweaks: %d/10)\n\n"+
		"• Overall Score:           %s\n"+
		"• Multi-Thread CPU Score:  %d  (%.0f operations/sec across 8 threads)\n"+
		"• Memory Throughput Score: %d  (%.2f GB/s allocation bandwidth)\n"+
		"• UI Latency Index Score:  %d  (Cumulative Animation Wait: %.0f ms)\n"+
		"• Thermal Throttle State:  %s",
		title, timeStr, res.ActiveTweaksCount,
		ScoreBadge.Render(fmt.Sprintf("%d PTS", res.OverallScore)),
		res.CPUScore, res.CPUOpsPerSec,
		res.MemScore, res.MemThroughputGBs,
		res.UILatencyScore, res.AnimationDelayMs,
		throttleStatus,
	)

	return b.Render(content)
}

func renderComparisonCard(base, curr benchmark.Result) string {
	b := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorSuccess).
		Padding(0, 1).
		Width(94)

	comp := benchmark.CompareResults(base, curr)

	title := lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).
		Render("📈 BEFORE vs AFTER IMPROVEMENT DELTA")

	formatDelta := func(val float64, invert bool) string {
		isGood := val > 0
		if invert {
			isGood = val < 0
		}
		prefix := "+"
		if val < 0 {
			prefix = ""
		}
		text := fmt.Sprintf("%s%.1f%%", prefix, val)
		if isGood {
			return DeltaGood.Render(text + " ✓ (Faster)")
		}
		return DeltaBad.Render(text)
	}

	content := fmt.Sprintf("%s\n\n"+
		"• Overall Score Delta:       %s  (%d → %d)\n"+
		"• CPU Throughput Delta:      %s  (%.0f → %.0f ops/s)\n"+
		"• Memory Throughput Delta:   %s  (%.2f → %.2f GB/s)\n"+
		"• UI Animation Delay Delta:  %s  (%.0f ms → %.0f ms cumulative wait)\n\n"+
		"Summary: mach7 optimizations reduced artificial UI delay by %.0f ms and restored smooth responsiveness.",
		title,
		formatDelta(comp.OverallDeltaPct, false), base.OverallScore, curr.OverallScore,
		formatDelta(comp.CPUDeltaPct, false), base.CPUOpsPerSec, curr.CPUOpsPerSec,
		formatDelta(comp.MemDeltaPct, false), base.MemThroughputGBs, curr.MemThroughputGBs,
		formatDelta(comp.UIDelayDeltaPct, true), base.AnimationDelayMs, curr.AnimationDelayMs,
		base.AnimationDelayMs-curr.AnimationDelayMs,
	)

	return b.Render(content)
}
