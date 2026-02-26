package kgclient

import (
	"slices"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// ExtractFilePaths
// ---------------------------------------------------------------------------

func TestExtractFilePaths_Empty(t *testing.T) {
	t.Parallel()
	if got := ExtractFilePaths(""); len(got) != 0 {
		t.Fatalf("expected empty slice, got %v", got)
	}
}

func TestExtractFilePaths_NoFiles(t *testing.T) {
	t.Parallel()
	text := "The task is complete. All good. No code was changed."
	if got := ExtractFilePaths(text); len(got) != 0 {
		t.Fatalf("expected no paths, got %v", got)
	}
}

func TestExtractFilePaths_AbsolutePaths(t *testing.T) {
	t.Parallel()
	text := "modified /home/user/project/internal/foo/bar.go and /tmp/out.txt"
	got := ExtractFilePaths(text)
	if !slices.Contains(got, "/home/user/project/internal/foo/bar.go") {
		t.Errorf("expected absolute path, got %v", got)
	}
}

func TestExtractFilePaths_RelativePaths(t *testing.T) {
	t.Parallel()
	text := `
Files changed:
  internal/server/task_lifecycle.go
  internal/kgclient/writeback.go
`
	got := ExtractFilePaths(text)
	if !slices.Contains(got, "internal/server/task_lifecycle.go") {
		t.Errorf("expected relative path in %v", got)
	}
	if !slices.Contains(got, "internal/kgclient/writeback.go") {
		t.Errorf("expected relative path in %v", got)
	}
}

func TestExtractFilePaths_Deduplication(t *testing.T) {
	t.Parallel()
	text := "touched internal/foo/bar.go and also internal/foo/bar.go again"
	got := ExtractFilePaths(text)
	count := 0
	for _, p := range got {
		if p == "internal/foo/bar.go" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected exactly 1 occurrence of path, got %d in %v", count, got)
	}
}

func TestExtractFilePaths_GitDiffOutput(t *testing.T) {
	t.Parallel()
	text := `
 internal/server/task_lifecycle.go | 10 ++++------
 internal/kgclient/writeback.go    |  5 +++++
`
	got := ExtractFilePaths(text)
	if !slices.Contains(got, "internal/server/task_lifecycle.go") {
		t.Errorf("expected task_lifecycle.go in %v", got)
	}
	if !slices.Contains(got, "internal/kgclient/writeback.go") {
		t.Errorf("expected writeback.go in %v", got)
	}
}

// ---------------------------------------------------------------------------
// buildObservations
// ---------------------------------------------------------------------------

func TestBuildObservations_ContainsRequiredKeys(t *testing.T) {
	t.Parallel()
	start := time.Now().Add(-2 * time.Minute)
	end := time.Now()
	output := "Edited internal/server/task_lifecycle.go to wire WriteBack"

	obs := buildObservations("engineer", "Implement KG write-back", output, start, end)

	has := func(prefix string) bool {
		for _, o := range obs {
			if len(o) >= len(prefix) && o[:len(prefix)] == prefix {
				return true
			}
		}
		return false
	}

	for _, prefix := range []string{"role: ", "duration: ", "task: ", "changed_files: "} {
		if !has(prefix) {
			t.Errorf("missing observation with prefix %q in %v", prefix, obs)
		}
	}
}

func TestBuildObservations_LongTaskDescTruncated(t *testing.T) {
	t.Parallel()
	longDesc := string(make([]byte, 400))
	for i := range longDesc {
		longDesc = longDesc[:i] + "x" + longDesc[i+1:]
	}
	obs := buildObservations("engineer", longDesc, "", time.Now().Add(-1*time.Minute), time.Now())

	// "task: " (6) + 300 body chars + 3-byte UTF-8 "…" = 309 bytes max
	const maxLen = 6 + 300 + 3 // len("task: ") + maxSummaryChars + len("…")
	for _, o := range obs {
		if len(o) > maxLen {
			t.Errorf("task observation not truncated: len=%d (max %d)", len(o), maxLen)
		}
	}
}

func TestBuildObservations_EmptyOutputNoChangedFiles(t *testing.T) {
	t.Parallel()
	obs := buildObservations("planner", "Do something", "", time.Now(), time.Now())

	for _, o := range obs {
		if len(o) > 15 && o[:14] == "changed_files:" {
			t.Errorf("unexpected changed_files observation with empty output: %s", o)
		}
	}
}

func TestBuildObservations_DurationPresent(t *testing.T) {
	t.Parallel()
	start := time.Now().Add(-90 * time.Second)
	end := time.Now()

	obs := buildObservations("tester", "Run tests", "", start, end)

	found := false
	for _, o := range obs {
		if len(o) > 10 && o[:10] == "duration: " {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("duration observation not found in %v", obs)
	}
}

// ---------------------------------------------------------------------------
// WriteBack (nil manager is a no-op — just verify it does not panic)
// ---------------------------------------------------------------------------

func TestWriteBack_NilManager(t *testing.T) {
	t.Parallel()
	// Must not panic
	WriteBack(nil, nil, "/project", "engineer", "do stuff", "some output", time.Now().Add(-1*time.Minute))
}
