package main

import (
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
	Pinned    bool      `json:"pinned"`
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
	mux.HandleFunc("/api/history/pin", a.handlePin)

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

	limit := 50
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > 200 {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}
		limit = parsed
	}
	query := strings.TrimSpace(r.URL.Query().Get("q"))

	history, err := a.history.list(limit, query)
	if err != nil {
		http.Error(w, "failed to read history", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(history)
}

func (a *app) handlePin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID     int64 `json:"id"`
		Pinned bool  `json:"pinned"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if req.ID <= 0 {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}
	if err := a.history.setPinned(req.ID, req.Pinned); err != nil {
		http.Error(w, "failed to update pin", http.StatusInternalServerError)
		return
	}
	entry, err := a.history.byID(req.ID)
	if err != nil {
		http.Error(w, "entry not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(entry)
}

func runClient(args []string) {
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
			remote, err := fetchClipboard(*serverURL)
			if err == nil && remote.Text != "" && remote.Source != *source && remote.Text != text {
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
		updated_at TEXT NOT NULL,
		pinned INTEGER NOT NULL DEFAULT 0
	);
	ALTER TABLE clipboard_history ADD COLUMN pinned INTEGER NOT NULL DEFAULT 0;`)
	if err == nil {
		return nil
	}
	// Ignore duplicate column errors from migration attempt.
	if strings.Contains(err.Error(), "duplicate column name") {
		return nil
	}
	return err
}

func (s *sqliteHistory) insert(text, source string) (clipboardUpdate, error) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	query := fmt.Sprintf("INSERT INTO clipboard_history(text,source,updated_at,pinned) VALUES(%s,%s,%s,0); SELECT last_insert_rowid();", sqlQuote(text), sqlQuote(source), sqlQuote(now))
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

func (s *sqliteHistory) setPinned(id int64, pinned bool) error {
	pinInt := 0
	if pinned {
		pinInt = 1
	}
	_, err := s.runSQL(fmt.Sprintf("UPDATE clipboard_history SET pinned=%d WHERE id=%d;", pinInt, id))
	return err
}

func (s *sqliteHistory) latest() (clipboardUpdate, error) {
	rows, err := s.selectRows("SELECT id,text,source,updated_at,pinned FROM clipboard_history ORDER BY pinned DESC, id DESC LIMIT 1;")
	if err != nil || len(rows) == 0 {
		if err == nil {
			err = errors.New("no rows")
		}
		return clipboardUpdate{}, err
	}
	return rows[0], nil
}

func (s *sqliteHistory) byID(id int64) (clipboardUpdate, error) {
	rows, err := s.selectRows(fmt.Sprintf("SELECT id,text,source,updated_at,pinned FROM clipboard_history WHERE id=%d LIMIT 1;", id))
	if err != nil || len(rows) == 0 {
		if err == nil {
			err = errors.New("not found")
		}
		return clipboardUpdate{}, err
	}
	return rows[0], nil
}

func (s *sqliteHistory) list(limit int, search string) ([]clipboardUpdate, error) {
	query := "SELECT id,text,source,updated_at,pinned FROM clipboard_history"
	if search != "" {
		query += " WHERE lower(text) LIKE " + sqlQuote("%"+strings.ToLower(search)+"%")
	}
	query += fmt.Sprintf(" ORDER BY pinned DESC, id DESC LIMIT %d;", limit)
	return s.selectRows(query)
}

