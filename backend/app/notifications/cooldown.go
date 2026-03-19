package notifications

import (
	"context"
	"sync"
	"time"
)

type cooldownTracker struct {
	mu    sync.RWMutex
	fired map[int]time.Time
}

var cooldowns = &cooldownTracker{fired: make(map[int]time.Time)}

func (m *cooldownTracker) canFire(ruleId int, cooldownMinutes int) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	last, ok := m.fired[ruleId]
	if !ok {
		return true
	}
	return time.Since(last) > time.Duration(cooldownMinutes)*time.Minute
}

func (m *cooldownTracker) recordFire(ruleId int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.fired[ruleId] = time.Now()
}

type dedupTracker struct {
	mu   sync.RWMutex
	seen map[string]time.Time
}

var dedup = &dedupTracker{seen: make(map[string]time.Time)}

func (m *dedupTracker) isDuplicate(key string, cooldown time.Duration) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	last, ok := m.seen[key]
	return ok && time.Since(last) < cooldown
}

func (m *dedupTracker) record(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.seen[key] = time.Now()
}

func (m *dedupTracker) purgeExpired(maxAge time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	for k, t := range m.seen {
		if now.Sub(t) > maxAge {
			delete(m.seen, k)
		}
	}
}

func startDedupPurger(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				dedup.purgeExpired(24 * time.Hour)
			}
		}
	}()
}
