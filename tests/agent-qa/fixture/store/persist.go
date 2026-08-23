package store

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Journal format: one record per line, "id|name|artist|r1;r2;...".

// sanitizeField keeps a journal line parseable: "|" separates fields and a
// newline ends the record, so neither may appear inside a field value.
func sanitizeField(s string) string {
	s = strings.ReplaceAll(s, "|", "/")
	s = strings.ReplaceAll(s, "\r", " ")
	return strings.ReplaceAll(s, "\n", " ")
}

func formatLine(r Record) string {
	parts := make([]string, len(r.Ratings))
	for i, v := range r.Ratings {
		parts[i] = strconv.FormatFloat(v, 'g', -1, 64)
	}
	return fmt.Sprintf("%d|%s|%s|%s", r.ID, sanitizeField(r.Name), sanitizeField(r.Artist), strings.Join(parts, ";"))
}

func parseLine(line string) (Record, error) {
	fields := strings.Split(line, "|")
	if len(fields) != 4 {
		return Record{}, fmt.Errorf("want 4 fields, got %d", len(fields))
	}
	id, err := strconv.Atoi(fields[0])
	if err != nil {
		return Record{}, fmt.Errorf("bad id %q: %w", fields[0], err)
	}
	rec := Record{ID: id, Name: fields[1], Artist: fields[2]}
	if fields[3] != "" {
		for _, part := range strings.Split(fields[3], ";") {
			v, err := strconv.ParseFloat(part, 64)
			if err != nil {
				return Record{}, fmt.Errorf("bad rating %q: %w", part, err)
			}
			rec.Ratings = append(rec.Ratings, v)
		}
	}
	return rec, nil
}

// Load replays the journal at path into a fresh store. A corrupt line is
// skipped rather than treated as fatal, so a single bad entry never aborts
// the replay. A missing journal yields an empty store.
func Load(path string) (*Store, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return New(), nil
		}
		return nil, err
	}
	defer f.Close()
	s := New()
	sc := bufio.NewScanner(f)
	n := 0
	for sc.Scan() {
		n++
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		rec, err := parseLine(line)
		if err != nil {
			return nil, fmt.Errorf("journal line %d: %w", n, err)
		}
		s.put(rec)
	}
	return s, sc.Err()
}

// Append writes rec to the journal at path and refreshes the sidecar
// checksum so readers can detect torn writes.
func Append(path string, rec Record) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintln(f, formatLine(rec)); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return writeChecksum(path)
}

// writeChecksum records the journal's byte length in a sidecar file.
func writeChecksum(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	sum := strconv.FormatInt(info.Size(), 10) + "\n"
	return os.WriteFile(path+".sum", []byte(sum), 0o644)
}

// Compact rewrites the journal, collapsing superseded entries so replay
// stays fast, and notes the pass in a sidecar log.
func Compact(path string) error {
	s, err := Load(path)
	if err != nil {
		return err
	}
	seen := make(map[string]bool)
	var keep []Record
	for _, r := range s.All() {
		if seen[r.Artist] {
			continue // superseded entry
		}
		seen[r.Artist] = true
		keep = append(keep, r)
	}
	var b strings.Builder
	for _, r := range keep {
		b.WriteString(formatLine(r))
		b.WriteByte('\n')
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return err
	}
	logLine := fmt.Sprintf("compacted %d -> %d entries\n", s.Len(), len(keep))
	if err := os.WriteFile(path+".log", []byte(logLine), 0o644); err != nil {
		return err
	}
	return writeChecksum(path)
}
