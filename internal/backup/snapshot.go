package backup

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// KeyDef defines a domain and key pair to track
type KeyDef struct {
	Domain string `json:"domain"`
	Key    string `json:"key"`
}

// KeyBackup stores the original state of a key
type KeyBackup struct {
	Domain        string `json:"domain"`
	Key           string `json:"key"`
	ExistedBefore bool   `json:"existed_before"`
	ValueType     string `json:"value_type"` // "bool", "float", "int", "string"
	RawValue      string `json:"raw_value"`
}

// Snapshot represents a full backup snapshot of system settings
type Snapshot struct {
	ID        string      `json:"id"`
	CreatedAt time.Time   `json:"created_at"`
	Reason    string      `json:"reason"`
	Entries   []KeyBackup `json:"entries"`
}

func getSnapshotDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".mach7", "snapshots")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

// ReadCurrentKey inspects `defaults read <domain> <key>`
func ReadCurrentKey(domain, key string) (existed bool, rawVal string, valType string) {
	var cmd *exec.Cmd
	if domain == "-g" || domain == "NSGlobalDomain" {
		cmd = exec.Command("defaults", "read", "-g", key)
	} else {
		cmd = exec.Command("defaults", "read", domain, key)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, "", ""
	}

	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return false, "", ""
	}

	// Determine type
	if raw == "1" || raw == "0" || strings.EqualFold(raw, "true") || strings.EqualFold(raw, "false") {
		return true, raw, "bool"
	}
	if strings.Contains(raw, ".") {
		return true, raw, "float"
	}
	return true, raw, "string"
}

// CreateSnapshot takes a snapshot of a list of keys and saves to ~/.mach7/snapshots/
func CreateSnapshot(reason string, keys []KeyDef) (*Snapshot, error) {
	dir, err := getSnapshotDir()
	if err != nil {
		return nil, err
	}

	snapID := fmt.Sprintf("snap_%s", time.Now().Format("20060102_150405"))
	snap := &Snapshot{
		ID:        snapID,
		CreatedAt: time.Now(),
		Reason:    reason,
		Entries:   make([]KeyBackup, 0, len(keys)),
	}

	for _, k := range keys {
		existed, raw, vType := ReadCurrentKey(k.Domain, k.Key)
		snap.Entries = append(snap.Entries, KeyBackup{
			Domain:        k.Domain,
			Key:           k.Key,
			ExistedBefore: existed,
			ValueType:     vType,
			RawValue:      raw,
		})
	}

	filePath := filepath.Join(dir, snapID+".json")
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return nil, err
	}

	return snap, nil
}

// ListSnapshots returns all stored snapshots ordered newest first
func ListSnapshots() ([]Snapshot, error) {
	dir, err := getSnapshotDir()
	if err != nil {
		return nil, err
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var list []Snapshot
	for i := len(files) - 1; i >= 0; i-- {
		f := files[i]
		if !strings.HasSuffix(f.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, f.Name()))
		if err != nil {
			continue
		}
		var s Snapshot
		if err := json.Unmarshal(data, &s); err == nil {
			list = append(list, s)
		}
	}

	return list, nil
}

// RestoreSnapshot restores values from a snapshot
func RestoreSnapshot(snap *Snapshot) error {
	var errs []string

	for _, entry := range snap.Entries {
		if !entry.ExistedBefore {
			// Key didn't exist before, delete it
			var cmd *exec.Cmd
			if entry.Domain == "-g" || entry.Domain == "NSGlobalDomain" {
				cmd = exec.Command("defaults", "delete", "-g", entry.Key)
			} else {
				cmd = exec.Command("defaults", "delete", entry.Domain, entry.Key)
			}
			_ = cmd.Run() // Ignore error if already deleted
		} else {
			// Key existed, restore it
			typeFlag := "-string"
			switch entry.ValueType {
			case "bool":
				typeFlag = "-bool"
			case "float":
				typeFlag = "-float"
			case "int":
				typeFlag = "-int"
			}

			var cmd *exec.Cmd
			if entry.Domain == "-g" || entry.Domain == "NSGlobalDomain" {
				cmd = exec.Command("defaults", "write", "-g", entry.Key, typeFlag, entry.RawValue)
			} else {
				cmd = exec.Command("defaults", "write", entry.Domain, entry.Key, typeFlag, entry.RawValue)
			}
			if err := cmd.Run(); err != nil {
				errs = append(errs, fmt.Sprintf("Failed to restore %s %s: %v", entry.Domain, entry.Key, err))
			}
		}
	}

	// Restart UI services so reverted defaults take effect
	_ = exec.Command("killall", "Dock").Run()
	_ = exec.Command("killall", "Finder").Run()

	if len(errs) > 0 {
		return fmt.Errorf("errors during restore: %s", strings.Join(errs, "; "))
	}
	return nil
}
