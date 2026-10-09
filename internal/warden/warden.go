package warden

import (
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// DaemonTarget defines a tracked background daemon
type DaemonTarget struct {
	Name        string
	Binary      string
	Description string
	Impact      string
}

// DaemonStatus represents live status of a tracked daemon
type DaemonStatus struct {
	Target    DaemonTarget
	IsRunning bool
	PIDs      []int
	IsPaused  bool
}

// TrackedTargets lists known CPU/thermal heavy daemons on Intel Macs
var TrackedTargets = []DaemonTarget{
	{
		Name:        "Media Analysis Engine",
		Binary:      "mediaanalysisd",
		Description: "Analyzes photos and videos for faces, objects, and visual search. AVX2 heavy on Intel.",
		Impact:      "High CPU & Heat",
	},
	{
		Name:        "Photo Library Daemon",
		Binary:      "photolibraryd",
		Description: "Processes photo thumbnails, syncs metadata, and calculates photo curation scores.",
		Impact:      "High Disk & RAM",
	},
	{
		Name:        "Suggest & Intelligence Daemon",
		Binary:      "suggestd",
		Description: "Generates proactive suggestions for Mail, Calendar, and Spotlight using ML models.",
		Impact:      "Spike CPU Usage",
	},
	{
		Name:        "Trial Experiment Daemon",
		Binary:      "triald",
		Description: "Downloads A/B trial configurations and ML model variants from Apple servers.",
		Impact:      "Background Network/CPU",
	},
	{
		Name:        "Spotlight Content Indexer",
		Binary:      "mds_stores",
		Description: "Primary metadata indexer for local files. High disk I/O during developer builds.",
		Impact:      "Continuous Disk I/O",
	},
}

// ScanDaemons checks which daemons are currently running
func ScanDaemons() []DaemonStatus {
	var statuses []DaemonStatus

	for _, target := range TrackedTargets {
		pids := findPIDs(target.Binary)
		status := DaemonStatus{
			Target:    target,
			IsRunning: len(pids) > 0,
			PIDs:      pids,
		}
		statuses = append(statuses, status)
	}

	return statuses
}

func findPIDs(binary string) []int {
	out, err := exec.Command("pgrep", binary).Output()
	if err != nil {
		return nil
	}

	var pids []int
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, l := range lines {
		if pid, err := strconv.Atoi(strings.TrimSpace(l)); err == nil && pid > 0 {
			pids = append(pids, pid)
		}
	}
	return pids
}

// PauseDaemon sends SIGSTOP to freeze daemon execution temporarily
func PauseDaemon(status DaemonStatus) error {
	for _, pid := range status.PIDs {
		_ = syscall.Kill(pid, syscall.SIGSTOP)
	}
	return nil
}

// ResumeDaemon sends SIGCONT to unfreeze daemon execution
func ResumeDaemon(status DaemonStatus) error {
	for _, pid := range status.PIDs {
		_ = syscall.Kill(pid, syscall.SIGCONT)
	}
	return nil
}

// RestartDaemon sends SIGTERM to restart daemon and release accumulated memory
func RestartDaemon(status DaemonStatus) error {
	for _, pid := range status.PIDs {
		_ = syscall.Kill(pid, syscall.SIGTERM)
	}
	return nil
}
