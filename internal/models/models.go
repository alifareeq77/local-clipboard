package models

import "time"

// ClipboardUpdate is a single clipboard entry (in-memory or from history).
type ClipboardUpdate struct {
	ID        int64     `json:"id"`
	Text      string    `json:"text"`
	Source    string    `json:"source"`
	UpdatedAt time.Time `json:"updated_at"`
	Pinned    bool      `json:"pinned"`
}
