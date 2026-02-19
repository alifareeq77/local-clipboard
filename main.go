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
	"strings"
	"sync"
	"time"
)

type clipboardUpdate struct {
	Text      string    `json:"text"`
	Source    string    `json:"source"`
	UpdatedAt time.Time `json:"updated_at"`
}

type clipboardStore struct {
	mu     sync.RWMutex
	latest clipboardUpdate
}

func (s *clipboardStore) set(text, source string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.latest = clipboardUpdate{
		Text:      text,
		Source:    source,
		UpdatedAt: time.Now().UTC(),
	}
}

func (s *clipboardStore) get() clipboardUpdate {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.latest
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
	_ = fs.Parse(args)

	store := &clipboardStore{}
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(indexHTML))
	})

	mux.HandleFunc("/api/clipboard", func(w http.ResponseWriter, r *http.Request) {
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
			store.set(req.Text, req.Source)
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(store.get())
		case http.MethodGet:
			latest := store.get()
			if strings.TrimSpace(latest.Text) == "" {
				http.Error(w, "clipboard is empty", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(latest)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Printf("clipboard server listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
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

	log.Printf("watching local clipboard every %s and pushing updates to %s", interval.String(), *serverURL)

	var lastSent string
	for {
		text, err := runClipboardRead(localRead)
		if err != nil {
			log.Printf("read clipboard error: %v", err)
			time.Sleep(*interval)
			continue
		}
		text = strings.TrimSpace(text)
		if text != "" && text != lastSent {
			if err := postClipboard(*serverURL, text, *source); err != nil {
				log.Printf("push clipboard error: %v", err)
			} else {
				lastSent = text
				log.Printf("uploaded clipboard (%d bytes)", len(text))
			}
		}

		if localWrite != nil {
			if remote, err := fetchClipboard(*serverURL); err == nil && remote.Text != "" && remote.Source != *source {
				if remote.Text != text {
					if err := runClipboardWrite(localWrite, remote.Text); err == nil {
						lastSent = remote.Text
						log.Printf("applied remote clipboard from %s", remote.Source)
					}
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
			write := &clipboardCmd{name: "wl-copy"}
			return clipboardCmd{name: "wl-paste"}, write, nil
		}
		return clipboardCmd{name: "wl-paste"}, nil, nil
	}
	if _, err := exec.LookPath("xclip"); err == nil {
		write := &clipboardCmd{name: "xclip", args: []string{"-selection", "clipboard"}}
		return clipboardCmd{name: "xclip", args: []string{"-o", "-selection", "clipboard"}}, write, nil
	}
	if _, err := exec.LookPath("xsel"); err == nil {
		write := &clipboardCmd{name: "xsel", args: []string{"--clipboard", "--input"}}
		return clipboardCmd{name: "xsel", args: []string{"--clipboard", "--output"}}, write, nil
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

const indexHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>Local Clipboard Bridge</title>
  <style>
    body { font-family: sans-serif; margin: 2rem; max-width: 760px; }
    textarea { width: 100%; min-height: 140px; }
    button { margin-top: 0.8rem; padding: 0.7rem 1.2rem; }
    .card { border: 1px solid #ddd; border-radius: 10px; padding: 1rem; margin-top: 1rem; }
    .muted { color: #666; }
  </style>
</head>
<body>
  <h1>Local Clipboard Bridge</h1>
  <p class="muted">Open this page on your phone and push text into your laptop clipboard server.</p>

  <div class="card">
    <h2>Send from phone</h2>
    <textarea id="clipInput" placeholder="Paste text here"></textarea>
    <button id="sendBtn">Send to server</button>
    <p id="sendStatus" class="muted"></p>
  </div>

  <div class="card">
    <h2>Latest on server</h2>
    <pre id="latest">Loading...</pre>
  </div>

<script>
async function loadLatest() {
  try {
    const res = await fetch('/api/clipboard');
    if (!res.ok) {
      document.getElementById('latest').textContent = 'Clipboard is empty.';
      return;
    }
    const data = await res.json();
    document.getElementById('latest').textContent = JSON.stringify(data, null, 2);
  } catch (err) {
    document.getElementById('latest').textContent = 'Failed to load: ' + err.message;
  }
}

document.getElementById('sendBtn').addEventListener('click', async () => {
  const text = document.getElementById('clipInput').value.trim();
  if (!text) {
    document.getElementById('sendStatus').textContent = 'Please enter text first.';
    return;
  }

  const res = await fetch('/api/clipboard', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ text, source: 'mobile-web' })
  });

  if (!res.ok) {
    document.getElementById('sendStatus').textContent = 'Send failed.';
    return;
  }

  document.getElementById('sendStatus').textContent = 'Sent successfully.';
  await loadLatest();
});

loadLatest();
setInterval(loadLatest, 3000);
</script>
</body>
</html>`
