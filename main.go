package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

type clipboardUpdate struct {
	ID        int64     `json:"id"`
	Text      string    `json:"text"`
	Source    string    `json:"source"`
	UpdatedAt time.Time `json:"updated_at"`
}

type clipboardStore struct {
	mu     sync.RWMutex
	latest clipboardUpdate
}

func (s *clipboardStore) set(v clipboardUpdate) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.latest = v
}

func (s *clipboardStore) get() clipboardUpdate {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.latest
}

type sqliteHistory struct {
	path string
	mu   sync.Mutex
}

type app struct {
	store   *clipboardStore
	history *sqliteHistory
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run . <server|client> [flags]")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "server":
		runServer(os.Args[2:])
	case "client":
		runClient(os.Args[2:])
	default:
		fmt.Printf("unknown mode %q, expected server or client\n", os.Args[1])
		os.Exit(1)
	}
}

func runServer(args []string) {
	fs := flag.NewFlagSet("server", flag.ExitOnError)
	addr := fs.String("addr", ":8080", "listen address for the web server")
	dbPath := fs.String("db", "clipboard.db", "path to sqlite database")
	_ = fs.Parse(args)

	history := &sqliteHistory{path: *dbPath}
	if err := history.init(); err != nil {
		log.Fatalf("failed to initialize sqlite history: %v", err)
	}

	a := &app{store: &clipboardStore{}, history: history}
	if latest, err := history.latest(); err == nil {
		a.store.set(latest)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", a.handleIndex)
	mux.HandleFunc("/api/clipboard", a.handleClipboard)
	mux.HandleFunc("/api/history", a.handleHistory)

	log.Printf("clipboard server listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}

func (a *app) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(indexHTML))
}

func (a *app) handleClipboard(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var req struct {
			Text   string `json:"text"`
			Source string `json:"source"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		req.Text = strings.TrimSpace(req.Text)
		if req.Text == "" {
			http.Error(w, "text is required", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(req.Source) == "" {
			req.Source = "unknown"
		}
		entry, err := a.history.insert(req.Text, req.Source)
		if err != nil {
			http.Error(w, "failed to save clipboard", http.StatusInternalServerError)
			return
		}
		a.store.set(entry)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(entry)
	case http.MethodGet:
		latest := a.store.get()
		if strings.TrimSpace(latest.Text) == "" {
			http.Error(w, "clipboard is empty", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(latest)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (a *app) handleHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	limit := 25
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > 200 {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}
		limit = parsed
	}
	history, err := a.history.list(limit)
	if err != nil {
		http.Error(w, "failed to read history", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(history)
}

func runClient(args []string) { /* unchanged */
	fs := flag.NewFlagSet("client", flag.ExitOnError)
	serverURL := fs.String("server", "http://127.0.0.1:8080", "base URL of clipboard server")
	interval := fs.Duration("interval", 1*time.Second, "poll interval for local clipboard")
	source := fs.String("source", hostName(), "source label for this machine")
	_ = fs.Parse(args)

	localRead, localWrite, err := detectClipboardCommands()
	if err != nil {
		log.Fatalf("clipboard command setup failed: %v", err)
	}
	if localWrite == nil {
		log.Printf("note: no clipboard write command found; this client will only push local copy events")
	}

	var lastSent string
	for {
		text, err := runClipboardRead(localRead)
		if err != nil {
			time.Sleep(*interval)
			continue
		}
		text = strings.TrimSpace(text)
		if text != "" && text != lastSent {
			if err := postClipboard(*serverURL, text, *source); err == nil {
				lastSent = text
			}
		}
		if localWrite != nil {
			if remote, err := fetchClipboard(*serverURL); err == nil && remote.Text != "" && remote.Source != *source && remote.Text != text {
				if err := runClipboardWrite(localWrite, remote.Text); err == nil {
					lastSent = remote.Text
				}
			}
		}
		time.Sleep(*interval)
	}
}

type clipboardCmd struct {
	name string
	args []string
}

func detectClipboardCommands() (clipboardCmd, *clipboardCmd, error) {
	if _, err := exec.LookPath("wl-paste"); err == nil {
		if _, err := exec.LookPath("wl-copy"); err == nil {
			w := &clipboardCmd{name: "wl-copy"}
			return clipboardCmd{name: "wl-paste"}, w, nil
		}
		return clipboardCmd{name: "wl-paste"}, nil, nil
	}
	if _, err := exec.LookPath("xclip"); err == nil {
		w := &clipboardCmd{name: "xclip", args: []string{"-selection", "clipboard"}}
		return clipboardCmd{name: "xclip", args: []string{"-o", "-selection", "clipboard"}}, w, nil
	}
	if _, err := exec.LookPath("xsel"); err == nil {
		w := &clipboardCmd{name: "xsel", args: []string{"--clipboard", "--input"}}
		return clipboardCmd{name: "xsel", args: []string{"--clipboard", "--output"}}, w, nil
	}
	return clipboardCmd{}, nil, errors.New("no supported clipboard command found (install wl-clipboard, xclip, or xsel)")
}
func runClipboardRead(cmd clipboardCmd) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	c := exec.CommandContext(ctx, cmd.name, cmd.args...)
	out, err := c.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
func runClipboardWrite(cmd *clipboardCmd, text string) error {
	if cmd == nil {
		return errors.New("clipboard write command not configured")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	c := exec.CommandContext(ctx, cmd.name, cmd.args...)
	c.Stdin = strings.NewReader(text)
	return c.Run()
}
func postClipboard(serverURL, text, source string) error {
	payload := map[string]string{"text": text, "source": source}
	b, _ := json.Marshal(payload)
	resp, err := http.Post(strings.TrimRight(serverURL, "/")+"/api/clipboard", "application/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("unexpected status %s", resp.Status)
	}
	return nil
}
func fetchClipboard(serverURL string) (clipboardUpdate, error) {
	resp, err := http.Get(strings.TrimRight(serverURL, "/") + "/api/clipboard")
	if err != nil {
		return clipboardUpdate{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return clipboardUpdate{}, fmt.Errorf("status %s", resp.Status)
	}
	var out clipboardUpdate
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return clipboardUpdate{}, err
	}
	return out, nil
}
func hostName() string {
	h, err := os.Hostname()
	if err != nil || strings.TrimSpace(h) == "" {
		return "linux-client"
	}
	return h
}

func (s *sqliteHistory) init() error {
	_, err := s.runSQL(`CREATE TABLE IF NOT EXISTS clipboard_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		text TEXT NOT NULL,
		source TEXT NOT NULL,
		updated_at TEXT NOT NULL
	);`)
	return err
}

func (s *sqliteHistory) insert(text, source string) (clipboardUpdate, error) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	query := fmt.Sprintf("INSERT INTO clipboard_history(text,source,updated_at) VALUES(%s,%s,%s); SELECT last_insert_rowid();", sqlQuote(text), sqlQuote(source), sqlQuote(now))
	out, err := s.runSQL(query)
	if err != nil {
		return clipboardUpdate{}, err
	}
	id, err := strconv.ParseInt(strings.TrimSpace(out), 10, 64)
	if err != nil {
		return clipboardUpdate{}, err
	}
	return s.byID(id)
}

func (s *sqliteHistory) latest() (clipboardUpdate, error) {
	rows, err := s.selectRows("SELECT id,text,source,updated_at FROM clipboard_history ORDER BY id DESC LIMIT 1;")
	if err != nil || len(rows) == 0 {
		if err == nil {
			err = errors.New("no rows")
		}
		return clipboardUpdate{}, err
	}
	return rows[0], nil
}

func (s *sqliteHistory) byID(id int64) (clipboardUpdate, error) {
	rows, err := s.selectRows(fmt.Sprintf("SELECT id,text,source,updated_at FROM clipboard_history WHERE id=%d LIMIT 1;", id))
	if err != nil || len(rows) == 0 {
		if err == nil {
			err = errors.New("not found")
		}
		return clipboardUpdate{}, err
	}
	return rows[0], nil
}

func (s *sqliteHistory) list(limit int) ([]clipboardUpdate, error) {
	return s.selectRows(fmt.Sprintf("SELECT id,text,source,updated_at FROM clipboard_history ORDER BY id DESC LIMIT %d;", limit))
}

func (s *sqliteHistory) selectRows(query string) ([]clipboardUpdate, error) {
	out, err := s.runSQL(".mode tabs\n" + query)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(out) == "" {
		return []clipboardUpdate{}, nil
	}
	scanner := bufio.NewScanner(strings.NewReader(out))
	var rows []clipboardUpdate
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), "\t", 4)
		if len(parts) != 4 {
			continue
		}
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		t, _ := time.Parse(time.RFC3339Nano, parts[3])
		rows = append(rows, clipboardUpdate{ID: id, Text: parts[1], Source: parts[2], UpdatedAt: t})
	}
	return rows, scanner.Err()
}

func (s *sqliteHistory) runSQL(sql string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cmd := exec.Command("sqlite3", s.path)
	cmd.Stdin = strings.NewReader(sql)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("sqlite3: %w: %s", err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

func sqlQuote(v string) string { return "'" + strings.ReplaceAll(v, "'", "''") + "'" }

const indexHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>Local Clipboard Bridge</title>
  <style>
    :root { color-scheme: light dark; }
    * { box-sizing: border-box; }
    body { font-family: Inter, system-ui, -apple-system, sans-serif; margin: 0; background: #0b1020; color: #e6ebff; }
    .wrap { max-width: 1000px; margin: 0 auto; padding: 1rem; }
    .grid { display: grid; gap: 1rem; grid-template-columns: 1fr; }
    .card { background: rgba(255,255,255,0.06); border: 1px solid rgba(255,255,255,0.12); border-radius: 14px; padding: 1rem; }
    h1,h2 { margin: 0 0 .6rem; }
    textarea { width: 100%; min-height: 120px; border-radius: 10px; border: 1px solid #334; padding: .8rem; background:#0f1730; color:#fff; }
    button { margin-top: .6rem; border: 0; border-radius: 10px; padding: .65rem .9rem; background: #4f7cff; color: white; cursor: pointer; }
    .muted { color: #a7b1d3; font-size: .95rem; }
    .history { display: grid; gap: .6rem; max-height: 55vh; overflow: auto; }
    .item { border: 1px solid rgba(255,255,255,.15); border-radius: 10px; padding: .65rem; background: rgba(0,0,0,.2); }
    .item-top { display:flex; justify-content:space-between; gap:1rem; font-size:.85rem; color:#b7c0e5; }
    .item pre { margin:.4rem 0 0; white-space:pre-wrap; word-break:break-word; }
    @media (min-width: 860px) { .grid { grid-template-columns: 1fr 1fr; } }
  </style>
</head>
<body>
<div class="wrap">
  <h1>📋 Local Clipboard Bridge</h1>
  <p class="muted">Responsive UI + SQLite history.</p>
  <div class="grid">
    <section class="card">
      <h2>Send text</h2>
      <textarea id="clipInput" placeholder="Paste text here"></textarea>
      <button id="sendBtn">Send to server</button>
      <p id="sendStatus" class="muted"></p>
      <h2 style="margin-top:1rem">Latest</h2>
      <pre id="latest" class="item">Loading...</pre>
    </section>
    <section class="card">
      <h2>History</h2>
      <p class="muted">Newest first (up to 25 items)</p>
      <div id="history" class="history"></div>
    </section>
  </div>
</div>
<script>
function escHtml(value) {
  return (value || '').replace(/[<>&]/g, function(m){ return ({'<':'&lt;','>':'&gt;','&':'&amp;'})[m]; });
}

async function loadLatest() {
  const latestEl = document.getElementById('latest');
  try {
    const res = await fetch('/api/clipboard');
    if (!res.ok) { latestEl.textContent = 'Clipboard is empty.'; return; }
    const data = await res.json();
    latestEl.textContent = data.text + "\n\nsource: " + data.source + "\nat: " + new Date(data.updated_at).toLocaleString();
  } catch (err) { latestEl.textContent = 'Failed: ' + err.message; }
}

async function loadHistory() {
  const host = document.getElementById('history');
  try {
    const res = await fetch('/api/history?limit=25');
    if (!res.ok) { host.innerHTML = '<p class="muted">No history yet.</p>'; return; }
    const items = await res.json();
    if (!items.length) { host.innerHTML = '<p class="muted">No history yet.</p>'; return; }
    host.innerHTML = items.map(function(x){
      return '<article class="item"><div class="item-top"><span>' + escHtml(x.source) + '</span><span>' + new Date(x.updated_at).toLocaleString() + '</span></div><pre>' + escHtml(x.text) + '</pre></article>';
    }).join('');
  } catch (err) {
    host.innerHTML = '<p class="muted">Failed to load history.</p>';
  }
}

document.getElementById('sendBtn').addEventListener('click', async function(){
  const text = document.getElementById('clipInput').value.trim();
  if (!text) {
    document.getElementById('sendStatus').textContent = 'Please enter text first.';
    return;
  }
  const res = await fetch('/api/clipboard', {
    method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ text: text, source: 'mobile-web' })
  });
  document.getElementById('sendStatus').textContent = res.ok ? 'Sent successfully.' : 'Send failed.';
  await loadLatest();
  await loadHistory();
});

loadLatest();
loadHistory();
setInterval(function(){ loadLatest(); loadHistory(); }, 3000);
</script>
</body>
</html>`
