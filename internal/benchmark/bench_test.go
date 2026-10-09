package benchmark

import (
	"testing"
)

func TestCPUBenchmark(t *testing.T) {
	ops, score := runCPUBenchmark()
	if ops <= 0 {
		t.Errorf("Expected positive CPU ops/sec, got %f", ops)
	}
	if score <= 0 {
		t.Errorf("Expected positive CPU score, got %d", score)
	}
}

func TestMemBenchmark(t *testing.T) {
	gbps, score := runMemBenchmark()
	if gbps <= 0 {
		t.Errorf("Expected positive memory throughput, got %f", gbps)
	}
	if score <= 0 {
		t.Errorf("Expected positive memory score, got %d", score)
	}
}

func TestUILatencyBenchmark(t *testing.T) {
	delayMs, micros, score := runUILatencyBenchmark()
	if delayMs < 0 {
		t.Errorf("Expected non-negative delay, got %f", delayMs)
	}
	if micros <= 0 {
		t.Errorf("Expected positive micros latency, got %f", micros)
	}
	if score <= 0 {
		t.Errorf("Expected positive score, got %d", score)
	}
}

func TestCompareResults(t *testing.T) {
	base := Result{
		CPUScore:         5000,
		CPUOpsPerSec:     1500000,
		MemScore:         3000,
		MemThroughputGBs: 4.5,
		AnimationDelayMs: 1900,
		OverallScore:     5000,
	}

	curr := Result{
		CPUScore:         6000,
		CPUOpsPerSec:     1800000,
		MemScore:         3300,
		MemThroughputGBs: 5.0,
		AnimationDelayMs: 250,
		OverallScore:     7500,
	}

	comp := CompareResults(base, curr)
	if comp.CPUDeltaPct != 20.0 {
		t.Errorf("Expected CPU delta +20%%, got %f", comp.CPUDeltaPct)
	}
	if comp.OverallDeltaPct != 50.0 {
		t.Errorf("Expected overall delta +50%%, got %f", comp.OverallDeltaPct)
	}
	if comp.UIDelayDeltaPct >= 0 {
		t.Errorf("Expected negative delay delta (faster), got %f", comp.UIDelayDeltaPct)
	}
}
