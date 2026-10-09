# ⚡ MACH7 — macOS 26 Intel Core i7 Performance Suite

> **Regain macOS Ventura / Sonoma responsiveness on macOS 26 without sacrificing modern features.**  
> Crafted specifically for Intel MacBook Pros (Core i7 + Iris Plus / Radeon) using **Go** and **Bubble Tea**.

---

## 🧐 Why macOS 26 Feels Sluggish on Intel Core i7

On modern macOS versions (macOS 15 through macOS 26), the operating system is architected primarily around Apple Silicon's Unified Memory Architecture and Neural Engine. On Intel MacBooks (such as the 2020 MacBook Pro 13" i7-1068NG7):

1. **The Shared 28W Thermal Trap**: The CPU and Iris Plus GPU share a single 28W TDP package. Heavy Metal compositing (desktop blurs, Stage Manager transitions, fluid animations) maxes out the GPU, causing the CPU to thermal-throttle down to 1.4–1.8 GHz.
2. **Retina 3360×2100 Framebuffer Load**: Driving a Retina panel with multi-pass blurs consumes substantial GPU shader bandwidth and dynamic VRAM.
3. **Background AI/ML Fallback**: Daemons like `mediaanalysisd`, `suggestd`, and `triald` lack an Apple Neural Engine, falling back to CPU AVX2 instructions that silently cook the processor.
4. **Artificial Animation Durations**: Hardcoded 200–300ms bezier curves on window resize, Dock sliding, and QuickLook previews introduce perceived micro-stutters.

**MACH7 fixes these bottlenecks directly while keeping Stage Manager, Widgets, Continuity, and all modern features 100% operational.**

---

## 🚀 Key Features

* **🖥️ Interactive Bubble Tea TUI**: Sleek, keyboard-driven terminal interface with live telemetry.
* **🌡️ Real-Time System HUD**: Live CPU speed limits (`pmset -g therm`), thermal throttle warnings, RAM pressure, and NVMe swap monitor.
* **🌊 One-Click Profiles**:
  * **Ventura Smoothness (Balanced)**: Eliminates UI delay curves, instant window resizing, instant QuickLook, snappy Dock/Mission Control. Retains full transparency and aesthetics.
  * **Maximum FPS & Thermal Headroom**: Reduces multi-pass GPU transparency, removes app launch bounces, maximizes Iris Plus GPU headroom.
  * **Developer & Compiler Boost**: Guards build directories (`~/.cache`, `~/go/pkg`, `~/.npm`, `~/.cargo`) from Spotlight indexing, purges RAM caches.
  * **Stock Defaults**: Instantly resets everything back to factory Apple settings.
* **📊 Built-in Benchmarks & Delta Testing**: Quantifiable Before vs After performance suite measuring multi-thread CPU throughput, memory bandwidth, and perceived UI animation latency.
* **🛡️ Daemon Warden**: Audits runaway background ML daemons running AVX2 instructions with one-touch Freeze (`SIGSTOP`) and Resume (`SIGCONT`).
* **💾 Zero-Risk Snapshots & Rollback**: Automatically creates JSON backups before modifying any setting so you can revert anytime with `[Enter]`.

---

## ⌨️ TUI Navigation & Neovim Motions

Launch the interactive UI:
```bash
./mach7
```

### Motion Controls
| Key | Action |
| :--- | :--- |
| `[h]` / `[l]` or `[Tab]` | Switch tabs Left / Right (Dashboard, Profiles, Tweaks, Warden, Snapshots, Benchmarks) |
| `[j]` / `[k]` or `[↑]` / `[↓]` | Move down / up through list items |
| `[g]` / `[G]` | Jump directly to Top / Bottom of list |
| `[Ctrl+d]` / `[Ctrl+u]` | Scroll half-page down / up |
| `[1]` – `[6]` | Jump directly to tab 1–6 |

### Actions
| Key | Action |
| :--- | :--- |
| `[Enter]` | Apply selected Profile, Restore Snapshot, or Run Benchmark |
| `[Space]` | Toggle individual tweak On / Off (in Tweaks tab) |
| `[B]` | Run live performance benchmark from any screen |
| `[S]` | Save Current benchmark as "Before" Baseline (in Benchmarks tab) / Save Snapshot (in Snapshots tab) |
| `[C]` | Clear saved benchmark baseline |
| `[P]` | Purge inactive memory cache (in HUD tab) / Freeze daemon (in Warden tab) |
| `[R]` | Restart Dock & Finder services / Resume daemon (in Warden tab) |
| `[F]` | Restart daemon to free leaked memory (in Warden tab) |
| `[A]` | Freeze all non-critical background AI daemons (in Warden tab) |
| `[X]` | Factory reset all tweaks back to stock defaults |
| `[Q]` / `[Ctrl+C]` | Exit MACH7 |

---

## ⚡ Headless CLI Mode

Use `mach7` directly in scripts, crons, or terminal shortcuts without launching the TUI:

```bash
# Print live telemetry, real-time CPU %, disk I/O and active profile
./mach7 --status

# Run the performance benchmark suite and display scores
./mach7 --bench

# Run benchmark and save as 'Before' Baseline
./mach7 --bench --save-baseline

# Apply the balanced "Ventura Smoothness" profile
./mach7 --profile ventura

# Apply the ultra-responsive "Maximum FPS" profile
./mach7 --profile max

# Apply the Developer profile
./mach7 --profile dev

# Purge inactive RAM memory caches
./mach7 --purge

# Revert all changes back to macOS stock defaults
./mach7 --restore
```

---

## 🛠️ Install from Source

Requires **Go 1.25+** and macOS. From the repository root, install the `mach7` executable into your Go binary directory:

```bash
go install ./cmd/mach7
```

Make sure `$(go env GOPATH)/bin` (or the directory configured by `GOBIN`) is on your `PATH`, then run:

```bash
mach7
```

To build a standalone executable in the current directory instead:

```bash
go build -o mach7 ./cmd/mach7
./mach7
```

## 🍺 Homebrew

Pushing a version tag (for example, `v0.1.0`) triggers a GitHub Actions release for Intel and Apple Silicon and updates the Homebrew formula. Because the formula lives in this repository (not in a `homebrew-*` repository), tap it with its explicit URL, then install:

```bash
brew tap diiplexus/mach7 https://github.com/Diiplexus/MACH7.git
brew install diiplexus/mach7/mach7
```

Tagged releases build macOS binaries for Intel (`amd64`) and Apple Silicon (`arm64`) and update the Homebrew formula in `Formula/`.
