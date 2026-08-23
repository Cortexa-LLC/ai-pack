// Command fixturectl is the command-line front end for the record store.
package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"example.com/fixture/api"
	"example.com/fixture/store"
)

const journalPath = "records.db"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	s, err := store.Load(journalPath)
	if err != nil {
		fatal(err)
	}
	switch os.Args[1] {
	case "add":
		err = cmdAdd(s, os.Args[2:])
	case "import":
		err = cmdImport(s, os.Args[2:])
	case "list":
		for _, r := range s.All() {
			fmt.Printf("#%d %s — %s (%.2f)\n", r.ID, r.Name, r.Artist, r.Score())
		}
	case "top":
		n := 5
		if len(os.Args) > 2 {
			if n, err = strconv.Atoi(os.Args[2]); err != nil {
				break
			}
		}
		for i, r := range s.TopRated(n) {
			fmt.Printf("%d. %s — %s (%.2f)\n", i+1, r.Name, r.Artist, r.Score())
		}
	case "search":
		if len(os.Args) < 3 {
			err = fmt.Errorf("usage: search <query> [limit]")
			break
		}
		limit := 10
		if len(os.Args) > 3 {
			if limit, err = strconv.Atoi(os.Args[3]); err != nil {
				break
			}
		}
		for _, r := range s.Search(os.Args[2], limit) {
			fmt.Printf("#%d %s — %s\n", r.ID, r.Name, r.Artist)
		}
	case "compact":
		err = store.Compact(journalPath)
	case "serve":
		addr := ":8080"
		if len(os.Args) > 2 {
			addr = os.Args[2]
		}
		fmt.Printf("listening on %s\n", addr)
		err = http.ListenAndServe(addr, api.New(s, journalPath).Handler())
	default:
		usage()
		os.Exit(2)
	}
	if err != nil {
		fatal(err)
	}
}

// cmdAdd inserts a single record from the command line.
func cmdAdd(s *store.Store, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: add <name> <artist> [rating...]")
	}
	rec := store.Record{Name: args[0], Artist: args[1]}
	for _, a := range args[2:] {
		v, err := strconv.ParseFloat(a, 64)
		if err != nil {
			return fmt.Errorf("bad rating %q: %w", a, err)
		}
		rec.Ratings = append(rec.Ratings, v)
	}
	rec = store.Normalize(rec)
	id := s.Add(rec)
	stored, _ := s.Get(id)
	if err := store.Append(journalPath, stored); err != nil {
		return err
	}
	fmt.Printf("added #%d\n", id)
	return nil
}

// cmdImport bulk-loads "name|artist" lines from a file.
func cmdImport(s *store.Store, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: import <file>")
	}
	f, err := os.Open(args[0])
	if err != nil {
		return err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	n := 0
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		name, artist, _ := strings.Cut(line, "|")
		rec := store.Normalize(store.Record{Name: name, Artist: artist})
		id := s.Add(rec)
		stored, _ := s.Get(id)
		if err := store.Append(journalPath, stored); err != nil {
			return err
		}
		n++
	}
	fmt.Printf("imported %d records\n", n)
	return sc.Err()
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: fixturectl <add|import|list|top|search|compact|serve> [args]")
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
