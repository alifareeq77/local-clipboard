package server

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"sync"
	"time"
)

const maxRequestLogs = 500
const maxBodyLogSize = 64 * 1024 // 64KB per body

// RequestLogEntry is a single HTTP request log entry.
type RequestLogEntry struct {
	Method       string    `json:"method"`
	Path         string    `json:"path"`
	Status       int       `json:"status"`
	RemoteAddr   string    `json:"remote_addr"`
	Timestamp    time.Time `json:"timestamp"`
	RequestBody  string    `json:"request_body,omitempty"`
	ResponseBody string    `json:"response_body,omitempty"`
}

// RequestLogs holds in-memory request log entries (ring buffer).
type RequestLogs struct {
	mu     sync.RWMutex
	entries []RequestLogEntry
}

// NewRequestLogs creates a new request log store.
func NewRequestLogs() *RequestLogs {
	return &RequestLogs{entries: make([]RequestLogEntry, 0, maxRequestLogs)}
}

// Add appends a log entry. Keeps at most maxRequestLogs entries.
func (l *RequestLogs) Add(e RequestLogEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = append(l.entries, e)
	if len(l.entries) > maxRequestLogs {
		l.entries = l.entries[len(l.entries)-maxRequestLogs:]
	}
}

// List returns a copy of recent entries (newest first).
func (l *RequestLogs) List() []RequestLogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make([]RequestLogEntry, len(l.entries))
	copy(out, l.entries)
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out
}

// responseRecorder buffers the response to capture body, then flushes to w at the end.
type responseRecorder struct {
	http.ResponseWriter
	status int
	buf    bytes.Buffer
}

func (r *responseRecorder) WriteHeader(code int) {
	r.status = code
}

func (r *responseRecorder) Write(p []byte) (int, error) {
	return r.buf.Write(p)
}

func (r *responseRecorder) Status() int {
	if r.status == 0 {
		return http.StatusOK
	}
	return r.status
}

func truncateForLog(b []byte, max int) string {
	if len(b) == 0 {
		return ""
	}
	if len(b) > max {
		b = b[:max]
	}
	return string(b)
}

// loggingMiddleware logs each request and passes to next.
func loggingMiddleware(logs *RequestLogs, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "" {
			path = "/"
		}
		if path == "/api/logs" {
			next.ServeHTTP(w, r)
			return
		}

		var requestBody []byte
		if r.Body != nil && (r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch) {
			requestBody, _ = io.ReadAll(io.LimitReader(r.Body, maxBodyLogSize+1))
			r.Body = io.NopCloser(bytes.NewReader(requestBody))
		}

		rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		// Flush buffered response to client
		if rec.status != 0 {
			w.WriteHeader(rec.status)
		}
		respBytes := rec.buf.Bytes()
		w.Write(respBytes)

		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		if ip == "" {
			ip = r.RemoteAddr
		}
		logs.Add(RequestLogEntry{
			Method:       r.Method,
			Path:         path,
			Status:       rec.Status(),
			RemoteAddr:   ip,
			Timestamp:    time.Now().UTC(),
			RequestBody:  truncateForLog(requestBody, maxBodyLogSize),
			ResponseBody: truncateForLog(respBytes, maxBodyLogSize),
		})
	})
}
