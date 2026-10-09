package profiles

import (
	"fmt"
	"strings"

	"mach7/internal/tweaks"
)

// Profile represents an optimization preset
type Profile struct {
	ID          string
	Name        string
	Tagline     string
	Description string
	Icon        string
	TweakIDs    []string
}

// GetBuiltinProfiles returns the curated profiles
func GetBuiltinProfiles() []Profile {
	return []Profile{
		{
			ID:          "ventura_feel",
			Name:        "Ventura Smoothness",
			Tagline:     "Balanced • Modern Aesthetics Preserved",
			Description: "Eliminates artificial animation delays while preserving full transparency, Stage Manager fluidness, and macOS 26 visuals.",
			Icon:        "🌊",
			TweakIDs: []string{
				"window_resize_time",
				"quicklook_speed",
				"dock_autohide_delay",
				"dock_autohide_speed",
				"mission_control_speed",
				"dialog_expand",
			},
		},
		{
			ID:          "max_performance",
			Name:        "Maximum FPS & Thermal Headroom",
			Tagline:     "Ultra Snappy • GPU & Thermal Relief",
			Description: "Reduces heavy GPU transparency passes, disables icon bounces and zoom animations, freeing Iris Plus GPU and CPU thermal budget.",
			Icon:        "⚡",
			TweakIDs: []string{
				"window_resize_time",
				"reduce_transparency",
				"quicklook_speed",
				"dock_autohide_delay",
				"dock_autohide_speed",
				"mission_control_speed",
				"dock_launch_anim",
				"finder_animations",
				"dialog_expand",
				"spotlight_dev_protection",
			},
		},
		{
			ID:          "developer_boost",
			Name:        "Developer & Compiler Boost",
			Tagline:     "Low I/O • Protected Build Caches",
			Description: "Optimizes UI latency and prevents Spotlight mds_stores from indexing node_modules, Go pkg, Cargo, and caches.",
			Icon:        "🛠️",
			TweakIDs: []string{
				"window_resize_time",
				"quicklook_speed",
				"dock_autohide_delay",
				"dock_autohide_speed",
				"finder_animations",
				"dialog_expand",
				"spotlight_dev_protection",
			},
		},
		{
			ID:          "stock_defaults",
			Name:        "Stock macOS (Reset to Default)",
			Tagline:     "Factory Reset • Revert All Tweaks",
			Description: "Reverts all modified animation speeds, transparency, and Dock/Finder preferences to out-of-the-box Apple defaults.",
			Icon:        "🍏",
			TweakIDs:    []string{}, // Empty means revert everything
		},
	}
}

// ApplyProfile applies a given profile's tweaks
func ApplyProfile(reg *tweaks.Registry, prof Profile) error {
	if prof.ID == "stock_defaults" {
		// Revert all
		for _, tw := range reg.Tweaks {
			_ = tw.RevertFunc()
		}
		tweaks.RestartAllUIServices()
		return nil
	}

	activeMap := make(map[string]bool)
	for _, id := range prof.TweakIDs {
		activeMap[id] = true
	}

	var warnings []string
	for _, tw := range reg.Tweaks {
		if activeMap[tw.ID] {
			// If already active, skip re-applying
			if tw.CheckStatusFunc() {
				continue
			}
			if err := tw.ApplyFunc(); err != nil {
				warnings = append(warnings, fmt.Sprintf("%s (%v)", tw.Title, err))
			}
		}
	}

	tweaks.RestartAllUIServices()

	if len(warnings) > 0 {
		return fmt.Errorf("applied with notices: %s", strings.Join(warnings, "; "))
	}
	return nil
}
