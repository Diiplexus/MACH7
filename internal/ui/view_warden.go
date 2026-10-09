package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"mach7/internal/warden"
)

func renderWarden(m Model) string {
	var s strings.Builder

	header := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).
		Render("🛡️ DAEMON WARDEN (INTEL AVX2 THROTTLER)")
	sub := lipgloss.NewStyle().Foreground(ColorDim).
		Render("Monitors and controls background AI & indexing daemons that overheat Intel CPUs.")

	s.WriteString(header + "\n" + sub + "\n\n")

	infoBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorSubtle).
		Padding(0, 1).
		Width(94).
		Render(lipgloss.NewStyle().Foreground(ColorWarning).Render(
			"⚠️ Why this matters on Intel Core i7:\n" +
				"Modern macOS features run background machine learning and photo/suggestion indexing.\n" +
				"Without Apple Silicon's Neural Engine, these run via AVX2 instructions on your 4-core i7,\n" +
				"pushing the 28W package to 95°C and causing WindowServer micro-stutters."))

	s.WriteString(infoBox + "\n\n")

	for i, d := range m.wardenList {
		isSelected := (i == m.wardenIndex)
		s.WriteString(renderDaemonRow(d, isSelected))
		s.WriteString("\n")
	}

	help := lipgloss.NewStyle().Foreground(ColorMuted).
		Render("\nNavigate: [↑/↓] • [P] Freeze Daemon • [R] Resume Daemon • [F] Restart / Free RAM • [A] Freeze All")
	s.WriteString(help)

	return s.String()
}

func renderDaemonRow(d warden.DaemonStatus, isSelected bool) string {
	cursor := "  "
	if isSelected {
		cursor = lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render("▶ ")
	}

	statusText := lipgloss.NewStyle().Foreground(ColorMuted).Render("○ INACTIVE")
	if d.IsRunning {
		pidStr := fmt.Sprintf("%v", d.PIDs)
		statusText = lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true).
			Render(fmt.Sprintf("● RUNNING (PID %s)", pidStr))
	}

	impact := ImpactBadge.Render(d.Target.Impact)

	titleStyle := lipgloss.NewStyle().Foreground(ColorText)
	if isSelected {
		titleStyle = lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true)
	}

	titleLine := fmt.Sprintf("%s%s  %s  (%s)  %s",
		cursor, statusText, titleStyle.Render(d.Target.Name), d.Target.Binary, impact)

	descLine := fmt.Sprintf("      %s",
		lipgloss.NewStyle().Foreground(ColorDim).Render(d.Target.Description))

	return fmt.Sprintf("%s\n%s", titleLine, descLine)
}
