package models

// Media types supported by the player. "blank" is a deliberate item in a
// playlist (e.g. a pause/black screen) — it is never inserted automatically,
// only when explicitly configured, per the assignment's requirement.
const (
	MediaImage = "image"
	MediaVideo = "video"
	MediaBlank = "blank"
)

func IsValidMediaType(t string) bool {
	switch t {
	case MediaImage, MediaVideo, MediaBlank:
		return true
	default:
		return false
	}
}

// MediaItem is one entry in a window's playlist.
// DurationSeconds is how long this item stays on screen before the
// playlist advances to the next item.
type MediaItem struct {
	ID              int64  `json:"id"`
	WindowID        int64  `json:"window_id"`
	OrderIndex      int    `json:"order_index"`
	Type            string `json:"type"`
	URL             string `json:"url"`
	DurationSeconds int    `json:"duration_seconds"`
}

// Window represents one display window with its own independent,
// continuously looping playlist.
type Window struct {
	ID    int64       `json:"id"`
	Name  string      `json:"name"`
	Items []MediaItem `json:"items"`
}

// CycleSeconds is the total play size each window is treated as looping
// within, per the assignment brief: 5 hours = 18000 seconds.
const CycleSeconds = 5 * 60 * 60
