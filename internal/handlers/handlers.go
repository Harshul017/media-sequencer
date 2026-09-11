package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"media-sequencer/internal/models"
	"media-sequencer/internal/store"
	"media-sequencer/internal/syncstate"
)

type Handlers struct {
	Store *store.Store
	Sync  *syncstate.Manager
}

func New(s *store.Store, sm *syncstate.Manager) *Handlers {
	return &Handlers{Store: s, Sync: sm}
}

// windowsResponse wraps the window list with the server's current time.
// The frontend uses server_time (not its own clock) as the reference
// point for computing playback position, so that all windows — even on
// different machines with slightly different clocks — agree on where
// they are in the 5-hour cycle.
type windowsResponse struct {
	ServerTime string          `json:"server_time"`
	CycleSeconds int           `json:"cycle_seconds"`
	Windows    []models.Window `json:"windows"`
}

// ListWindows handles GET /windows.
func (h *Handlers) ListWindows(w http.ResponseWriter, r *http.Request) {
	windows, err := h.Store.ListWindows()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load windows")
		return
	}
	writeJSON(w, http.StatusOK, windowsResponse{
		ServerTime:   time.Now().UTC().Format(time.RFC3339),
		CycleSeconds: models.CycleSeconds,
		Windows:      windows,
	})
}

type addMediaRequest struct {
	Type            string `json:"type"`
	URL             string `json:"url"`
	DurationSeconds int    `json:"duration_seconds"`
}

// AddMedia handles POST /windows/{id}/media — appends a new item to the
// end of that window's playlist.
func (h *Handlers) AddMedia(w http.ResponseWriter, r *http.Request) {
	windowID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid window id")
		return
	}

	exists, err := h.Store.WindowExists(windowID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not verify window")
		return
	}
	if !exists {
		writeError(w, http.StatusNotFound, "window not found")
		return
	}

	var req addMediaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if !models.IsValidMediaType(req.Type) {
		writeError(w, http.StatusBadRequest, "type must be image, video, or blank")
		return
	}
	if req.Type != models.MediaBlank && req.URL == "" {
		writeError(w, http.StatusBadRequest, "url is required for image/video items")
		return
	}
	if req.DurationSeconds <= 0 {
		writeError(w, http.StatusBadRequest, "duration_seconds must be greater than 0")
		return
	}

	count, err := h.Store.CountMediaItems(windowID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not count existing media")
		return
	}

	item, err := h.Store.AddMediaItem(windowID, count, req.Type, req.URL, req.DurationSeconds)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not add media item")
		return
	}

	writeJSON(w, http.StatusCreated, item)
}

// GetSyncState handles GET /sync-state — polled by every window roughly
// once a second.
func (h *Handlers) GetSyncState(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.Sync.Current())
}

type triggerSyncRequest struct {
	Type            string `json:"type"`
	URL             string `json:"url"`
	DurationSeconds int    `json:"duration_seconds"`
}

// TriggerSync handles POST /sync — starts a synced moment across all
// windows for the given duration.
func (h *Handlers) TriggerSync(w http.ResponseWriter, r *http.Request) {
	var req triggerSyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if !models.IsValidMediaType(req.Type) {
		writeError(w, http.StatusBadRequest, "type must be image, video, or blank")
		return
	}
	if req.DurationSeconds <= 0 {
		writeError(w, http.StatusBadRequest, "duration_seconds must be greater than 0")
		return
	}

	state := h.Sync.Trigger(req.Type, req.URL, req.DurationSeconds)
	writeJSON(w, http.StatusOK, state)
}
