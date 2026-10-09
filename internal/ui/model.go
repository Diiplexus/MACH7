package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"mach7/internal/backup"
	"mach7/internal/benchmark"
	"mach7/internal/profiles"
	"mach7/internal/system"
	"mach7/internal/tweaks"
	"mach7/internal/warden"
)

type TabType int

const (
	TabHUD TabType = iota
	TabProfiles
	TabTweaks
	TabWarden
	TabBackups
	TabBenchmarks
)

type Profile = profiles.Profile

type Model struct {
	activeTab     TabType
	hwInfo        system.HardwareInfo
	thermStatus   system.ThermalStatus
	memStatus     system.MemoryStatus
	cpuMetrics    system.CPUMetrics
	registry      *tweaks.Registry
	profiles      []Profile
	wardenList    []warden.DaemonStatus
	snapshots     []backup.Snapshot

	profileIndex  int
	tweakIndex    int
	wardenIndex   int
	backupIndex   int

	lastBench      *benchmark.Result
	baselineBench  *benchmark.Result
	isBenchmarking bool

	toastMsg      string
	toastIsErr    bool

	width         int
	height        int
}

type tickMsg time.Time
type clearToastMsg struct{}
type benchFinishedMsg benchmark.Result

func tickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func clearToastCmd() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return clearToastMsg{}
	})
}

func runBenchmarkCmd(activeTweaks int) tea.Cmd {
	return func() tea.Msg {
		res := benchmark.RunBenchmark(activeTweaks)
		return benchFinishedMsg(res)
	}
}

// InitialModel initializes the Bubble Tea state
func InitialModel() Model {
	hw := system.GetHardwareInfo()
	therm := system.GetThermalStatus()
	mem := system.GetMemoryStatus(hw.TotalRAMBytes)
	cpu := system.GetLiveCPUMetrics()
	reg := tweaks.NewRegistry()
	profs := profiles.GetBuiltinProfiles()
	wardens := warden.ScanDaemons()
	snaps, _ := backup.ListSnapshots()
	base, _ := benchmark.LoadBaseline()

	return Model{
		activeTab:     TabHUD,
		hwInfo:        hw,
		thermStatus:   therm,
		memStatus:     mem,
		cpuMetrics:    cpu,
		registry:      reg,
		profiles:      profs,
		wardenList:    wardens,
		snapshots:     snaps,
		baselineBench: base,
		width:         100,
		height:        35,
	}
}

