package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	ColorPrimary   = lipgloss.Color("#00D8F6") // Electric Cyan
	ColorSecondary = lipgloss.Color("#6366F1") // Indigo
	ColorSuccess   = lipgloss.Color("#10B981") // Emerald
	ColorWarning   = lipgloss.Color("#F59E0B") // Amber
	ColorDanger    = lipgloss.Color("#EF4444") // Rose
	ColorMuted     = lipgloss.Color("#64748B") // Slate 500
	ColorSubtle    = lipgloss.Color("#334155") // Slate 700
	ColorText      = lipgloss.Color("#F8FAFC") // Slate 50
	ColorDim       = lipgloss.Color("#94A3B8") // Slate 400
	ColorBgCard    = lipgloss.Color("#0F172A") // Slate 900
	ColorHighlight = lipgloss.Color("#1E293B") // Slate 800

	// App Header
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Padding(0, 1)

	BadgeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(ColorSecondary).
			Padding(0, 1)

	HardwareBadge = lipgloss.NewStyle().
			Foreground(ColorDim).
			Background(ColorSubtle).
			Padding(0, 1)

	// Tabs
	TabActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#000000")).
			Background(ColorPrimary).
			Padding(0, 2)

	TabInactive = lipgloss.NewStyle().
			Foreground(ColorDim).
			Background(ColorSubtle).
			Padding(0, 2)

	TabBorder = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(ColorSubtle)

	// Cards & Containers
	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSubtle).
			Padding(0, 1).
			MarginBottom(1)

	CardActiveStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(0, 1).
			MarginBottom(1)

	// Items
	SelectedItem = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Background(ColorHighlight).
			Padding(0, 1)

	UnselectedItem = lipgloss.NewStyle().
			Foreground(ColorText).
			Padding(0, 1)

	// Status Badges
	ActiveBadge = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorSuccess).
			SetString("● ON")

	InactiveBadge = lipgloss.NewStyle().
			Foreground(ColorMuted).
			SetString("○ OFF")

	ImpactBadge = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Background(lipgloss.Color("#083344")).
			Padding(0, 1)

	// Notifications / Toast
	ToastStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(ColorSuccess).
			Padding(0, 2)

	ToastErrStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(ColorDanger).
			Padding(0, 2)

	// Profile Badges
	ActiveProfileBadge = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#000000")).
			Background(ColorSuccess).
			Padding(0, 1)

	// Deltas & Benchmark styles
	DeltaGood = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorSuccess)

	DeltaBad = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorDanger)

	ScoreBadge = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Background(ColorHighlight).
			Padding(0, 1)

	// Footer / Shortcuts
	FooterKey = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary)

	FooterDesc = lipgloss.NewStyle().
			Foreground(ColorDim)
)
