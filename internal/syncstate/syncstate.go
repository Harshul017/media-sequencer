package syncstate

import (
	"sync"
	"time"
)

// State describes what every window should be showing when a sync is
// active, plus how much total time has ever been "spent" in sync — see
// TotalPausedSeconds below.
type State struct {
	Active             bool   `json:"active"`
	Type               string `json:"type,omitempty"`
	URL                string `json:"url,omitempty"`
	EndsAt             string `json:"ends_at,omitempty"` // RFC3339, empty when inactive
	TotalPausedSeconds int    `json:"total_paused_seconds"`
}

// Manager is a small, mutex-protected holder for the current sync state,
// safe to read/write from concurrent HTTP requests.
type Manager struct {
	mu                 sync.RWMutex
	state              State
	totalPausedSeconds int
}

func NewManager() *Manager {
	return &Manager{state: State{Active: false}}
}

// Trigger starts a sync: every window should show this item until the
// given duration elapses.
//
// It also immediately adds the full duration to totalPausedSeconds. This
// is what makes normal playback truly pause rather than skip ahead: every
// window computes its position as (wall-clock time - totalPausedSeconds),
// so the instant this number increases, each window's computed position
// jumps backward by exactly the sync duration, then ticks forward
// normally — arriving back at today's real time only once the sync
// window has actually elapsed. The net effect is a true pause-and-resume
// with no per-window state needed.
func (m *Manager) Trigger(mediaType, url string, durationSeconds int) State {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.totalPausedSeconds += durationSeconds
	m.state = State{
		Active:             true,
		Type:               mediaType,
		URL:                url,
		EndsAt:             time.Now().UTC().Add(time.Duration(durationSeconds) * time.Second).Format(time.RFC3339),
		TotalPausedSeconds: m.totalPausedSeconds,
	}
	return m.state
}

// Current returns the current sync state, automatically expiring the
// "active" flag if its end time has already passed — so callers never
// see a stale "active: true" after the duration is actually over.
// TotalPausedSeconds is never reset; it only ever grows.
func (m *Manager) Current() State {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.state.Active {
		endsAt, err := time.Parse(time.RFC3339, m.state.EndsAt)
		if err == nil && time.Now().UTC().After(endsAt) {
			m.state = State{Active: false}
		}
	}
	m.state.TotalPausedSeconds = m.totalPausedSeconds
	return m.state
}