func (m Model) Init() tea.Cmd {
	return tickCmd()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tickMsg:
		m.thermStatus = system.GetThermalStatus()
		m.memStatus = system.GetMemoryStatus(m.hwInfo.TotalRAMBytes)
		m.cpuMetrics = system.GetLiveCPUMetrics()
		m.wardenList = warden.ScanDaemons()
		cmds = append(cmds, tickCmd())

	case benchFinishedMsg:
		res := benchmark.Result(msg)
		m.lastBench = &res
		m.isBenchmarking = false
		m.toastMsg = fmt.Sprintf("✓ Benchmark complete! Score: %d PTS", res.OverallScore)
		m.toastIsErr = false
		cmds = append(cmds, clearToastCmd())

	case clearToastMsg:
		m.toastMsg = ""
		m.toastIsErr = false

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		// Tab switching via Tab / Shift-Tab
		case "tab":
			m.activeTab = (m.activeTab + 1) % 6

		case "shift+tab":
			if m.activeTab == 0 {
				m.activeTab = 5
			} else {
				m.activeTab--
			}

		// Neovim horizontal tab motion
		case "h", "left":
			if m.activeTab == 0 {
				m.activeTab = 5
			} else {
				m.activeTab--
			}

		case "l", "right":
			m.activeTab = (m.activeTab + 1) % 6

		// Quick numeric tab selection
		case "1":
			m.activeTab = TabHUD
		case "2":
			m.activeTab = TabProfiles
		case "3":
			m.activeTab = TabTweaks
		case "4":
			m.activeTab = TabWarden
		case "5":
			m.activeTab = TabBackups
		case "6":
			m.activeTab = TabBenchmarks

		// Neovim vertical motion: Up (k / up)
		case "up", "k":
			switch m.activeTab {
			case TabProfiles:
				if m.profileIndex > 0 {
					m.profileIndex--
				}
			case TabTweaks:
				if m.tweakIndex > 0 {
					m.tweakIndex--
				}
			case TabWarden:
				if m.wardenIndex > 0 {
					m.wardenIndex--
				}
			case TabBackups:
				if m.backupIndex > 0 {
					m.backupIndex--
				}
			}

		// Neovim vertical motion: Down (j / down)
		case "down", "j":
			switch m.activeTab {
			case TabProfiles:
				if m.profileIndex < len(m.profiles)-1 {
					m.profileIndex++
				}
			case TabTweaks:
				if m.tweakIndex < len(m.registry.Tweaks)-1 {
					m.tweakIndex++
				}
			case TabWarden:
				if m.wardenIndex < len(m.wardenList)-1 {
					m.wardenIndex++
				}
			case TabBackups:
				if m.backupIndex < len(m.snapshots)-1 {
					m.backupIndex++
				}
			}

		// Neovim jump to top (g)
		case "g":
			switch m.activeTab {
			case TabProfiles:
				m.profileIndex = 0
			case TabTweaks:
				m.tweakIndex = 0
			case TabWarden:
				m.wardenIndex = 0
			case TabBackups:
				m.backupIndex = 0
			}

		// Neovim jump to bottom (G)
		case "G":
			switch m.activeTab {
			case TabProfiles:
				m.profileIndex = len(m.profiles) - 1
			case TabTweaks:
				m.tweakIndex = len(m.registry.Tweaks) - 1
			case TabWarden:
				m.wardenIndex = len(m.wardenList) - 1
			case TabBackups:
				if len(m.snapshots) > 0 {
					m.backupIndex = len(m.snapshots) - 1
				}
			}

		// Neovim half-page jump down (ctrl+d)
		case "ctrl+d":
			switch m.activeTab {
			case TabProfiles:
				m.profileIndex = min(m.profileIndex+4, len(m.profiles)-1)
			case TabTweaks:
				m.tweakIndex = min(m.tweakIndex+4, len(m.registry.Tweaks)-1)
			case TabWarden:
				m.wardenIndex = min(m.wardenIndex+4, len(m.wardenList)-1)
			case TabBackups:
				if len(m.snapshots) > 0 {
					m.backupIndex = min(m.backupIndex+4, len(m.snapshots)-1)
				}
			}

		// Neovim half-page jump up (ctrl+u)
		case "ctrl+u":
			switch m.activeTab {
			case TabProfiles:
				m.profileIndex = max(m.profileIndex-4, 0)
			case TabTweaks:
				m.tweakIndex = max(m.tweakIndex-4, 0)
			case TabWarden:
				m.wardenIndex = max(m.wardenIndex-4, 0)
			case TabBackups:
				m.backupIndex = max(m.backupIndex-4, 0)
			}

		// Action: Toggle / Apply / Run Benchmark
		case "enter":
			switch m.activeTab {
			case TabProfiles:
				selected := m.profiles[m.profileIndex]
				var allKeys []backup.KeyDef
				for _, tw := range m.registry.Tweaks {
					allKeys = append(allKeys, tw.Keys...)
				}
				_, _ = backup.CreateSnapshot(fmt.Sprintf("Before applying %s", selected.Name), allKeys)

				err := profiles.ApplyProfile(m.registry, selected)
				if err != nil {
					m.toastMsg = fmt.Sprintf("Profile notice: %v", err)
					m.toastIsErr = false
				} else {
					m.toastMsg = fmt.Sprintf("✓ Successfully applied '%s' profile!", selected.Name)
					m.toastIsErr = false
				}
				m.snapshots, _ = backup.ListSnapshots()
				cmds = append(cmds, clearToastCmd())

			case TabBackups:
				if len(m.snapshots) > 0 {
					snap := m.snapshots[m.backupIndex]
					err := backup.RestoreSnapshot(&snap)
					if err != nil {
						m.toastMsg = fmt.Sprintf("Error restoring snapshot: %v", err)
						m.toastIsErr = true
					} else {
						m.toastMsg = fmt.Sprintf("✓ Restored settings from %s!", snap.ID)
						m.toastIsErr = false
					}
					cmds = append(cmds, clearToastCmd())
				}

			case TabBenchmarks:
				if !m.isBenchmarking {
					m.isBenchmarking = true
					activeCount := 0
					for _, tw := range m.registry.Tweaks {
						if tw.CheckStatusFunc() {
							activeCount++
						}
					}
					cmds = append(cmds, runBenchmarkCmd(activeCount))
				}
			}

		case " ":
			if m.activeTab == TabTweaks {
				tw := m.registry.Tweaks[m.tweakIndex]
				current := tw.CheckStatusFunc()

				if len(tw.Keys) > 0 {
					_, _ = backup.CreateSnapshot(fmt.Sprintf("Before toggling %s", tw.Title), tw.Keys)
				}

				if current {
					if err := tw.RevertFunc(); err != nil {
						m.toastMsg = fmt.Sprintf("Note: %v", err)
						m.toastIsErr = true
					} else {
						m.toastMsg = fmt.Sprintf("Disabled '%s'", tw.Title)
						m.toastIsErr = false
					}
				} else {
					if err := tw.ApplyFunc(); err != nil {
						m.toastMsg = fmt.Sprintf("Note: %v", err)
						m.toastIsErr = true
					} else {
						m.toastMsg = fmt.Sprintf("Enabled '%s'", tw.Title)
						m.toastIsErr = false
					}
				}
				m.snapshots, _ = backup.ListSnapshots()
				cmds = append(cmds, clearToastCmd())
			}

		// Trigger Benchmark from any screen via 'b' / 'B'
		case "b", "B":
			if !m.isBenchmarking {
				m.activeTab = TabBenchmarks
				m.isBenchmarking = true
				activeCount := 0
				for _, tw := range m.registry.Tweaks {
					if tw.CheckStatusFunc() {
						activeCount++
					}
				}
				cmds = append(cmds, runBenchmarkCmd(activeCount))
			}

		// Memory purge
		case "p", "P":
			if m.activeTab == TabHUD {
				_ = tweaks.PurgeInactiveMemory()
				m.memStatus = system.GetMemoryStatus(m.hwInfo.TotalRAMBytes)
				m.toastMsg = "✓ Flushed inactive RAM cache via purge"
				m.toastIsErr = false
				cmds = append(cmds, clearToastCmd())
			} else if m.activeTab == TabWarden && len(m.wardenList) > 0 {
				d := m.wardenList[m.wardenIndex]
				_ = warden.PauseDaemon(d)
				m.toastMsg = fmt.Sprintf("Paused %s", d.Target.Name)
				m.wardenList = warden.ScanDaemons()
				cmds = append(cmds, clearToastCmd())
			}

		// Resume daemon / Restart UI
		case "r", "R":
			if m.activeTab == TabWarden && len(m.wardenList) > 0 {
				d := m.wardenList[m.wardenIndex]
				_ = warden.ResumeDaemon(d)
				m.toastMsg = fmt.Sprintf("Resumed %s", d.Target.Name)
				m.wardenList = warden.ScanDaemons()
				cmds = append(cmds, clearToastCmd())
			} else {
				tweaks.RestartAllUIServices()
				m.toastMsg = "✓ Restarted Dock and Finder services"
				m.toastIsErr = false
				cmds = append(cmds, clearToastCmd())
			}

		// Restart daemon (free memory)
		case "f", "F":
			if m.activeTab == TabWarden && len(m.wardenList) > 0 {
				d := m.wardenList[m.wardenIndex]
				_ = warden.RestartDaemon(d)
				m.toastMsg = fmt.Sprintf("Restarted %s (freed memory)", d.Target.Name)
				m.wardenList = warden.ScanDaemons()
				cmds = append(cmds, clearToastCmd())
			}

		// Freeze all non-critical daemons
		case "a", "A":
			if m.activeTab == TabWarden {
				for _, d := range m.wardenList {
					_ = warden.PauseDaemon(d)
				}
				m.toastMsg = "Paused all tracked background ML daemons"
				m.wardenList = warden.ScanDaemons()
				cmds = append(cmds, clearToastCmd())
			}

		// Snapshot manual create OR Save Benchmark Baseline
		case "s", "S":
			if m.activeTab == TabBenchmarks && m.lastBench != nil {
				_ = benchmark.SaveBaseline(*m.lastBench)
				m.baselineBench = m.lastBench
				m.toastMsg = "✓ Saved current benchmark as Before Baseline!"
				m.toastIsErr = false
				cmds = append(cmds, clearToastCmd())
			} else {
				var allKeys []backup.KeyDef
				for _, tw := range m.registry.Tweaks {
					allKeys = append(allKeys, tw.Keys...)
				}
				snap, err := backup.CreateSnapshot("Manual Baseline Snapshot", allKeys)
				if err != nil {
					m.toastMsg = fmt.Sprintf("Snapshot error: %v", err)
					m.toastIsErr = true
				} else {
					m.toastMsg = fmt.Sprintf("✓ Created snapshot %s", snap.ID)
					m.toastIsErr = false
					m.snapshots, _ = backup.ListSnapshots()
				}
				cmds = append(cmds, clearToastCmd())
			}

		// Clear baseline (in Benchmarks tab)
		case "c", "C":
			if m.activeTab == TabBenchmarks {
				m.baselineBench = nil
				m.toastMsg = "Baseline cleared."
				m.toastIsErr = false
				cmds = append(cmds, clearToastCmd())
			}

		// Factory reset / restore all to stock
		case "x", "X":
			for _, tw := range m.registry.Tweaks {
				_ = tw.RevertFunc()
			}
			tweaks.RestartAllUIServices()
			m.toastMsg = "✓ Reverted all settings to factory stock defaults"
			m.toastIsErr = false
			cmds = append(cmds, clearToastCmd())
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	var s strings.Builder

	// 1. App Header
	title := HeaderStyle.Render("⚡ MACH7 // macOS 26 Intel Optimizer")
	badge := BadgeStyle.Render("i7-1068NG7 • Iris Plus")
	headRow := lipgloss.JoinHorizontal(lipgloss.Center, title, " ", badge)
	s.WriteString(headRow + "\n\n")

	// 2. Tab Bar
	tabs := []string{
		"[1] Dashboard",
		"[2] Profiles",
		"[3] Granular Tweaks",
		"[4] Daemon Warden",
		"[5] Snapshots",
		"[6] Benchmarks",
	}
	var renderedTabs []string
	for i, tab := range tabs {
		if TabType(i) == m.activeTab {
			renderedTabs = append(renderedTabs, TabActive.Render(tab))
		} else {
			renderedTabs = append(renderedTabs, TabInactive.Render(tab))
		}
	}
	tabRow := TabBorder.Render(strings.Join(renderedTabs, " "))
	s.WriteString(tabRow + "\n\n")

	// 3. Main View Content
	switch m.activeTab {
	case TabHUD:
		s.WriteString(renderHUD(m))
	case TabProfiles:
		s.WriteString(renderProfiles(m))
	case TabTweaks:
		s.WriteString(renderTweaks(m))
	case TabWarden:
		s.WriteString(renderWarden(m))
	case TabBackups:
		s.WriteString(renderBackups(m))
	case TabBenchmarks:
		s.WriteString(renderBenchmarks(m))
	}

	// 4. Toast Notification
	if m.toastMsg != "" {
		style := ToastStyle
		if m.toastIsErr {
			style = ToastErrStyle
		}
		s.WriteString("\n\n" + style.Render(m.toastMsg))
	}

	// 5. Global Footer with Neovim motion hints
	s.WriteString("\n\n" + renderGlobalFooter())

	return s.String()
}

func renderGlobalFooter() string {
	shortcuts := []string{
		FooterKey.Render("[h/l]") + " " + FooterDesc.Render("Tabs"),
		FooterKey.Render("[j/k]") + " " + FooterDesc.Render("Navigate"),
		FooterKey.Render("[1-6]") + " " + FooterDesc.Render("Jump"),
		FooterKey.Render("[B]") + " " + FooterDesc.Render("Benchmark"),
		FooterKey.Render("[R]") + " " + FooterDesc.Render("Restart UI"),
		FooterKey.Render("[Q]") + " " + FooterDesc.Render("Quit"),
	}
	return strings.Join(shortcuts, "  •  ")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
