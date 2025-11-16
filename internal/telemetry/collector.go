package telemetry

import (
	"sync"
	"time"
)

// TelemetryCollector tracks simple network metrics used by the AI optimizer.
// It is intentionally lightweight: a single instance per sender process.
type TelemetryCollector struct {
	mu sync.RWMutex

	windowStart time.Time
	bytesSent   uint64
	lastRTT     time.Duration
}

// NewTelemetryCollector creates a new collector with an initialized time window.
func NewTelemetryCollector() *TelemetryCollector {
	return &TelemetryCollector{
		windowStart: time.Now(),
	}
}

// RecordBytesSent records that n bytes have been sent.
func (t *TelemetryCollector) RecordBytesSent(n int) {
	if n <= 0 {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.bytesSent += uint64(n)
}

// RecordRTT records the latest round-trip time measurement.
func (t *TelemetryCollector) RecordRTT(d time.Duration) {
	if d <= 0 {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.lastRTT = d
}

// BandwidthMbps returns a very simple estimate of bandwidth in megabits per second
// based on bytes sent in the current window divided by elapsed time.
// If not enough data is available, it returns 0.
func (t *TelemetryCollector) BandwidthMbps() float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()

	elapsed := time.Since(t.windowStart).Seconds()
	if elapsed <= 0 || t.bytesSent == 0 {
		return 0
	}

	// bits per second -> megabits per second
	bps := float64(t.bytesSent*8) / elapsed
	return bps / 1e6
}

// LatencyMs returns the last recorded RTT in milliseconds.
// If no RTT has been recorded yet, it returns 0.
func (t *TelemetryCollector) LatencyMs() float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.lastRTT <= 0 {
		return 0
	}
	return float64(t.lastRTT.Milliseconds())
}


