package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"mach7/internal/backup"
)

func renderBackups(m Model) string {
	var s strings.Builder

	header := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).
		Render("💾 SNAPSHOTS & ZERO-RISK ROLLBACK")
	sub := lipgloss.NewStyle().Foreground(ColorDim).
		Render("mach7 takes automatic snapshots of your macOS defaults before modifying anything.")

	s.WriteString(header + "\n" + sub + "\n\n")

	if len(m.snapshots) == 0 {
		emptyBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSubtle).
			Padding(1, 2).
			Width(94).
			Render(lipgloss.NewStyle().Foreground(ColorDim).Render(
				"No snapshots saved yet.\n" +
					"A snapshot is automatically created whenever you apply a profile or tweak.\n" +
					"Press [S] to create a manual baseline snapshot now."))
		s.WriteString(emptyBox + "\n")
	} else {
		for i, snap := range m.snapshots {
			isSelected := (i == m.backupIndex)
			s.WriteString(renderSnapshotCard(snap, isSelected))
			s.WriteString("\n")
		}
	}

	help := lipgloss.NewStyle().Foreground(ColorMuted).
		Render("\nNavigate: [↑/↓] • Restore Selected: [ENTER] • Create New Snapshot: [S] • Factory Reset: [X]")
	s.WriteString(help)

	return s.String()
}

func renderSnapshotCard(snap backup.Snapshot, isSelected bool) string {
	borderColor := ColorSubtle
	if isSelected {
		borderColor = ColorPrimary
	}

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(94)

	cursor := "  "
	if isSelected {
		cursor = lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render("▶ ")
	}

	timeStr := snap.CreatedAt.Format("2006-01-02 15:04:05")
	title := fmt.Sprintf("%s📦 Snapshot: %s  (%s)", cursor, snap.ID, timeStr)
	desc := fmt.Sprintf("   • Reason: %s\n   • Backed up keys: %d settings",
		snap.Reason, len(snap.Entries))

	return cardStyle.Render(fmt.Sprintf("%s\n%s", title, desc))
}
