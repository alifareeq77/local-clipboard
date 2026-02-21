package server

import (
	"local-clipboard/internal/history"
	"local-clipboard/internal/store"
)

// App holds server dependencies (in-memory store, history, and request logs).
type App struct {
	Store   *store.Store
	History history.History
	Logs    *RequestLogs
}
