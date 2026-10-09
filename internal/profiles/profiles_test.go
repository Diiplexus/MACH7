package profiles

import (
	"testing"

	"mach7/internal/tweaks"
)

func TestGetBuiltinProfiles(t *testing.T) {
	profs := GetBuiltinProfiles()
	if len(profs) < 4 {
		t.Fatalf("Expected at least 4 profiles, got %d", len(profs))
	}

	names := map[string]bool{}
	for _, p := range profs {
		if p.ID == "" || p.Name == "" {
			t.Errorf("Profile missing ID or Name: %+v", p)
		}
		names[p.ID] = true
	}

	if !names["ventura_feel"] || !names["max_performance"] || !names["stock_defaults"] {
		t.Errorf("Missing essential profiles in registry: %+v", names)
	}
}

func TestDetectActiveProfile(t *testing.T) {
	reg := tweaks.NewRegistry()
	detected := DetectActiveProfile(reg)
	// On this system, since tweaks are active, it should return a non-empty string or custom mix
	t.Logf("Currently detected active profile: '%s'", detected)
}
