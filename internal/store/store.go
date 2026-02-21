package store

import (
	"local-clipboard/internal/models"
	"sync"
)

// Store holds the latest clipboard value in memory.
type Store struct {
	mu     sync.RWMutex
	latest models.ClipboardUpdate
}

// New returns a new Store.
func New() *Store {
	return &Store{}
}

// Set updates the latest clipboard value.
func (s *Store) Set(v models.ClipboardUpdate) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.latest = v
}

// Get returns the latest clipboard value.
func (s *Store) Get() models.ClipboardUpdate {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.latest
}