func (s *sqliteHistory) selectRows(query string) ([]clipboardUpdate, error) {
	out, err := s.runSQL(".mode json\n" + query)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(out) == "" {
		return []clipboardUpdate{}, nil
	}

	var raw []struct {
		ID        int64  `json:"id"`
		Text      string `json:"text"`
		Source    string `json:"source"`
		UpdatedAt string `json:"updated_at"`
		Pinned    int    `json:"pinned"`
	}
	if err := json.Unmarshal([]byte(out), &raw); err != nil {
		return nil, err
	}

	rows := make([]clipboardUpdate, 0, len(raw))
	for _, row := range raw {
		t, _ := time.Parse(time.RFC3339Nano, row.UpdatedAt)
		rows = append(rows, clipboardUpdate{
			ID:        row.ID,
			Text:      row.Text,
			Source:    row.Source,
			UpdatedAt: t,
			Pinned:    row.Pinned == 1,
		})
	}
	return rows, nil
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

func sqlQuote(v string) string {
	return "'" + strings.ReplaceAll(v, "'", "''") + "'"
}

const indexHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1, viewport-fit=cover" />
  <title>Clipboard Bridge</title>
  <script src="https://cdn.tailwindcss.com"></script>
  <script>
    tailwind.config = {
      theme: {
        extend: {
          colors: {
            pine: {
              950: '#050c08',
              900: '#0a1711',
              800: '#102219',
              700: '#173528',
              500: '#46d88c',
              400: '#67d4a0',
              300: '#9fceb5'
            }
          }
        }
      }
    };
  </script>
</head>
<body class="bg-gradient-to-b from-[#132b1f] via-pine-950 to-[#030805] text-green-50 min-h-screen text-[13px] leading-tight">
  <div class="max-w-4xl mx-auto p-2 sm:p-3">
    <h1 class="text-sm sm:text-base font-semibold tracking-wide">📋 Local Clipboard</h1>
    <p class="text-[11px] text-pine-300 mt-0.5 mb-2">Tailwind compact mobile UI · searchable history · pin + copy actions.</p>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-2">
      <section class="rounded-xl border border-[#3c6b52] bg-[#0d1e16] shadow-lg shadow-black/30 p-2">
        <div class="flex items-center justify-between gap-2 mb-1.5">
          <h2 class="text-xs font-semibold">Send text</h2>
          <span class="text-[10px] text-pine-300">mobile → laptop</span>
        </div>
        <textarea id="clipInput" class="w-full min-h-[84px] rounded-lg border border-[#5ea884] bg-pine-900 px-2 py-2 text-[12px] text-green-50 focus:outline-none focus:ring-2 focus:ring-pine-500/40" placeholder="Paste text..."></textarea>
        <div class="mt-1.5 flex gap-1.5">
          <button id="sendBtn" class="rounded-lg border border-transparent bg-gradient-to-b from-[#52e498] to-pine-500 text-[#082515] font-semibold px-2.5 py-1.5 text-[11px]">Send</button>
          <button id="clearBtn" class="rounded-lg border border-[#33654d] bg-pine-700 text-green-100 font-semibold px-2.5 py-1.5 text-[11px]">Clear</button>
        </div>
        <p id="sendStatus" class="min-h-4 text-[11px] text-pine-300 mt-1"></p>

        <div class="flex items-center justify-between gap-2 mt-2 mb-1">
          <h2 class="text-xs font-semibold">Latest</h2>
          <span class="text-[10px] text-pine-300">current clipboard</span>
        </div>
        <pre id="latest" class="rounded-lg border border-[#5ea884] bg-pine-900 p-2 text-[12px] whitespace-pre-wrap break-words max-h-40 overflow-auto">Loading...</pre>
      </section>

      <section class="rounded-xl border border-[#3c6b52] bg-[#0d1e16] shadow-lg shadow-black/30 p-2">
        <div class="flex items-center justify-between gap-2 mb-1.5">
          <h2 class="text-xs font-semibold">History</h2>
          <span class="text-[10px] text-pine-300">pins stay on top</span>
        </div>
        <div class="grid grid-cols-[1fr_auto] gap-1.5 mb-1.5">
          <input id="searchInput" class="rounded-lg border border-[#5ea884] bg-pine-900 px-2 py-1.5 text-[12px] text-green-50 focus:outline-none focus:ring-2 focus:ring-pine-500/40" placeholder="Search copied text..." />
          <button id="searchBtn" class="rounded-lg border border-[#33654d] bg-pine-700 text-green-100 font-semibold px-2 py-1.5 text-[11px]">Find</button>
        </div>
        <div id="history" class="grid gap-1 max-h-[68vh] overflow-auto pr-0.5"></div>
      </section>
    </div>
  </div>

<script>
let currentQuery = '';
let currentItems = [];
let sending = false;

function escHtml(value) {
  return (value || '').replace(/[<>&]/g, function(m){ return ({'<':'&lt;','>':'&gt;','&':'&amp;'})[m]; });
}

function shortText(text) {
  if (!text) return '';
  if (text.length <= 240) return text;
  return text.slice(0, 240) + '…';
}

function iconClipboard() {
  return '<svg viewBox="0 0 24 24" class="w-3.5 h-3.5" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="3" width="6" height="4" rx="1"></rect><path d="M9 5H7a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V7a2 2 0 0 0-2-2h-2"></path></svg>';
}

function iconPin(filled) {
  if (filled) {
    return '<svg viewBox="0 0 24 24" class="w-3.5 h-3.5" fill="currentColor"><path d="M16 3l5 5-2 2-2-2-3 3v4l-2 2-1-1-4 4-1-1 4-4-1-1 2-2h4l3-3-2-2 2-2z"/></svg>';
  }
  return '<svg viewBox="0 0 24 24" class="w-3.5 h-3.5" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 3l7 7-3 3-3-3-3 3v4l-2 2-1-1-4 4-1-1 4-4-1-1 2-2h4l3-3-3-3 3-3z"/></svg>';
}

async function loadLatest() {
  const latestEl = document.getElementById('latest');
  try {
    const res = await fetch('/api/clipboard', { cache: 'no-store' });
    if (!res.ok) { latestEl.textContent = 'Clipboard is empty.'; return; }
    const data = await res.json();
    latestEl.textContent = data.text;
  } catch (err) {
    latestEl.textContent = 'Failed: ' + err.message;
  }
}

async function updatePin(id, pinned) {
  const res = await fetch('/api/history/pin', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ id: id, pinned: pinned })
  });
  if (!res.ok) throw new Error('Pin update failed');
}

