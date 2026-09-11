package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"media-sequencer/internal/handlers"
	"media-sequencer/internal/middleware"
	"media-sequencer/internal/store"
	"media-sequencer/internal/syncstate"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "media-sequencer.db"
	}

	s, err := store.New(dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	syncMgr := syncstate.NewManager()
	h := handlers.New(s, syncMgr)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /windows", h.ListWindows)
	mux.HandleFunc("POST /windows/{id}/media", h.AddMedia)
	mux.HandleFunc("GET /sync-state", h.GetSyncState)
	mux.HandleFunc("POST /sync", h.TriggerSync)

	handler := middleware.CORS(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}
