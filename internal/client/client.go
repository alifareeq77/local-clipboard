package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"local-clipboard/internal/clipboard"
	"local-clipboard/internal/models"
)

// Config holds client options.
type Config struct {
	ServerURL string
	Interval  time.Duration
	Source    string
}

// Run runs the clipboard client: poll local clipboard, push to server, optionally pull remote.
func Run(cfg Config) {
	localRead, localWrite, err := clipboard.Detect()
	if err != nil {
		log.Fatalf("clipboard command setup failed: %v", err)
	}
	if localWrite == nil {
		log.Printf("note: no clipboard write command found; this client will only push local copy events")
	}

	baseURL := strings.TrimRight(cfg.ServerURL, "/")
	var lastSent string
	for {
		text, err := clipboard.Read(localRead)
		if err != nil {
			time.Sleep(cfg.Interval)
			continue
		}
		text = strings.TrimSpace(text)
		if text != "" && text != lastSent {
			if err := PostClipboard(baseURL, text, cfg.Source); err == nil {
				lastSent = text
			}
		}

		if localWrite != nil {
			remote, err := FetchClipboard(baseURL)
			if err == nil && remote.Text != "" && remote.Source != cfg.Source && remote.Text != text {
				if err := clipboard.Write(localWrite, remote.Text); err == nil {
					lastSent = remote.Text
				}
			}
		}
		time.Sleep(cfg.Interval)
	}
}

// PostClipboard sends text to the server clipboard API.
func PostClipboard(baseURL, text, source string) error {
	payload := map[string]string{"text": text, "source": source}
	b, _ := json.Marshal(payload)
	resp, err := http.Post(baseURL+"/api/clipboard", "application/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("unexpected status %s", resp.Status)
	}
	return nil
}

// FetchClipboard returns the current clipboard from the server.
func FetchClipboard(baseURL string) (models.ClipboardUpdate, error) {
	resp, err := http.Get(baseURL + "/api/clipboard")
	if err != nil {
		return models.ClipboardUpdate{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return models.ClipboardUpdate{}, fmt.Errorf("status %s", resp.Status)
	}
	var out models.ClipboardUpdate
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return models.ClipboardUpdate{}, err
	}
	return out, nil
}

// HostName returns the machine hostname for use as source, or "linux-client" if unavailable.
func HostName() string {
	h, err := os.Hostname()
	if err != nil || strings.TrimSpace(h) == "" {
		return "linux-client"
	}
	return h
}
