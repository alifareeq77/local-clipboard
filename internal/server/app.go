package server

import (
	"local-clipboard/internal/history"
	"local-clipboard/internal/store"
)

// App holds server dependencies (in-memory store, history, request logs, and server info).
type App struct {
	Store      *store.Store
	History    history.History
	Logs       *RequestLogs
	ServerURLs []string // LAN URLs where this server is reachable (e.g. http://192.168.1.5:8080)
}
