package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"
)

const maxNoteSize = 10 * 1024 // 10 KB

var ttlOptions = map[string]time.Duration{
	"5m":    5 * time.Minute,
	"1h":    time.Hour,
	"1d":    24 * time.Hour,
	"once":  7 * 24 * time.Hour, // until first view; hard cap 7 days
}

type createRequest struct {
	Ciphertext string `json:"ciphertext"`
	TTL        string `json:"ttl"`
}

type createResponse struct {
	ID string `json:"id"`
}

type noteResponse struct {
	Ciphertext string `json:"ciphertext"`
}

func (a *App) handleCreateNote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxNoteSize+1024)

	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	if req.Ciphertext == "" {
		http.Error(w, `{"error":"ciphertext required"}`, http.StatusBadRequest)
		return
	}

	raw, err := base64.StdEncoding.DecodeString(req.Ciphertext)
	if err != nil {
		http.Error(w, `{"error":"ciphertext must be base64"}`, http.StatusBadRequest)
		return
	}
	if len(raw) == 0 || len(raw) > maxNoteSize {
		http.Error(w, `{"error":"note too large or empty"}`, http.StatusBadRequest)
		return
	}

	ttl, ok := ttlOptions[req.TTL]
	if !ok {
		ttl = ttlOptions["once"]
	}

	id, err := generateID()
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	a.store.Put(id, raw, ttl)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(createResponse{ID: id})
}

func (a *App) handleGetNote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	id := r.PathValue("id")
	if id == "" || len(id) > 64 {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}

	ciphertext, ok := a.store.GetAndDelete(id)
	if !ok {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(noteResponse{
		Ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
	})
}

func (a *App) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "static/index.html")
}

func (a *App) handleNotePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/note.html")
}

func generateID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
