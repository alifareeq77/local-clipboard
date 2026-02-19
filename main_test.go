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
	entry := clipboardUpdate{Text: "hello", Source: "test", Pinned: true}
	store.set(entry)
	latest := store.get()
	if latest.Text != "hello" || latest.Source != "test" || !latest.Pinned {
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

func TestHistorySearchAndPin(t *testing.T) {
	h := &sqliteHistory{path: t.TempDir() + "/test.db"}
	if err := h.init(); err != nil {
		t.Fatal(err)
	}
	a := &app{store: &clipboardStore{}, history: h}

	first, _ := h.insert("alpha snippet", "src1")
	_, _ = h.insert("beta note", "src2")
	if err := h.setPinned(first.ID, true); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/history?limit=10&q=alpha", nil)
	rr := httptest.NewRecorder()
	a.handleHistory(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}

	var got []clipboardUpdate
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Text != "alpha snippet" || !got[0].Pinned {
		t.Fatalf("unexpected search results: %+v", got)
	}
}
