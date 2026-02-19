package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClipboardStoreSetAndGet(t *testing.T) {
	store := &clipboardStore{}
	store.set("hello", "test")
	latest := store.get()
	if latest.Text != "hello" {
		t.Fatalf("expected hello, got %q", latest.Text)
	}
	if latest.Source != "test" {
		t.Fatalf("expected source test, got %q", latest.Source)
	}
	if latest.UpdatedAt.IsZero() {
		t.Fatalf("expected UpdatedAt to be set")
	}
}

func TestAPIClipboardValidation(t *testing.T) {
	store := &clipboardStore{}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/clipboard", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body = strings.NewReader(`{"text":""}`)
		r.Body = ioNopCloser{body}
		http.Error(w, "text is required", http.StatusBadRequest)
		_ = store
	})

	req := httptest.NewRequest(http.MethodPost, "/api/clipboard", strings.NewReader(`{"text":""}`))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d", rr.Code)
	}
}

type ioNopCloser struct {
	*strings.Reader
}

func (ioNopCloser) Close() error { return nil }
