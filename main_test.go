package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClipboardStoreSetAndGet(t *testing.T) {
	store := &clipboardStore{}
	entry := clipboardUpdate{Text: "hello", Source: "test"}
	store.set(entry)
	latest := store.get()
	if latest.Text != "hello" || latest.Source != "test" {
		t.Fatalf("unexpected latest: %+v", latest)
	}
}

func TestAPIClipboardValidation(t *testing.T) {
	h := &sqliteHistory{path: t.TempDir() + "/test.db"}
	if err := h.init(); err != nil {
		t.Fatal(err)
	}
	a := &app{store: &clipboardStore{}, history: h}

	req := httptest.NewRequest(http.MethodPost, "/api/clipboard", strings.NewReader(`{"text":""}`))
	rr := httptest.NewRecorder()
	a.handleClipboard(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d", rr.Code)
	}
}

func TestHistoryEndpoint(t *testing.T) {
	h := &sqliteHistory{path: t.TempDir() + "/test.db"}
	if err := h.init(); err != nil {
		t.Fatal(err)
	}
	a := &app{store: &clipboardStore{}, history: h}

	_, _ = h.insert("one", "src1")
	_, _ = h.insert("two", "src2")

	req := httptest.NewRequest(http.MethodGet, "/api/history?limit=10", nil)
	rr := httptest.NewRecorder()
	a.handleHistory(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}
	var got []clipboardUpdate
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Text != "two" {
		t.Fatalf("unexpected history: %+v", got)
	}
}
