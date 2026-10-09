# MACH7 Copilot Instructions

## Commands

- Build the TUI/CLI binary: `go build -o mach7 ./cmd/mach7`
- Run all tests: `go test ./...`
- Run one test: `go test ./internal/benchmark -run '^TestCompareResults$'`
- Run static analysis: `go vet ./...`
- Format changed Go files: `gofmt -w <files>`

The module requires Go 1.25. The benchmark tests execute local CPU, memory, and timing work rather than mocked measurements, so they are host-dependent and can take noticeably longer than ordinary unit tests.

## Architecture

- `cmd/mach7/main.go` is the sole entry point. It dispatches headless flags (`--status`, `--bench`, `--profile`, `--purge`, `--restore`) before starting the Bubble Tea program in `internal/ui`.
- `internal/ui` owns the Bubble Tea `Model`: keyboard handling, six-tab rendering, periodic telemetry refreshes, asynchronous benchmarks, and user-facing toast messages. UI actions call domain packages directly; there is no service layer.
- `internal/tweaks` is the source of truth for optimizations. `NewRegistry()` builds `Tweak` records with status, apply, and revert closures around macOS `defaults`, `purge`, and Dock/Finder restart commands.
- `internal/profiles` composes tweak IDs into named presets and detects the active preset by querying the registry. Profile application enables included tweaks but does not automatically disable tweaks absent from the selected non-stock profile.
- `internal/backup` snapshots the original `defaults` values represented by a tweak's `Keys` before UI-initiated mutations. Snapshots are JSON files in `~/.mach7/snapshots`; benchmark baselines are stored separately in `~/.mach7`.
- `internal/system` parses macOS command output (`sysctl`, `pmset`, `vm_stat`, `iostat`) into telemetry structs. `internal/warden` tracks selected daemons and controls their PIDs with POSIX signals.

## Repository Conventions

- Keep a tweak's `ID` stable and use it consistently in `profiles.GetBuiltinProfiles()` and profile detection. Add every `defaults` key it changes to `Tweak.Keys` so snapshots can restore the pre-change values; file-based tweaks intentionally have no keys.
- Make `CheckStatusFunc`, `ApplyFunc`, and `RevertFunc` agree on one concrete macOS state. Apply/revert operations should be idempotent because profiles skip tweaks already reported active and the UI lets users toggle repeatedly.
- When a tweak affects Dock or Finder preferences, restart the relevant process after mutation. Profile-wide and stock reset operations call `RestartAllUIServices()` after their per-tweak work.
- Preserve best-effort behavior for environment-dependent macOS operations: the UI reports failures via its toast state, profile application aggregates per-tweak failures as notices, and command-output parsers retain usable default values when a system command is unavailable.
- UI mutations create snapshots before applying profiles or keyed tweaks. Headless CLI mutations currently do not create snapshots, so do not claim that command path has automatic rollback.
- The project targets macOS and invokes native utilities directly. Keep platform-specific behavior in the owning `internal/system`, `internal/tweaks`, `internal/backup`, or `internal/warden` package rather than duplicating subprocess logic in UI views.
