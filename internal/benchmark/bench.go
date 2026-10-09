package benchmark

import (
	"crypto/sha256"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"mach7/internal/backup"
	"mach7/internal/system"
)

// Result holds the performance test results
type Result struct {
	Timestamp            time.Time `json:"timestamp"`
	CPUScore             int       `json:"cpu_score"`
	CPUOpsPerSec         float64   `json:"cpu_ops_per_sec"`
	MemScore             int       `json:"mem_score"`
	MemThroughputGBs     float64   `json:"mem_throughput_gbs"`
	UILatencyScore       int       `json:"ui_latency_score"`
	AnimationDelayMs     float64   `json:"animation_delay_ms"`
	DefaultsLatencyMicros float64   `json:"defaults_latency_micros"`
	OverallScore         int       `json:"overall_score"`
	IsThrottled          bool      `json:"is_throttled"`
	ActiveTweaksCount    int       `json:"active_tweaks_count"`
}

// Comparison shows the delta between baseline and current
type Comparison struct {
	Baseline         Result
	Current          Result
	CPUDeltaPct      float64
	MemDeltaPct      float64
	UIDelayDeltaPct  float64 // Negative is better (less delay)
	OverallDeltaPct  float64
}

// RunBenchmark executes the suite of performance tests
func RunBenchmark(activeTweaks int) Result {
	res := Result{
		Timestamp:         time.Now(),
		ActiveTweaksCount: activeTweaks,
	}

	// 1. CPU Concurrency & Crunch Test (1.5 seconds)
	res.CPUOpsPerSec, res.CPUScore = runCPUBenchmark()

	// 2. Memory Throughput Test (1 second)
	res.MemThroughputGBs, res.MemScore = runMemBenchmark()

	// 3. UI Latency & Animation Delay Index
	res.AnimationDelayMs, res.DefaultsLatencyMicros, res.UILatencyScore = runUILatencyBenchmark()

	// 4. Check if thermal throttling occurred
	therm := system.GetThermalStatus()
	res.IsThrottled = therm.IsThrottled

	// 5. Aggregate overall score
	res.OverallScore = int(float64(res.CPUScore)*0.45 + float64(res.MemScore)*0.25 + float64(res.UILatencyScore)*0.30)

	return res
}

func runCPUBenchmark() (float64, int) {
	numWorkers := runtime.NumCPU()
	duration := 1500 * time.Millisecond
	stopCh := make(chan struct{})
	var totalOps uint64

	var wg sync.WaitGroup
	start := time.Now()

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			var localOps uint64
			buf := make([]byte, 512)
			for j := range buf {
				buf[j] = byte((workerID*31 + j) & 0xFF)
			}

			for {
				select {
				case <-stopCh:
					atomic.AddUint64(&totalOps, localOps)
					return
				default:
					h := sha256.Sum256(buf)
					buf[0] = h[0]
					localOps++
				}
			}
		}(i)
	}

	time.Sleep(duration)
	close(stopCh)
	wg.Wait()
	elapsed := time.Since(start).Seconds()

	opsPerSec := float64(totalOps) / elapsed
	// Standardize to a friendly score (~5,000 baseline on 4C/8T)
	score := int(opsPerSec / 350.0)
	return opsPerSec, score
}

func runMemBenchmark() (float64, int) {
	bufferSize := 16 * 1024 * 1024 // 16 MB chunks
	buf := make([]byte, bufferSize)

	iterations := 40
	start := time.Now()

	for it := 0; it < iterations; it++ {
		step := 64
		for i := 0; i < bufferSize; i += step {
			buf[i] = byte(it & 0xFF)
		}
	}

	elapsed := time.Since(start).Seconds()
	bytesTransferred := float64(bufferSize*iterations) / (1024 * 1024 * 1024)
	gbPerSec := bytesTransferred / elapsed

	score := int(gbPerSec * 850.0)
	return gbPerSec, score
}

