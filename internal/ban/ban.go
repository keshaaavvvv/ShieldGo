package ban

import (
	"sync"
	"time"
)

type entry struct {
	bannedUntil time.Time
	strikes     int
}

type Manager struct {
	mu          sync.Mutex
	entries     map[string]*entry
	strikeLimit int
	initialBan  time.Duration
	maxBan      time.Duration
}

func New(
	strikeLimit int,
	initialBan time.Duration,
	maxBan time.Duration,
) *Manager {
	return &Manager{
		entries:     make(map[string]*entry),
		strikeLimit: strikeLimit,
		initialBan:  initialBan,
		maxBan:      maxBan,
	}
}

func (m *Manager) IsBanned(ip string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	e, ok := m.entries[ip]
	if !ok {
		return false
	}

	return time.Now().Before(e.bannedUntil)
}

func (m *Manager) Violation(ip string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()

	e, ok := m.entries[ip]
	if !ok {
		e = &entry{}
		m.entries[ip] = e
	}

	e.strikes++

	if e.strikes < m.strikeLimit {
		return false
	}

	banDuration := m.initialBan
	repeats := e.strikes / m.strikeLimit

	for i := 1; i < repeats; i++ {
		banDuration *= 2

		if banDuration >= m.maxBan {
			banDuration = m.maxBan
			break
		}
	}

	if banDuration > m.maxBan {
		banDuration = m.maxBan
	}

	e.bannedUntil = now.Add(banDuration)

	return true
}

func (m *Manager) Cleanup(maxAge time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)

	for ip, e := range m.entries {
		if e.bannedUntil.Before(cutoff) {
			delete(m.entries, ip)
		}
	}
}