function fallbackCopyText(txt) {
  const ta = document.createElement('textarea');
  ta.value = txt;
  ta.setAttribute('readonly', '');
  ta.style.position = 'fixed';
  ta.style.opacity = '0';
  ta.style.left = '-9999px';
  document.body.appendChild(ta);
  ta.select();
  ta.setSelectionRange(0, ta.value.length);
  let ok = false;
  try { ok = document.execCommand('copy'); } catch (e) { ok = false; }
  document.body.removeChild(ta);
  return ok;
}

async function copyText(txt, btn) {
  try {
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(txt);
      btn.innerHTML = '✓';
      setTimeout(function(){ btn.innerHTML = iconClipboard(); }, 900);
      return;
    }
    if (fallbackCopyText(txt)) {
      btn.innerHTML = '✓';
      setTimeout(function(){ btn.innerHTML = iconClipboard(); }, 900);
      return;
    }
    btn.innerHTML = '!';
  } catch (e) {
    if (fallbackCopyText(txt)) {
      btn.innerHTML = '✓';
      setTimeout(function(){ btn.innerHTML = iconClipboard(); }, 900);
      return;
    }
    btn.innerHTML = '!';
  }
}

async function loadHistory() {
  const host = document.getElementById('history');
  const query = currentQuery ? '&q=' + encodeURIComponent(currentQuery) : '';
  try {
    const res = await fetch('/api/history?limit=80' + query, { cache: 'no-store' });
    if (!res.ok) { host.innerHTML = '<div class="text-[11px] text-pine-300 text-center p-2 border border-dashed border-[#467c5f] rounded-lg">No history yet.</div>'; return; }
    currentItems = await res.json();
    if (!currentItems.length) { host.innerHTML = '<div class="text-[11px] text-pine-300 text-center p-2 border border-dashed border-[#467c5f] rounded-lg">No matching history.</div>'; return; }

    host.innerHTML = currentItems.map(function(item){
      const pinClass = item.pinned ? 'bg-[#226a47] border-[#67d4a0]' : 'bg-[#1a3a2b] border-[#3b7759]';
      return '<article class="rounded-lg border border-[#5ea884] bg-gradient-to-b from-[#102219] to-pine-900 p-1.5">'
        + '<div class="flex items-center justify-between gap-2">'
          + '<div class="flex gap-1">'
            + '<button class="icon-btn copy-btn inline-flex items-center justify-center w-7 h-7 rounded-md border border-[#3b7759] bg-[#1a3a2b] text-green-100" data-id="' + item.id + '" title="Copy">' + iconClipboard() + '</button>'
            + '<button class="icon-btn pin-btn inline-flex items-center justify-center w-7 h-7 rounded-md border ' + pinClass + ' text-green-100" data-id="' + item.id + '" data-pinned="' + item.pinned + '" title="Pin">' + iconPin(item.pinned) + '</button>'
          + '</div>'
        + '</div>'
        + '<div class="mt-1 text-[12px] leading-tight whitespace-pre-wrap break-words text-green-50">' + escHtml(shortText(item.text)) + '</div>'
        + '<div class="mt-1 pt-1 border-t border-dashed border-[#315f47] text-[10px] text-pine-300 flex gap-2 flex-wrap"><span>from ' + escHtml(item.source) + '</span><span>' + new Date(item.updated_at).toLocaleString() + '</span></div>'
      + '</article>';
    }).join('');

    host.querySelectorAll('.copy-btn').forEach(function(btn){
      btn.addEventListener('click', function(){
        const id = Number(btn.getAttribute('data-id'));
        const item = currentItems.find(function(x){ return x.id === id; });
        if (item) copyText(item.text || '', btn);
      });
    });

    host.querySelectorAll('.pin-btn').forEach(function(btn){
      btn.addEventListener('click', async function(){
        const id = Number(btn.getAttribute('data-id'));
        const pinned = btn.getAttribute('data-pinned') === 'true';
        try {
          await updatePin(id, !pinned);
          await loadHistory();
        } catch (e) {
          document.getElementById('sendStatus').textContent = 'Pin update failed.';
        }
      });
    });
  } catch (err) {
    host.innerHTML = '<div class="text-[11px] text-pine-300 text-center p-2 border border-dashed border-[#467c5f] rounded-lg">Failed to load history.</div>';
  }
}

