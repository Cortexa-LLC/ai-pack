package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAddAssignsSequentialIDs(t *testing.T) {
	s := New()
	first := s.Add(Record{Name: "Blue Train", Artist: "John Coltrane"})
	second := s.Add(Record{Name: "Giant Steps", Artist: "John Coltrane"})
	if first != 1 || second != 2 {
		t.Fatalf("want IDs 1,2; got %d,%d", first, second)
	}
}

func TestScoreIsMeanOfRatings(t *testing.T) {
	r := Record{Ratings: []float64{4, 5}}
	if got := r.Score(); got != 4.5 {
		t.Fatalf("want 4.5, got %g", got)
	}
	if got := (Record{}).Score(); got != 0 {
		t.Fatalf("want 0 for unrated record, got %g", got)
	}
}

func TestNormalizeFillsMissingArtist(t *testing.T) {
	r := Normalize(Record{Name: "  Untitled  ", Artist: " "})
	if r.Name != "Untitled" || r.Artist != "Unknown Artist" {
		t.Fatalf("unexpected normalization: %+v", r)
	}
}

func TestSearchFindsByArtist(t *testing.T) {
	s := New()
	s.Add(Record{Name: "Blue Train", Artist: "John Coltrane"})
	s.Add(Record{Name: "Giant Steps", Artist: "John Coltrane"})
	if got := s.Search("coltrane", 0); len(got) != 2 {
		t.Fatalf("want 2 matches, got %d", len(got))
	}
}

func TestAppendAndLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.db")
	s := New()
	id := s.Add(Record{Name: "Kind of Blue", Artist: "Miles Davis", Ratings: []float64{5, 4.5}})
	rec, _ := s.Get(id)
	if err := Append(path, rec); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := loaded.Get(id)
	if !ok || got.Name != "Kind of Blue" || len(got.Ratings) != 2 {
		t.Fatalf("round trip lost data: %+v (ok=%v)", got, ok)
	}
}

// TestCompact covers the journal rewrite path end to end.
func TestCompact(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.db")
	recs := []Record{
		{ID: 1, Name: "Blue Train", Artist: "John Coltrane", Ratings: []float64{5}},
		{ID: 2, Name: "Giant Steps", Artist: "John Coltrane", Ratings: []float64{4.5}},
		{ID: 3, Name: "Kind of Blue", Artist: "Miles Davis", Ratings: []float64{5}},
	}
	for _, r := range recs {
		if err := Append(path, r); err != nil {
			t.Fatal(err)
		}
	}
	if err := Compact(path); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path + ".log")
	if err != nil {
		t.Fatalf("compaction log not written: %v", err)
	}
	if !strings.Contains(string(data), "compacted") {
		t.Fatalf("compaction pass not recorded: %q", data)
	}
}
