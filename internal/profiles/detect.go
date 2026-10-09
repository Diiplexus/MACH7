package profiles

import (
	"mach7/internal/tweaks"
)

// DetectActiveProfile identifies which profile currently matches system state
func DetectActiveProfile(reg *tweaks.Registry) string {
	activeMap := make(map[string]bool)
	activeCount := 0
	for _, tw := range reg.Tweaks {
		if tw.CheckStatusFunc() {
			activeMap[tw.ID] = true
			activeCount++
		}
	}

	if activeCount <= 1 {
		return "stock_defaults"
	}

	// Check Maximum FPS (most inclusive)
	maxProf := GetBuiltinProfiles()[1] // max_performance
	allMaxMatch := true
	for _, id := range maxProf.TweakIDs {
		if !activeMap[id] {
			allMaxMatch = false
			break
		}
	}
	if allMaxMatch {
		return "max_performance"
	}

	// Check Developer Boost
	devProf := GetBuiltinProfiles()[2] // developer_boost
	allDevMatch := true
	for _, id := range devProf.TweakIDs {
		if !activeMap[id] {
			allDevMatch = false
			break
		}
	}
	if allDevMatch && !activeMap["reduce_transparency"] {
		return "developer_boost"
	}

	// Check Ventura Feel
	venProf := GetBuiltinProfiles()[0] // ventura_feel
	allVenMatch := true
	for _, id := range venProf.TweakIDs {
		if !activeMap[id] {
			allVenMatch = false
			break
		}
	}
	if allVenMatch {
		return "ventura_feel"
	}

	return "" // Custom mix of tweaks
}