document.getElementById('sendBtn').addEventListener('click', async function(){
  if (sending) return;
  const text = document.getElementById('clipInput').value.trim();
  if (!text) {
    document.getElementById('sendStatus').textContent = 'Please enter text first.';
    return;
  }

  sending = true;
  const statusEl = document.getElementById('sendStatus');
  statusEl.textContent = 'Saving...';

  try {
    const res = await fetch('/api/clipboard', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ text: text, source: 'mobile-web' })
    });
    if (!res.ok) {
      const msg = await res.text();
      statusEl.textContent = 'Send failed: ' + (msg || res.statusText);
      return;
    }
    statusEl.textContent = 'Saved.';
    document.getElementById('clipInput').value = '';
    await loadLatest();
    await loadHistory();
  } catch (err) {
    statusEl.textContent = 'Send failed: network error';
  } finally {
    sending = false;
  }
});

document.getElementById('clearBtn').addEventListener('click', function(){
  document.getElementById('clipInput').value = '';
});

document.getElementById('searchBtn').addEventListener('click', async function(){
  currentQuery = document.getElementById('searchInput').value.trim();
  await loadHistory();
});

document.getElementById('searchInput').addEventListener('keydown', async function(e){
  if (e.key === 'Enter') {
    currentQuery = document.getElementById('searchInput').value.trim();
    await loadHistory();
  }
});

loadLatest();
loadHistory();
setInterval(function(){ loadLatest(); }, 3500);
</script>
</body>
</html>`
