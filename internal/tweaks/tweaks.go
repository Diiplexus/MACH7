package tweaks

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"mach7/internal/backup"
)

// Category groups tweaks logically
type Category string

const (
	CatCompositor Category = "Compositor & Graphics"
	CatDock       Category = "Dock & Mission Control"
	CatFinder     Category = "Finder & Dialogs"
	CatSystem     Category = "System & Indexing"
)

// Tweak defines a single optimization toggle
type Tweak struct {
	ID          string
	Title       string
	Description string
	Category    Category
	Impact      string // e.g., "High GPU Relief", "Zero Latency"
	Keys        []backup.KeyDef

	CheckStatusFunc func() bool
	ApplyFunc       func() error
	RevertFunc      func() error
}

// Registry holds all available optimizations
type Registry struct {
	Tweaks []Tweak
}

// NewRegistry initializes and returns all built-in tweaks
func NewRegistry() *Registry {
	r := &Registry{}
	r.registerAll()
	return r
}

func (r *Registry) registerAll() {
	r.Tweaks = []Tweak{
		// 1. Window Resize Time (Instant)
		{
			ID:          "window_resize_time",
			Title:       "Instant Window Resize Animation",
			Description: "Sets NSWindowResizeTime to 0.001s, eliminating sluggish modal sheet and window resize delays.",
			Category:    CatCompositor,
			Impact:      "Zero Latency",
			Keys: []backup.KeyDef{
				{Domain: "-g", Key: "NSWindowResizeTime"},
			},
			CheckStatusFunc: func() bool {
				existed, val, _ := backup.ReadCurrentKey("-g", "NSWindowResizeTime")
				if !existed {
					return false
				}
				f, err := strconv.ParseFloat(val, 64)
				return err == nil && f <= 0.01
			},
			ApplyFunc: func() error {
				return runDefaultsWrite("-g", "NSWindowResizeTime", "-float", "0.001")
			},
			RevertFunc: func() error {
				return runDefaultsDelete("-g", "NSWindowResizeTime")
			},
		},

		// 2. Reduce Transparency (Iris Plus GPU Lifesaver)
		{
			ID:          "reduce_transparency",
			Title:       "Reduce UI Transparency Passes",
			Description: "Removes heavy multi-pass blur on titlebars, Dock, and menu bar. Huge relief for Intel Iris Plus GPU.",
			Category:    CatCompositor,
			Impact:      "High GPU Relief",
			Keys: []backup.KeyDef{
				{Domain: "com.apple.universalaccess", Key: "reduceTransparency"},
			},
			CheckStatusFunc: func() bool {
				existed, val, _ := backup.ReadCurrentKey("com.apple.universalaccess", "reduceTransparency")
				return existed && (val == "1" || strings.EqualFold(val, "true"))
			},
			ApplyFunc: func() error {
				existed, val, _ := backup.ReadCurrentKey("com.apple.universalaccess", "reduceTransparency")
				if existed && (val == "1" || strings.EqualFold(val, "true")) {
					return nil // Already active
				}
				if err := runDefaultsWrite("com.apple.universalaccess", "reduceTransparency", "-bool", "true"); err != nil {
					return fmt.Errorf("macOS protects Accessibility: toggle via System Settings > Accessibility > Display > Reduce transparency")
				}
				return nil
			},
			RevertFunc: func() error {
				existed, val, _ := backup.ReadCurrentKey("com.apple.universalaccess", "reduceTransparency")
				if !existed || val == "0" || strings.EqualFold(val, "false") {
					return nil // Already disabled
				}
				if err := runDefaultsWrite("com.apple.universalaccess", "reduceTransparency", "-bool", "false"); err != nil {
					return fmt.Errorf("macOS protects Accessibility: toggle via System Settings > Accessibility > Display > Reduce transparency")
				}
				return nil
			},
		},

		// 3. Quick Look Animation Duration
		{
			ID:          "quicklook_speed",
			Title:       "Instant Quick Look Preview",
			Description: "Removes zoom/fade animation delay when pressing Spacebar to preview files in Finder.",
			Category:    CatFinder,
			Impact:      "Instant Response",
			Keys: []backup.KeyDef{
				{Domain: "-g", Key: "QLPanelAnimationDuration"},
			},
			CheckStatusFunc: func() bool {
				existed, val, _ := backup.ReadCurrentKey("-g", "QLPanelAnimationDuration")
				if !existed {
					return false
				}
				f, err := strconv.ParseFloat(val, 64)
				return err == nil && f <= 0.01
			},
			ApplyFunc: func() error {
				return runDefaultsWrite("-g", "QLPanelAnimationDuration", "-float", "0")
			},
			RevertFunc: func() error {
				return runDefaultsDelete("-g", "QLPanelAnimationDuration")
			},
		},

		// 4. Dock Auto-Hide Delay to Zero
		{
			ID:          "dock_autohide_delay",
			Title:       "Instant Dock Auto-Hide Response",
			Description: "Removes the 0.5s artificial hover delay before the Dock begins sliding into view.",
			Category:    CatDock,
			Impact:      "Zero Latency",
			Keys: []backup.KeyDef{
				{Domain: "com.apple.dock", Key: "autohide-delay"},
			},
			CheckStatusFunc: func() bool {
				existed, val, _ := backup.ReadCurrentKey("com.apple.dock", "autohide-delay")
				if !existed {
					return false
				}
				f, err := strconv.ParseFloat(val, 64)
				return err == nil && f == 0
			},
			ApplyFunc: func() error {
				if err := runDefaultsWrite("com.apple.dock", "autohide-delay", "-float", "0"); err != nil {
					return err
				}
				return restartDock()
			},
			RevertFunc: func() error {
				_ = runDefaultsDelete("com.apple.dock", "autohide-delay")
				return restartDock()
			},
		},

		// 5. Dock Animation Speedup (0.15s)
		{
			ID:          "dock_autohide_speed",
			Title:       "Snappy Dock Slide Animation",
			Description: "Accelerates Dock slide animation from 0.8s down to 0.15s for crisp, fast navigation.",
			Category:    CatDock,
			Impact:      "High FPS Feel",
			Keys: []backup.KeyDef{
				{Domain: "com.apple.dock", Key: "autohide-time-modifier"},
			},
			CheckStatusFunc: func() bool {
				existed, val, _ := backup.ReadCurrentKey("com.apple.dock", "autohide-time-modifier")
				if !existed {
					return false
				}
				f, err := strconv.ParseFloat(val, 64)
				return err == nil && f <= 0.2
			},
			ApplyFunc: func() error {
				if err := runDefaultsWrite("com.apple.dock", "autohide-time-modifier", "-float", "0.15"); err != nil {
					return err
				}
				return restartDock()
			},
			RevertFunc: func() error {
				_ = runDefaultsDelete("com.apple.dock", "autohide-time-modifier")
				return restartDock()
			},
		},

		// 6. Mission Control Animation Duration
		{
			ID:          "mission_control_speed",
			Title:       "Rapid Mission Control / Exposé",
			Description: "Reduces expose-animation-duration to 0.1s for immediate window overview without frame lag.",
			Category:    CatDock,
			Impact:      "Zero Stutter",
			Keys: []backup.KeyDef{
				{Domain: "com.apple.dock", Key: "expose-animation-duration"},
			},
			CheckStatusFunc: func() bool {
				existed, val, _ := backup.ReadCurrentKey("com.apple.dock", "expose-animation-duration")
				if !existed {
					return false
				}
				f, err := strconv.ParseFloat(val, 64)
				return err == nil && f <= 0.15
			},
			ApplyFunc: func() error {
				if err := runDefaultsWrite("com.apple.dock", "expose-animation-duration", "-float", "0.1"); err != nil {
					return err
				}
				return restartDock()
			},
			RevertFunc: func() error {
				_ = runDefaultsDelete("com.apple.dock", "expose-animation-duration")
				return restartDock()
			},
		},

		// 7. Disable Dock App Launch Bounce
		{
			ID:          "dock_launch_anim",
			Title:       "Disable App Launch Icon Bouncing",
			Description: "Stops icons from continuously bouncing in Dock during app launch, saving CPU & GPU draw calls.",
			Category:    CatDock,
			Impact:      "Lower CPU Spikes",
			Keys: []backup.KeyDef{
				{Domain: "com.apple.dock", Key: "launchanim"},
			},
			CheckStatusFunc: func() bool {
				existed, val, _ := backup.ReadCurrentKey("com.apple.dock", "launchanim")
				return existed && (val == "0" || strings.EqualFold(val, "false"))
			},
			ApplyFunc: func() error {
				if err := runDefaultsWrite("com.apple.dock", "launchanim", "-bool", "false"); err != nil {
					return err
				}
				return restartDock()
			},
			RevertFunc: func() error {
				_ = runDefaultsDelete("com.apple.dock", "launchanim")
				return restartDock()
			},
		},

		// 8. Disable Finder Window Zoom Animations
		{
			ID:          "finder_animations",
			Title:       "Disable Finder Window Zoom Transitions",
			Description: "Opens Finder windows and info panels immediately without slow scaling animations.",
			Category:    CatFinder,
			Impact:      "Instant UI",
			Keys: []backup.KeyDef{
				{Domain: "com.apple.finder", Key: "DisableAllAnimations"},
			},
			CheckStatusFunc: func() bool {
				existed, val, _ := backup.ReadCurrentKey("com.apple.finder", "DisableAllAnimations")
				return existed && (val == "1" || strings.EqualFold(val, "true"))
			},
			ApplyFunc: func() error {
				if err := runDefaultsWrite("com.apple.finder", "DisableAllAnimations", "-bool", "true"); err != nil {
					return err
				}
				return restartFinder()
			},
			RevertFunc: func() error {
				_ = runDefaultsDelete("com.apple.finder", "DisableAllAnimations")
				return restartFinder()
			},
		},

		// 9. Auto-Expand Save & Print Panels
		{
			ID:          "dialog_expand",
			Title:       "Always Expanded Save/Print Dialogs",
			Description: "Eliminates the animation and recalculation of expanding save/print dialogs each time.",
			Category:    CatFinder,
			Impact:      "Smoother Workflow",
			Keys: []backup.KeyDef{
				{Domain: "-g", Key: "NSNavPanelExpandedStateForSaveMode"},
				{Domain: "-g", Key: "NSNavPanelExpandedStateForSaveMode2"},
			},
			CheckStatusFunc: func() bool {
				existed, val, _ := backup.ReadCurrentKey("-g", "NSNavPanelExpandedStateForSaveMode")
				return existed && (val == "1" || strings.EqualFold(val, "true"))
			},
			ApplyFunc: func() error {
				if err := runDefaultsWrite("-g", "NSNavPanelExpandedStateForSaveMode", "-bool", "true"); err != nil {
					return err
				}
				return runDefaultsWrite("-g", "NSNavPanelExpandedStateForSaveMode2", "-bool", "true")
			},
			RevertFunc: func() error {
				_ = runDefaultsDelete("-g", "NSNavPanelExpandedStateForSaveMode")
				_ = runDefaultsDelete("-g", "NSNavPanelExpandedStateForSaveMode2")
				return nil
			},
		},

		// 10. Developer Spotlight Ignore Rules
		{
			ID:          "spotlight_dev_protection",
			Title:       "Protect Developer Directories from Spotlight",
			Description: "Adds .metadata_never_index to ~/.cache, ~/go/pkg, ~/.npm, and ~/.cargo to prevent mds_stores CPU spikes.",
			Category:    CatSystem,
			Impact:      "Thermal & CPU Relief",
			Keys:        nil,
			CheckStatusFunc: func() bool {
				home, _ := os.UserHomeDir()
				marker := filepath.Join(home, ".cache", ".metadata_never_index")
				_, err := os.Stat(marker)
				return err == nil
			},
			ApplyFunc: func() error {
				home, err := os.UserHomeDir()
				if err != nil {
					return err
				}
				dirs := []string{
					filepath.Join(home, ".cache"),
					filepath.Join(home, "go", "pkg"),
					filepath.Join(home, ".npm"),
					filepath.Join(home, ".cargo"),
				}
				for _, d := range dirs {
					if err := os.MkdirAll(d, 0755); err == nil {
						marker := filepath.Join(d, ".metadata_never_index")
						_ = os.WriteFile(marker, []byte(""), 0644)
					}
				}
				return nil
			},
			RevertFunc: func() error {
				home, err := os.UserHomeDir()
				if err != nil {
					return err
				}
				dirs := []string{
					filepath.Join(home, ".cache"),
					filepath.Join(home, "go", "pkg"),
					filepath.Join(home, ".npm"),
					filepath.Join(home, ".cargo"),
				}
				for _, d := range dirs {
					marker := filepath.Join(d, ".metadata_never_index")
					_ = os.Remove(marker)
				}
				return nil
			},
		},
	}
}

// Helpers
func runDefaultsWrite(domain, key, vType, val string) error {
	var cmd *exec.Cmd
	if domain == "-g" || domain == "NSGlobalDomain" {
		cmd = exec.Command("defaults", "write", "-g", key, vType, val)
	} else {
		cmd = exec.Command("defaults", "write", domain, key, vType, val)
	}
	return cmd.Run()
}

func runDefaultsDelete(domain, key string) error {
	var cmd *exec.Cmd
	if domain == "-g" || domain == "NSGlobalDomain" {
		cmd = exec.Command("defaults", "delete", "-g", key)
	} else {
		cmd = exec.Command("defaults", "delete", domain, key)
	}
	return cmd.Run()
}

func restartDock() error {
	return exec.Command("killall", "Dock").Run()
}

func restartFinder() error {
	return exec.Command("killall", "Finder").Run()
}

// RestartAllUIServices restarts Dock and Finder
func RestartAllUIServices() {
	_ = restartDock()
	_ = restartFinder()
}

// PurgeInactiveMemory triggers inactive cache flush
func PurgeInactiveMemory() error {
	return exec.Command("purge").Run()
}
