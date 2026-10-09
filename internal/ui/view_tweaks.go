package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"mach7/internal/tweaks"
)

func renderTweaks(m Model) string {
	var s strings.Builder

	header := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).
		Render("🛠️ GRANULAR OPTIMIZATIONS")
	sub := lipgloss.NewStyle().Foreground(ColorDim).
		Render("Toggle specific performance tweaks individually. Press [SPACE] to toggle.")

	s.WriteString(header + "\n" + sub + "\n\n")

	currentCat := tweaks.Category("")

	for i, tw := range m.registry.Tweaks {
		if tw.Category != currentCat {
			currentCat = tw.Category
			catHeader := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).
				Render(fmt.Sprintf("\n─── %s ───", string(currentCat)))
			s.WriteString(catHeader + "\n")
		}

		isSelected := (i == m.tweakIndex)
		isActive := tw.CheckStatusFunc()

		s.WriteString(renderTweakRow(tw, isSelected, isActive))
		s.WriteString("\n")
	}

	help := lipgloss.NewStyle().Foreground(ColorMuted).
		Render("\nNavigate: [↑/↓] • Toggle Tweak: [SPACE] • Restart Dock/Finder: [R]")
	s.WriteString(help)

	return s.String()
}

func renderTweakRow(tw tweaks.Tweak, isSelected bool, isActive bool) string {
	cursor := "  "
	if isSelected {
		cursor = lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render("▶ ")
	}

	statusBadge := InactiveBadge.Render()
	if isActive {
		statusBadge = ActiveBadge.Render()
	}

	impact := ImpactBadge.Render(tw.Impact)

	titleStyle := lipgloss.NewStyle().Foreground(ColorText)
	if isSelected {
		titleStyle = lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true)
	}

	titleLine := fmt.Sprintf("%s%s  %s  %s",
		cursor, statusBadge, titleStyle.Render(tw.Title), impact)

	descLine := fmt.Sprintf("      %s",
		lipgloss.NewStyle().Foreground(ColorDim).Render(tw.Description))

	return fmt.Sprintf("%s\n%s", titleLine, descLine)
}
