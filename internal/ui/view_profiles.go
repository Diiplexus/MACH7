package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"mach7/internal/profiles"
)

func renderProfiles(m Model) string {
	var s strings.Builder

	header := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).
		Render("⚡ SELECT OPTIMIZATION PROFILE")
	sub := lipgloss.NewStyle().Foreground(ColorDim).
		Render("Pre-configured sets of tweaks tested specifically for Intel Core i7 & Iris Plus graphics.")

	s.WriteString(header + "\n" + sub + "\n\n")

	activeProfileID := profiles.DetectActiveProfile(m.registry)

	for i, prof := range m.profiles {
		isFocused := (i == m.profileIndex)
		isApplied := (prof.ID == activeProfileID)
		card := renderProfileCard(m, prof, isFocused, isApplied)
		s.WriteString(card)
		s.WriteString("\n")
	}

	help := lipgloss.NewStyle().Foreground(ColorMuted).
		Render("Navigate: [j/k or ↑/↓] • Apply Selected Profile: [ENTER]")
	s.WriteString("\n" + help)

	return s.String()
}

func renderProfileCard(m Model, prof Profile, isFocused bool, isApplied bool) string {
	borderColor := ColorSubtle
	if isApplied && isFocused {
		borderColor = ColorPrimary
	} else if isApplied {
		borderColor = ColorSuccess
	} else if isFocused {
		borderColor = ColorPrimary
	}

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(94)

	var title string
	cursor := "  "
	if isFocused {
		cursor = "▶ "
	}

	activeBadge := ""
	if isApplied {
		activeBadge = " " + ActiveProfileBadge.Render("✓ CURRENTLY ACTIVE")
	}

	if isFocused {
		title = lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).
			Render(fmt.Sprintf("%s%s  %s  [%s]", cursor, prof.Icon, prof.Name, prof.Tagline)) + activeBadge
	} else if isApplied {
		title = lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).
			Render(fmt.Sprintf("%s%s  %s  [%s]", cursor, prof.Icon, prof.Name, prof.Tagline)) + activeBadge
	} else {
		title = lipgloss.NewStyle().Bold(true).Foreground(ColorText).
			Render(fmt.Sprintf("%s%s  %s  [%s]", cursor, prof.Icon, prof.Name, prof.Tagline))
	}

	desc := lipgloss.NewStyle().Foreground(ColorDim).Render("  " + prof.Description)

	var tweaksList string
	if len(prof.TweakIDs) == 0 {
		tweaksList = lipgloss.NewStyle().Foreground(ColorWarning).
			Render("  • Reverts all modifications back to factory macOS defaults.")
	} else {
		var twNames []string
		for _, twID := range prof.TweakIDs {
			for _, regTw := range m.registry.Tweaks {
				if regTw.ID == twID {
					twNames = append(twNames, regTw.Title)
					break
				}
			}
		}
		tweaksList = lipgloss.NewStyle().Foreground(ColorMuted).
			Render(fmt.Sprintf("  • Includes %d optimizations: %s", len(twNames), strings.Join(twNames, ", ")))
	}

	content := fmt.Sprintf("%s\n%s\n%s", title, desc, tweaksList)
	return cardStyle.Render(content)
}
