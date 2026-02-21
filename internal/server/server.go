package server

import (
	_ "embed"
	"log"
	"net/http"

	"local-clipboard/internal/history"
	"local-clipboard/internal/store"
)

//go:embed static/index.html
var indexHTML []byte

// Config holds server options.
type Config struct {
	Addr     string // Listen address, e.g. ":8080"
	DBPath   string // Path to SQLite database
	StaticDir string // Root directory for static files (e.g. "web/dist"). Empty = use embedded fallback.
}

// Run starts the HTTP server. It does not return unless there is a fatal error.
func Run(cfg Config) {
	h := history.NewSqlite(cfg.DBPath)
	if err := h.Init(); err != nil {
		log.Fatalf("failed to initialize sqlite history: %v", err)
	}
	st := store.New()
	if latest, err := h.Latest(); err == nil {
		st.Set(latest)
	}
	requestLogs := NewRequestLogs()
	app := &App{Store: st, History: h, Logs: requestLogs}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/clipboard", app.handleClipboard)
	mux.HandleFunc("/api/history", app.handleHistory)
	mux.HandleFunc("/api/history/pin", app.handlePin)
	mux.HandleFunc("/api/history/delete", app.handleDelete)
	mux.HandleFunc("/api/logs", app.handleLogs)
	mux.Handle("/", &spaHandler{rootDir: cfg.StaticDir, embed: indexHTML})

	handler := loggingMiddleware(requestLogs, mux)
	log.Printf("clipboard server listening on %s", cfg.Addr)
	log.Fatal(http.ListenAndServe(cfg.Addr, handler))
}