func runUILatencyBenchmark() (float64, float64, int) {
	// 1. Calculate cumulative animation delays from defaults
	var totalDelayMs float64

	// Window resize
	if existed, val, _ := backup.ReadCurrentKey("-g", "NSWindowResizeTime"); existed {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			totalDelayMs += f * 1000.0
		} else {
			totalDelayMs += 200.0
		}
	} else {
		totalDelayMs += 200.0 // Apple default is 0.2s
	}

	// Dock hover delay
	if existed, val, _ := backup.ReadCurrentKey("com.apple.dock", "autohide-delay"); existed {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			totalDelayMs += f * 1000.0
		} else {
			totalDelayMs += 500.0
		}
	} else {
		totalDelayMs += 500.0 // Apple default is ~0.5s
	}

	// Dock animation speed
	if existed, val, _ := backup.ReadCurrentKey("com.apple.dock", "autohide-time-modifier"); existed {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			totalDelayMs += f * 1000.0
		} else {
			totalDelayMs += 800.0
		}
	} else {
		totalDelayMs += 800.0 // Apple default is ~0.8s
	}

	// QuickLook
	if existed, val, _ := backup.ReadCurrentKey("-g", "QLPanelAnimationDuration"); existed {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			totalDelayMs += f * 1000.0
		} else {
			totalDelayMs += 200.0
		}
	} else {
		totalDelayMs += 200.0 // Apple default is ~0.2s
	}

	// Mission Control
	if existed, val, _ := backup.ReadCurrentKey("com.apple.dock", "expose-animation-duration"); existed {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			totalDelayMs += f * 1000.0
		} else {
			totalDelayMs += 200.0
		}
	} else {
		totalDelayMs += 200.0 // Apple default is ~0.2s
	}

	// 2. Measure defaults read roundtrip latency
	readStart := time.Now()
	for i := 0; i < 5; i++ {
		_, _, _ = backup.ReadCurrentKey("-g", "NSWindowResizeTime")
	}
	defaultsMicros := float64(time.Since(readStart).Microseconds()) / 5.0

	// Score is inversely proportional to animation delay
	// Stock (~1900ms delay) yields ~2,000 points
	// Fully optimized (~250ms delay) yields ~9,500 points
	baseScore := 10000.0 - (totalDelayMs * 4.2)
	if baseScore < 500 {
		baseScore = 500
	}
	uiScore := int(math.Round(baseScore))

	return totalDelayMs, defaultsMicros, uiScore
}

// CompareResults calculates the delta between baseline and current
func CompareResults(base, curr Result) Comparison {
	c := Comparison{
		Baseline: base,
		Current:  curr,
	}

	if base.CPUScore > 0 {
		c.CPUDeltaPct = ((float64(curr.CPUScore) - float64(base.CPUScore)) / float64(base.CPUScore)) * 100.0
	}
	if base.MemScore > 0 {
		c.MemDeltaPct = ((float64(curr.MemScore) - float64(base.MemScore)) / float64(base.MemScore)) * 100.0
	}
	if base.AnimationDelayMs > 0 {
		c.UIDelayDeltaPct = ((curr.AnimationDelayMs - base.AnimationDelayMs) / base.AnimationDelayMs) * 100.0
	}
	if base.OverallScore > 0 {
		c.OverallDeltaPct = ((float64(curr.OverallScore) - float64(base.OverallScore)) / float64(base.OverallScore)) * 100.0
	}

	return c
}

func getBenchDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".mach7")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

// SaveBaseline stores the benchmark as baseline
func SaveBaseline(res Result) error {
	dir, err := getBenchDir()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "baseline.json"), data, 0644)
}

// LoadBaseline loads previously saved baseline if exists
func LoadBaseline() (*Result, error) {
	dir, err := getBenchDir()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(dir, "baseline.json"))
	if err != nil {
		return nil, err
	}
	var res Result
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}
	return &res, nil
}
