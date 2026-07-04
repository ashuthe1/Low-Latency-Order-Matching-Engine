package metrics

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/HdrHistogram/hdrhistogram-go"
)

// GlobalMetrics holds the lock-free counters and the latency histogram.
type GlobalMetrics struct {
	OrdersReceived   atomic.Uint64
	OrdersMatched    atomic.Uint64
	OrdersCancelled  atomic.Uint64
	TradesExecuted   atomic.Uint64
	
	StartTime        time.Time
	
	mu               sync.Mutex
	latencyHistogram *hdrhistogram.Histogram
}

// NewMetrics initializes the metrics tracker.
func NewMetrics() *GlobalMetrics {
	return &GlobalMetrics{
		StartTime: time.Now(),
		// Track latencies from 1 microsecond to 10 seconds, with 3 significant figures
		latencyHistogram: hdrhistogram.New(1, 10000000, 3), 
	}
}

// RecordLatency adds a new duration to the histogram safely.
func (m *GlobalMetrics) RecordLatency(d time.Duration) {
	micros := d.Microseconds()
	if micros == 0 {
		micros = 1 // Prevent 0-value insertion errors
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	m.latencyHistogram.RecordValue(micros)
}

// Snapshot latencies calculates the current percentiles in milliseconds.
func (m *GlobalMetrics) LatencySnapshot() (p50, p99, p999 float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Convert microseconds back to milliseconds for the API response
	p50 = float64(m.latencyHistogram.ValueAtQuantile(50)) / 1000.0
	p99 = float64(m.latencyHistogram.ValueAtQuantile(99)) / 1000.0
	p999 = float64(m.latencyHistogram.ValueAtQuantile(99.9)) / 1000.0
	
	return
}