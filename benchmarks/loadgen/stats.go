package main

import (
	"sort"
	"sync"
)

type latencyTracker struct {
	mu      sync.Mutex
	samples []float64
	ok      int64
	errors  int64
}

func newLatencyTracker() *latencyTracker {
	return &latencyTracker{samples: make([]float64, 0, 8192)}
}

func (l *latencyTracker) Record(milliseconds float64, err error) {
	l.mu.Lock()
	if err != nil {
		l.errors++
	} else {
		l.ok++
		l.samples = append(l.samples, milliseconds)
	}
	l.mu.Unlock()
}

type latencySnapshot struct {
	P50     float64 `json:"p50"`
	P95     float64 `json:"p95"`
	P99     float64 `json:"p99"`
	OK      int64   `json:"ok"`
	Errors  int64   `json:"errors"`
	ErrRate float64 `json:"errRate"`
}

func (l *latencyTracker) SnapshotAndReset() latencySnapshot {
	l.mu.Lock()
	defer l.mu.Unlock()

	s := latencySnapshot{OK: l.ok, Errors: l.errors}
	total := l.ok + l.errors
	if total > 0 {
		s.ErrRate = float64(l.errors) / float64(total)
	}
	if len(l.samples) > 0 {
		sort.Float64s(l.samples)
		s.P50 = percentile(l.samples, 0.50)
		s.P95 = percentile(l.samples, 0.95)
		s.P99 = percentile(l.samples, 0.99)
	}

	l.samples = l.samples[:0]
	l.ok = 0
	l.errors = 0
	return s
}

func percentile(sorted []float64, q float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(float64(len(sorted)-1) * q)
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}
