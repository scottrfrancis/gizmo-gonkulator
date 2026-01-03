// Package middleware provides HTTP middleware for the MCP server.
package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// Metrics collects server metrics for Prometheus-compatible output.
type Metrics struct {
	requestsTotal      map[string]*uint64 // method -> count
	requestDurations   []float64          // sliding window of recent durations
	activeSessions     *int64
	errorsTotal        map[int]*uint64 // error code -> count
	mu                 sync.RWMutex
	startTime          time.Time
	maxDurationSamples int
}

// NewMetrics creates a new metrics collector.
func NewMetrics() *Metrics {
	return &Metrics{
		requestsTotal:      make(map[string]*uint64),
		requestDurations:   make([]float64, 0, 1000),
		activeSessions:     new(int64),
		errorsTotal:        make(map[int]*uint64),
		startTime:          time.Now(),
		maxDurationSamples: 1000,
	}
}

// RecordRequest records a request with its method and duration.
func (m *Metrics) RecordRequest(method string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Increment request counter
	if _, ok := m.requestsTotal[method]; !ok {
		var zero uint64
		m.requestsTotal[method] = &zero
	}
	atomic.AddUint64(m.requestsTotal[method], 1)

	// Record duration (sliding window)
	if len(m.requestDurations) >= m.maxDurationSamples {
		// Remove oldest sample
		m.requestDurations = m.requestDurations[1:]
	}
	m.requestDurations = append(m.requestDurations, duration.Seconds())
}

// RecordError records an error by code.
func (m *Metrics) RecordError(code int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.errorsTotal[code]; !ok {
		var zero uint64
		m.errorsTotal[code] = &zero
	}
	atomic.AddUint64(m.errorsTotal[code], 1)
}

// SetActiveSessions sets the current active session count.
func (m *Metrics) SetActiveSessions(count int64) {
	atomic.StoreInt64(m.activeSessions, count)
}

// Handler returns an HTTP handler that serves Prometheus-format metrics.
func (m *Metrics) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m.mu.RLock()
		defer m.mu.RUnlock()

		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

		// Request counters
		fmt.Fprintln(w, "# HELP mcp_requests_total Total number of MCP requests by method")
		fmt.Fprintln(w, "# TYPE mcp_requests_total counter")
		for method, count := range m.requestsTotal {
			fmt.Fprintf(w, "mcp_requests_total{method=\"%s\"} %d\n", method, atomic.LoadUint64(count))
		}

		// Request duration (simplified - report average and p99)
		if len(m.requestDurations) > 0 {
			var sum float64
			for _, d := range m.requestDurations {
				sum += d
			}
			avg := sum / float64(len(m.requestDurations))

			fmt.Fprintln(w, "# HELP mcp_request_duration_seconds Request duration in seconds")
			fmt.Fprintln(w, "# TYPE mcp_request_duration_seconds gauge")
			fmt.Fprintf(w, "mcp_request_duration_seconds_avg %f\n", avg)
		}

		// Active sessions
		fmt.Fprintln(w, "# HELP mcp_active_sessions Current number of active sessions")
		fmt.Fprintln(w, "# TYPE mcp_active_sessions gauge")
		fmt.Fprintf(w, "mcp_active_sessions %d\n", atomic.LoadInt64(m.activeSessions))

		// Error counters
		fmt.Fprintln(w, "# HELP mcp_errors_total Total number of errors by code")
		fmt.Fprintln(w, "# TYPE mcp_errors_total counter")
		for code, count := range m.errorsTotal {
			fmt.Fprintf(w, "mcp_errors_total{code=\"%d\"} %d\n", code, atomic.LoadUint64(count))
		}

		// Uptime
		fmt.Fprintln(w, "# HELP mcp_uptime_seconds Server uptime in seconds")
		fmt.Fprintln(w, "# TYPE mcp_uptime_seconds gauge")
		fmt.Fprintf(w, "mcp_uptime_seconds %f\n", time.Since(m.startTime).Seconds())
	}
}
