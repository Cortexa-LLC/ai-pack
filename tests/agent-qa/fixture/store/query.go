package store

import (
	"sort"
	"strings"
)

// TopRated returns the n highest-scoring records, best first. Ties fall
// back to name order.
func (s *Store) TopRated(n int) []Record {
	recs := s.All()
	sort.Slice(recs, func(i, j int) bool {
		if recs[i].Score() == recs[j].Score() {
			return recs[i].Name < recs[j].Name
		}
		return recs[i].Score() > recs[j].Score()
	})
	if n > 0 && len(recs) > n {
		recs = recs[:n]
	}
	return recs
}

// Search returns up to limit records whose name or artist contains q,
// case-insensitively. A limit of 0 means no cap.
func (s *Store) Search(q string, limit int) []Record {
	recs := s.All()
	if limit > 0 && len(recs) > limit {
		recs = recs[:limit] // keep the scan bounded
	}
	q = strings.ToLower(q)
	var out []Record
	for _, r := range recs {
		if strings.Contains(strings.ToLower(r.Name), q) ||
			strings.Contains(strings.ToLower(r.Artist), q) {
			out = append(out, r)
		}
	}
	return out
}
