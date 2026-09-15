package main

import (
	"sync"
	"time"
)

type Note struct {
	Ciphertext []byte
	ExpiresAt  time.Time
}

type Store struct {
	mu    sync.Mutex
	notes map[string]Note
}

func NewStore() *Store {
	s := &Store{notes: make(map[string]Note)}
	go s.cleanupLoop()
	return s
}

func (s *Store) Put(id string, ciphertext []byte, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.notes[id] = Note{
		Ciphertext: ciphertext,
		ExpiresAt:  time.Now().Add(ttl),
	}
}

// GetAndDelete atomically retrieves and removes a note.
// Returns nil, false if missing or expired.
func (s *Store) GetAndDelete(id string) ([]byte, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	note, ok := s.notes[id]
	if !ok {
		return nil, false
	}
	delete(s.notes, id)

	if time.Now().After(note.ExpiresAt) {
		return nil, false
	}
	return note.Ciphertext, true
}

func (s *Store) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.notes)
}

func (s *Store) cleanupLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		s.cleanup()
	}
}

func (s *Store) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for id, note := range s.notes {
		if now.After(note.ExpiresAt) {
			delete(s.notes, id)
		}
	}
}
