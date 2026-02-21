package server

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func respondJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func respondError(w http.ResponseWriter, msg string, status int) {
	http.Error(w, msg, status)
}

func sanitizeForDB(s string) string {
	s = strings.ReplaceAll(s, "\x00", "")
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return s
}

func (a *App) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(indexHTML)
}

func (a *App) handleClipboard(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var text, source string
		ct := r.Header.Get("Content-Type")
		if strings.HasPrefix(ct, "application/x-www-form-urlencoded") {
			if err := r.ParseForm(); err != nil {
				respondError(w, "invalid form body", http.StatusBadRequest)
				return
			}
			text = r.FormValue("text")
			source = r.FormValue("source")
		} else {
			var req struct {
				Text   string `json:"text"`
				Source string `json:"source"`
			}
			body, _ := io.ReadAll(r.Body)
			r.Body.Close()
			if err := json.Unmarshal(body, &req); err != nil {
				respondError(w, "invalid JSON body", http.StatusBadRequest)
				return
			}
			text = req.Text
			source = req.Source
		}
		text = strings.TrimSpace(text)
		if text == "" {
			respondError(w, "text is required", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(source) == "" {
			source = "unknown"
		}
		text = sanitizeForDB(text)
		source = sanitizeForDB(source)
		entry, err := a.History.Insert(text, source)
		if err != nil {
			log.Printf("clipboard insert failed: %v", err)
			respondError(w, "failed to save clipboard", http.StatusInternalServerError)
			return
		}
		a.Store.Set(entry)
		respondJSON(w, http.StatusCreated, entry)
	case http.MethodGet:
		latest := a.Store.Get()
		if strings.TrimSpace(latest.Text) == "" {
			respondError(w, "clipboard is empty", http.StatusNotFound)
			return
		}
		respondJSON(w, http.StatusOK, latest)
	default:
		respondError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (a *App) handleHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	limit := 50
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > 200 {
			respondError(w, "invalid limit", http.StatusBadRequest)
			return
		}
		limit = parsed
	}
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	items, err := a.History.List(limit, query)
	if err != nil {
		respondError(w, "failed to read history", http.StatusInternalServerError)
		return
	}
	respondJSON(w, http.StatusOK, items)
}

func (a *App) handlePin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		ID     int64 `json:"id"`
		Pinned bool  `json:"pinned"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if req.ID <= 0 {
		respondError(w, "id is required", http.StatusBadRequest)
		return
	}
	if err := a.History.SetPinned(req.ID, req.Pinned); err != nil {
		respondError(w, "failed to update pin", http.StatusInternalServerError)
		return
	}
	entry, err := a.History.ByID(req.ID)
	if err != nil {
		respondError(w, "entry not found", http.StatusNotFound)
		return
	}
	respondJSON(w, http.StatusOK, entry)
}

func (a *App) handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		respondError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if req.ID <= 0 {
		respondError(w, "id is required", http.StatusBadRequest)
		return
	}
	if err := a.History.Delete(req.ID); err != nil {
		respondError(w, "failed to delete", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if a.Logs == nil {
		respondJSON(w, http.StatusOK, []interface{}{})
		return
	}
	respondJSON(w, http.StatusOK, a.Logs.List())
}
