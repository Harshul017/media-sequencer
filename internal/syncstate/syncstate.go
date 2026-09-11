package syncstate

import (
	"sync"
	"time"
)

// State describes what every window should be showing when a sync is
// active. This is intentionally kept in memory, not the database — it's
// transient, short-lived coordination state, not data that needs to
// survive a restart. See README assumptions for why.
type State struct {
	Active  bool   `json:"active"`
	Type    string `json:"type,omitempty"`
	URL     string `json:"url,omitempty"`
	EndsAt  string `json:"ends_at,omitempty"` // RFC3339, empty when inactive
}

// Manager is a small, mutex-protected holder for the current sync state,
// safe to read/write from concurrent HTTP requests.
type Manager struct {
	mu    sync.RWMutex
	state State
}

func NewManager() *Manager {
	return &Manager{state: State{Active: false}}
}

// Trigger starts a sync: every window should show this item until the
// given duration elapses.
func (m *Manager) Trigger(mediaType, url string, durationSeconds int) State {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state = State{
		Active: true,
		Type:   mediaType,
		URL:    url,
		EndsAt: time.Now().UTC().Add(time.Duration(durationSeconds) * time.Second).Format(time.RFC3339),
	}
	return m.state
}

// Current returns the current sync state, automatically expiring it if
// its end time has already passed — so callers never see a stale
// "active: true" after the duration is actually over.
func (m *Manager) Current() State {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.state.Active {
		endsAt, err := time.Parse(time.RFC3339, m.state.EndsAt)
		if err == nil && time.Now().UTC().After(endsAt) {
			m.state = State{Active: false}
		}
	}
	return m.state
}
