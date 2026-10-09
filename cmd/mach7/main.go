package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"mach7/internal/benchmark"
	"mach7/internal/profiles"
	"mach7/internal/system"
	"mach7/internal/tweaks"
	"mach7/internal/ui"
)

var (
	flagProfile   = flag.String("profile", "", "Apply optimization profile directly (ventura, max, dev, stock)")
	flagStatus    = flag.Bool("status", false, "Print system thermals, memory and status summary")
	flagBench     = flag.Bool("bench", false, "Run performance benchmark suite and display score")
	flagSaveBase  = flag.Bool("save-baseline", false, "Save current benchmark as Before Baseline")
	flagPurge     = flag.Bool("purge", false, "Purge inactive RAM cache and exit")
	flagRestore   = flag.Bool("restore", false, "Revert all settings to macOS stock defaults and exit")
)

func main() {
	flag.Parse()

	// 1. Headless CLI handlers
	if *flagStatus {
		printStatus()
		return
	}

	if *flagBench {
		runCLIBenchmark(*flagSaveBase)
		return
	}

	if *flagPurge {
		fmt.Println("⚡ Purging inactive RAM cache...")
		if err := tweaks.PurgeInactiveMemory(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✓ Inactive RAM purged successfully.")
		return
	}

	if *flagRestore {
		fmt.Println("🍏 Restoring all settings to stock macOS defaults...")
		reg := tweaks.NewRegistry()
		for _, tw := range reg.Tweaks {
			_ = tw.RevertFunc()
		}
		tweaks.RestartAllUIServices()
		fmt.Println("✓ Reverted all tweaks and restarted Dock & Finder.")
		return
	}

	if *flagProfile != "" {
		handleDirectProfile(*flagProfile)
		return
	}

	// 2. Interactive TUI Mode
	p := tea.NewProgram(ui.InitialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running mach7 TUI: %v\n", err)
		os.Exit(1)
	}
}

func handleDirectProfile(name string) {
	reg := tweaks.NewRegistry()
	var targetProfile *profiles.Profile

	builtinProfs := profiles.GetBuiltinProfiles()
	switch name {
	case "ventura", "balanced":
		targetProfile = &builtinProfs[0]
	case "max", "fps", "performance":
		targetProfile = &builtinProfs[1]
	case "dev", "developer":
		targetProfile = &builtinProfs[2]
	case "stock", "default", "reset":
		targetProfile = &builtinProfs[3]
	default:
		fmt.Printf("Unknown profile '%s'. Available: ventura, max, dev, stock\n", name)
		os.Exit(1)
	}

	fmt.Printf("⚡ Applying '%s' (%s)...\n", targetProfile.Name, targetProfile.Tagline)
	if err := profiles.ApplyProfile(reg, *targetProfile); err != nil {
		fmt.Printf("Profile notice: %v\n", err)
	} else {
		fmt.Printf("✓ Successfully applied profile: %s\n", targetProfile.Name)
	}
}

func printStatus() {
	hw := system.GetHardwareInfo()
	therm := system.GetThermalStatus()
	mem := system.GetMemoryStatus(hw.TotalRAMBytes)
	cpu := system.GetLiveCPUMetrics()
	reg := tweaks.NewRegistry()

	activeCount := 0
	for _, tw := range reg.Tweaks {
		if tw.CheckStatusFunc() {
			activeCount++
		}
	}

	activeProfileID := profiles.DetectActiveProfile(reg)
	profileName := "Custom Mix"
	for _, p := range profiles.GetBuiltinProfiles() {
		if p.ID == activeProfileID {
			profileName = fmt.Sprintf("%s %s", p.Icon, p.Name)
			break
		}
	}

	fmt.Println("==================================================")
	fmt.Printf(" ⚡ MACH7 // System Telemetry & Status\n")
	fmt.Println("==================================================")
	fmt.Printf(" • Machine:        %s\n", hw.ModelName)
	fmt.Printf(" • CPU:            %s (%d cores / %d threads)\n", hw.CPUBrand, hw.CPUCores, hw.CPULogical)
	fmt.Printf(" • Active Profile: %s (%d / %d Tweaks)\n", profileName, activeCount, len(reg.Tweaks))
	fmt.Printf(" • Real-Time CPU:  %d%% (User: %d%% | Sys: %d%% | Idle: %d%%)\n", cpu.TotalUsagePct, cpu.UserPct, cpu.SystemPct, cpu.IdlePct)
	fmt.Printf(" • Disk I/O:       %.2f MB/s (%d TPS)\n", cpu.DiskMBs, cpu.DiskTPS)
	fmt.Printf(" • RAM Pressure:   %.1f%% (%.1f GB used / %.1f GB total)\n", mem.PressurePct, float64(mem.UsedBytes)/(1024*1024*1024), hw.TotalRAMGB)
	fmt.Printf(" • Swap Used:      %.1f MB / %.1f MB\n", mem.SwapUsedMB, mem.SwapTotalMB)
	fmt.Printf(" • CPU Throttled:  %v (Speed Limit: %d%%)\n", therm.IsThrottled, therm.CPUSpeedLimit)
	fmt.Printf(" • Power Source:   %s (%d%%)\n", therm.PowerSource, therm.BatteryPct)
	fmt.Println("==================================================")
}

func runCLIBenchmark(saveBase bool) {
	reg := tweaks.NewRegistry()
	activeCount := 0
	for _, tw := range reg.Tweaks {
		if tw.CheckStatusFunc() {
			activeCount++
		}
	}

	fmt.Println("⚡ Running MACH7 Performance Benchmark Suite across 8 threads...")
	fmt.Println("  [1/3] Testing multi-thread cryptographic compute & thread contention...")
	fmt.Println("  [2/3] Testing memory allocation & read/write bandwidth...")
	fmt.Println("  [3/3] Testing cumulative perceived UI animation wait latency...")

	res := benchmark.RunBenchmark(activeCount)

	fmt.Println("\n==================================================")
	fmt.Printf(" 📊 BENCHMARK RESULTS (%s)\n", res.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Println("==================================================")
	fmt.Printf(" • Overall System Score:   %d PTS\n", res.OverallScore)
	fmt.Printf(" • Multi-Thread CPU Score: %d  (%.0f ops/s)\n", res.CPUScore, res.CPUOpsPerSec)
	fmt.Printf(" • Memory Bandwidth Score: %d  (%.2f GB/s)\n", res.MemScore, res.MemThroughputGBs)
	fmt.Printf(" • UI Latency Index Score: %d  (%.0f ms cumulative animation wait)\n", res.UILatencyScore, res.AnimationDelayMs)
	fmt.Printf(" • Thermal Throttled:      %v\n", res.IsThrottled)
	fmt.Println("==================================================")

	base, _ := benchmark.LoadBaseline()
	if base != nil {
		comp := benchmark.CompareResults(*base, res)
		fmt.Println("\n📈 COMPARISON vs BASELINE:")
		fmt.Printf(" • Overall Score Delta:       %+.1f%%\n", comp.OverallDeltaPct)
		fmt.Printf(" • CPU Throughput Delta:      %+.1f%%\n", comp.CPUDeltaPct)
		fmt.Printf(" • Memory Throughput Delta:   %+.1f%%\n", comp.MemDeltaPct)
		fmt.Printf(" • UI Animation Delay Delta:  %+.1f%% (less delay = snappier)\n", comp.UIDelayDeltaPct)
	}

	if saveBase {
		_ = benchmark.SaveBaseline(res)
		fmt.Println("\n✓ Saved current benchmark as Before Baseline in ~/.mach7/baseline.json")
	} else if base == nil {
		fmt.Println("\n💡 Tip: Run with --save-baseline to store this as your 'Before' baseline!")
	}
}
