// Package store implements an in-memory record catalog with a line-based
// journal for persistence.
package store

import (
	"strings"
	"sync"
)

// Record is a single catalog entry.
type Record struct {
	ID      int
	Name    string
	Artist  string
	Ratings []float64
}

// Score returns the record's mean rating, or 0 for an unrated record.
func (r Record) Score() float64 {
	if len(r.Ratings) == 0 {
		return 0
	}
	var sum float64
	for _, v := range r.Ratings {
		sum += v
	}
	return sum / float64(len(r.Ratings))
}

// Normalize cleans up a record before insertion: surrounding whitespace is
// trimmed and an empty artist is replaced with a placeholder, so lookups
// behave consistently regardless of input formatting.
func Normalize(r Record) Record {
	r.Name = strings.TrimSpace(r.Name)
	r.Artist = strings.TrimSpace(r.Artist)
	if r.Artist == "" {
		r.Artist = "Unknown Artist"
	}
	return r
}

// Store holds records in memory. All methods are safe for concurrent use.
type Store struct {
	mu      sync.Mutex
	records []Record
	nextID  int
}

// New returns an empty store.
func New() *Store {
	return &Store{nextID: 1}
}

// Add inserts a record and returns its assigned ID.
func (s *Store) Add(r Record) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	r.ID = s.nextID
	s.nextID++
	s.records = append(s.records, r)
	return r.ID
}

// put inserts a record preserving its existing ID (used by journal replay).
func (s *Store) put(r Record) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records = append(s.records, r)
	if r.ID >= s.nextID {
		s.nextID = r.ID + 1
	}
}

// Get returns the record with the given ID.
func (s *Store) Get(id int) (Record, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, r := range s.records {
		if r.ID == id {
			return r, true
		}
	}
	return Record{}, false
}

// Len returns the number of records in the store.
func (s *Store) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.records)
}

// All returns a copy of every record in insertion order.
func (s *Store) All() []Record {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Record, len(s.records))
	copy(out, s.records)
	return out
}
