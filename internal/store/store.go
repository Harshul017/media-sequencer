package store

import (
	"database/sql"
	"errors"

	"media-sequencer/internal/models"

	_ "modernc.org/sqlite"
)

var ErrNotFound = errors.New("not found")

type Store struct {
	db *sql.DB
}

func New(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		return nil, err
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, err
	}
	if err := s.seedIfEmpty(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS windows (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS media_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		window_id INTEGER NOT NULL,
		order_index INTEGER NOT NULL,
		type TEXT NOT NULL,
		url TEXT NOT NULL,
		duration_seconds INTEGER NOT NULL,
		FOREIGN KEY (window_id) REFERENCES windows(id)
	);
	`
	_, err := s.db.Exec(schema)
	return err
}

// seedIfEmpty inserts example windows/media matching the assignment's
// scenario, but only if the database is fresh — so restarting the app
// doesn't duplicate seed data.
func (s *Store) seedIfEmpty() error {
	var count int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM windows`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	seed := []struct {
		name  string
		items []models.MediaItem
	}{
		{
			name: "Window 1",
			items: []models.MediaItem{
				{Type: models.MediaImage, URL: "https://picsum.photos/seed/m1/800/450", DurationSeconds: 5},
				{Type: models.MediaImage, URL: "https://picsum.photos/seed/m2/800/450", DurationSeconds: 5},
				{Type: models.MediaVideo, URL: "https://interactive-examples.mdn.mozilla.net/media/cc0-videos/flower.mp4", DurationSeconds: 10},
			},
		},
		{
			name: "Window 2",
			items: []models.MediaItem{
				{Type: models.MediaImage, URL: "https://picsum.photos/seed/m2/800/450", DurationSeconds: 5},
				{Type: models.MediaVideo, URL: "https://interactive-examples.mdn.mozilla.net/media/cc0-videos/flower.mp4", DurationSeconds: 8},
				{Type: models.MediaBlank, URL: "", DurationSeconds: 3},
			},
		},
		{
			name: "Window 3",
			items: []models.MediaItem{
				{Type: models.MediaVideo, URL: "https://interactive-examples.mdn.mozilla.net/media/cc0-videos/flower.mp4", DurationSeconds: 12},
				{Type: models.MediaImage, URL: "https://picsum.photos/seed/m2/800/450", DurationSeconds: 5},
			},
		},
	}

	for _, w := range seed {
		windowID, err := s.CreateWindow(w.name)
		if err != nil {
			return err
		}
		for i, item := range w.items {
			if _, err := s.AddMediaItem(windowID, i, item.Type, item.URL, item.DurationSeconds); err != nil {
				return err
			}
		}
	}
	return nil
}

// CreateWindow inserts a new window and returns its ID.
func (s *Store) CreateWindow(name string) (int64, error) {
	res, err := s.db.Exec(`INSERT INTO windows (name) VALUES (?)`, name)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// ListWindows returns every window with its playlist, ordered correctly.
func (s *Store) ListWindows() ([]models.Window, error) {
	rows, err := s.db.Query(`SELECT id, name FROM windows ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	windows := []models.Window{}
	for rows.Next() {
		var w models.Window
		if err := rows.Scan(&w.ID, &w.Name); err != nil {
			return nil, err
		}
		windows = append(windows, w)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range windows {
		items, err := s.getMediaForWindow(windows[i].ID)
		if err != nil {
			return nil, err
		}
		windows[i].Items = items
	}
	return windows, nil
}

// GetWindow fetches a single window with its playlist.
func (s *Store) GetWindow(id int64) (*models.Window, error) {
	row := s.db.QueryRow(`SELECT id, name FROM windows WHERE id = ?`, id)
	var w models.Window
	if err := row.Scan(&w.ID, &w.Name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	items, err := s.getMediaForWindow(w.ID)
	if err != nil {
		return nil, err
	}
	w.Items = items
	return &w, nil
}

func (s *Store) getMediaForWindow(windowID int64) ([]models.MediaItem, error) {
	rows, err := s.db.Query(
		`SELECT id, window_id, order_index, type, url, duration_seconds
		 FROM media_items WHERE window_id = ? ORDER BY order_index`, windowID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []models.MediaItem{}
	for rows.Next() {
		var m models.MediaItem
		if err := rows.Scan(&m.ID, &m.WindowID, &m.OrderIndex, &m.Type, &m.URL, &m.DurationSeconds); err != nil {
			return nil, err
		}
		items = append(items, m)
	}
	return items, rows.Err()
}

// AddMediaItem appends a new item to a window's playlist at the given
// order_index (typically "current item count", i.e. appended to the end).
func (s *Store) AddMediaItem(windowID int64, orderIndex int, mediaType, url string, durationSeconds int) (*models.MediaItem, error) {
	res, err := s.db.Exec(
		`INSERT INTO media_items (window_id, order_index, type, url, duration_seconds)
		 VALUES (?, ?, ?, ?, ?)`,
		windowID, orderIndex, mediaType, url, durationSeconds,
	)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &models.MediaItem{
		ID: id, WindowID: windowID, OrderIndex: orderIndex,
		Type: mediaType, URL: url, DurationSeconds: durationSeconds,
	}, nil
}

// CountMediaItems returns how many items a window currently has —
// used to compute the next order_index when appending.
func (s *Store) CountMediaItems(windowID int64) (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM media_items WHERE window_id = ?`, windowID).Scan(&count)
	return count, err
}

// WindowExists checks a window ID is valid before adding media to it.
func (s *Store) WindowExists(windowID int64) (bool, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM windows WHERE id = ?`, windowID).Scan(&count)
	return count > 0, err
}
