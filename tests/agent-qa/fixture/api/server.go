// Package api exposes the record store over HTTP.
package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"example.com/fixture/store"
)

// Server wires the store to HTTP handlers.
type Server struct {
	store   *store.Store
	journal string
}

// New returns a Server backed by s, persisting to journalPath.
func New(s *store.Store, journalPath string) *Server {
	return &Server{store: s, journal: journalPath}
}

// Handler returns the API route table.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /records", s.handleCreate)
	mux.HandleFunc("GET /records", s.handleList)
	mux.HandleFunc("GET /search", s.handleSearch)
	return mux
}

// handleCreate stores a record supplied as JSON and persists it to the
// journal.
func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	var rec store.Record
	if err := json.NewDecoder(r.Body).Decode(&rec); err != nil {
		http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
		return
	}
	if rec.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	id := s.store.Add(rec)
	stored, _ := s.store.Get(id)
	if err := s.persistWithRetry(stored); err != nil {
		http.Error(w, "persist: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"id": id})
}

// persistWithRetry appends rec to the journal, retrying once so a transient
// filesystem hiccup does not surface as a 500.
func (s *Server) persistWithRetry(rec store.Record) error {
	if err := store.Append(s.journal, rec); err != nil {
		return store.Append(s.journal, rec)
	}
	return nil
}

// handleList returns every record in the catalog.
func (s *Server) handleList(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.store.All())
}

// handleSearch returns records matching ?q=, capped by ?limit=.
func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	limit := 0
	if raw := r.URL.Query().Get("limit"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 0 {
			http.Error(w, "bad limit", http.StatusBadRequest)
			return
		}
		limit = v
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.store.Search(q, limit))
}
