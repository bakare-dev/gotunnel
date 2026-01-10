package metrics

import (
	"sync"
	"time"
)

type Metrics struct {
	mu sync.RWMutex

	TotalConnections int64
	ActiveStreams    int
	TotalStreams     int64

	BytesSent     int64
	BytesReceived int64

	HTTPRequests       int64
	HTTPRequestsByCode map[int]int64
	TotalLatency       time.Duration
	MinLatency         time.Duration
	MaxLatency         time.Duration

	SessionStart time.Time
}

func New() *Metrics {
	return &Metrics{
		HTTPRequestsByCode: make(map[int]int64),
		SessionStart:       time.Now(),
		MinLatency:         time.Duration(1<<63 - 1),
	}
}

func (m *Metrics) StreamOpened() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ActiveStreams++
	m.TotalStreams++
	m.TotalConnections++
}

func (m *Metrics) StreamClosed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ActiveStreams--
	if m.ActiveStreams < 0 {
		m.ActiveStreams = 0
	}
}

func (m *Metrics) AddBytesSent(n int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.BytesSent += n
}

func (m *Metrics) AddBytesReceived(n int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.BytesReceived += n
}

func (m *Metrics) RecordHTTPRequest(statusCode int, latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.HTTPRequests++
	m.HTTPRequestsByCode[statusCode]++
	m.TotalLatency += latency

	if latency < m.MinLatency {
		m.MinLatency = latency
	}
	if latency > m.MaxLatency {
		m.MaxLatency = latency
	}
}

func (m *Metrics) GetActiveStreams() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.ActiveStreams
}

func (m *Metrics) GetTotalStreams() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.TotalStreams
}

func (m *Metrics) GetBandwidth() (sent, received int64) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.BytesSent, m.BytesReceived
}

func (m *Metrics) GetHTTPStats() (total int64, avg time.Duration, min time.Duration, max time.Duration) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.HTTPRequests == 0 {
		return 0, 0, 0, 0
	}

	avgLatency := m.TotalLatency / time.Duration(m.HTTPRequests)
	return m.HTTPRequests, avgLatency, m.MinLatency, m.MaxLatency
}

func (m *Metrics) GetUptime() time.Duration {
	return time.Since(m.SessionStart)
}

func (m *Metrics) GetStatusCodeCounts() map[int]int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	counts := make(map[int]int64)
	for code, count := range m.HTTPRequestsByCode {
		counts[code] = count
	}
	return counts
}
