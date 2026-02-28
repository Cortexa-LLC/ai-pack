package kgclient

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

// makeLog writes a temporary execution.log with the given content and returns
// its path. The caller is responsible for cleanup via t.Cleanup.
func makeLog(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "execution.log")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to create test log: %v", err)
	}
	return path
}

// ---------------------------------------------------------------------------
// parseExecutionLog — unit tests
// ---------------------------------------------------------------------------

func TestParseExecutionLog_Empty(t *testing.T) {
	t.Parallel()
	path := makeLog(t, "")
	got, err := parseExecutionLog(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Turns != 0 {
		t.Errorf("turns: want 0, got %d", got.Turns)
	}
	if len(got.ToolCounts) != 0 {
		t.Errorf("ToolCounts: want empty, got %v", got.ToolCounts)
	}
	if len(got.FilesTouched) != 0 {
		t.Errorf("FilesTouched: want empty, got %v", got.FilesTouched)
	}
}

func TestParseExecutionLog_MissingFile(t *testing.T) {
	t.Parallel()
	_, err := parseExecutionLog("/nonexistent/path/execution.log")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

const typicalLog = `[13:07:21] 🚀 Agent execution started
[13:07:21]    Role: engineer
[13:07:21]    Task: Fix something important
[13:07:21] 🔄 Starting agentic loop (max_inactive: 10, max_consec_errors: 10, caching: enabled, extended_thinking: false)
[13:07:22]    Turn 1 (inactive: 0)...
[13:07:24]       API: 2000ms | in:5000 out:40 | cumulative:5040
[13:07:24]       🔧 Tool: Read(internal/server/foo.go)
[13:07:24]          ✓ package server...
[13:07:24]       🔧 Tool: Bash(cd /repo && go build ./...)
[13:07:25]          ✓ ok
[13:07:25]    Turn 2 (inactive: 0)...
[13:07:30]       API: 5000ms | in:12000 out:80 | cumulative:17040
[13:07:30]       🔧 Tool: Write(internal/server/bar.go)
[13:07:30]          ✓ ok
[13:07:30]       🔧 Tool: Edit(internal/server/baz.go)
[13:07:30]          ✓ ok
[13:07:30]       🔧 Tool: Bash(go test ./...)
[13:07:31]          ✓ PASS
[13:07:31]       🔧 Tool: Read(internal/server/foo.go)
[13:07:31]          ✓ package server...
[13:07:31]    Turn 3 (inactive: 0)...
[13:07:35]       API: 4000ms | in:15000 out:120 | cumulative:32040
[13:07:35]       🔧 Tool: TaskComplete(summary: fixed)
[13:07:35]    ✅ TaskComplete called
`

func TestParseExecutionLog_ToolCounts(t *testing.T) {
	t.Parallel()
	path := makeLog(t, typicalLog)
	got, err := parseExecutionLog(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ToolCounts["Read"] != 2 {
		t.Errorf("Read count: want 2, got %d", got.ToolCounts["Read"])
	}
	if got.ToolCounts["Bash"] != 2 {
		t.Errorf("Bash count: want 2, got %d", got.ToolCounts["Bash"])
	}
	if got.ToolCounts["Write"] != 1 {
		t.Errorf("Write count: want 1, got %d", got.ToolCounts["Write"])
	}
	if got.ToolCounts["Edit"] != 1 {
		t.Errorf("Edit count: want 1, got %d", got.ToolCounts["Edit"])
	}
	if got.ToolCounts["TaskComplete"] != 1 {
		t.Errorf("TaskComplete count: want 1, got %d", got.ToolCounts["TaskComplete"])
	}
}

func TestParseExecutionLog_Turns(t *testing.T) {
	t.Parallel()
	path := makeLog(t, typicalLog)
	got, err := parseExecutionLog(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Turns != 3 {
		t.Errorf("turns: want 3, got %d", got.Turns)
	}
}

func TestParseExecutionLog_TotalTokens(t *testing.T) {
	t.Parallel()
	path := makeLog(t, typicalLog)
	got, err := parseExecutionLog(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.TotalTokens != 32040 {
		t.Errorf("total_tokens: want 32040, got %d", got.TotalTokens)
	}
}

func TestParseExecutionLog_FilesTouched(t *testing.T) {
	t.Parallel()
	path := makeLog(t, typicalLog)
	got, err := parseExecutionLog(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{
		"internal/server/bar.go",
		"internal/server/baz.go",
		"internal/server/foo.go",
	}
	if !slices.Equal(got.FilesTouched, want) {
		t.Errorf("FilesTouched:\n  want: %v\n  got:  %v", want, got.FilesTouched)
	}
}

func TestParseExecutionLog_NoErrors(t *testing.T) {
	t.Parallel()
	path := makeLog(t, typicalLog)
	got, err := parseExecutionLog(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The "PASS" / "FAIL" words inside tool results shouldn't trip the error flag
	// here since the typical log doesn't actually contain a failure.
	// (HasErrors may be true because "PASS" line contains no trigger, but check
	// the log: it doesn't contain "error", "fail", "panic", or "❌" outside of
	// code context lines.)
	if got.HasErrors {
		t.Errorf("HasErrors: want false, got true")
	}
}

const errorLog = `[13:07:21] 🚀 Agent execution started
[13:07:21]    Turn 1 (inactive: 0)...
[13:07:22]       🔧 Tool: Bash(go test ./...)
[13:07:22]          ✗ FAIL: TestFoo (exit status 1)
[13:07:22]       API: 1000ms | in:1000 out:10 | cumulative:1010
`

func TestParseExecutionLog_WithErrors(t *testing.T) {
	t.Parallel()
	path := makeLog(t, errorLog)
	got, err := parseExecutionLog(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.HasErrors {
		t.Errorf("HasErrors: want true, got false")
	}
}

const errNilLog = `[13:07:21] 🚀 Agent execution started
[13:07:21]    Turn 1 (inactive: 0)...
[13:07:22]       🔧 Tool: Read(src/main.go)
[13:07:22]          ✓ if err != nil { return err }
[13:07:22]       API: 500ms | in:500 out:5 | cumulative:505
`

func TestParseExecutionLog_ErrNilCodeExcluded(t *testing.T) {
	t.Parallel()
	// Lines containing "err != nil" are common code patterns and should NOT
	// set HasErrors to true.
	path := makeLog(t, errNilLog)
	got, err := parseExecutionLog(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.HasErrors {
		t.Errorf("HasErrors: want false for code-context error lines, got true")
	}
}

// ---------------------------------------------------------------------------
// buildLogObservations — unit tests
// ---------------------------------------------------------------------------

func TestBuildLogObservations_ContainsRole(t *testing.T) {
	t.Parallel()
	parsed := &ParsedLog{
		Turns:      3,
		ToolCounts: map[string]int{"Bash": 2, "Read": 1},
		TotalTokens: 5000,
	}
	start := time.Now().Add(-2 * time.Minute)
	obs := buildLogObservations(parsed, "engineer", "Fix the bug", start, time.Now())

	hasRole := false
	hasTask := false
	hasDuration := false
	hasTurns := false
	hasTokens := false
	hasTools := false
	for _, o := range obs {
		switch {
		case o == "role: engineer":
			hasRole = true
		case len(o) > 6 && o[:6] == "task: ":
			hasTask = true
		case len(o) > 10 && o[:10] == "duration: ":
			hasDuration = true
		case len(o) > 7 && o[:7] == "turns: ":
			hasTurns = true
		case len(o) > 13 && o[:13] == "total_tokens:":
			hasTokens = true
		case len(o) > 11 && o[:11] == "tool_calls:":
			hasTools = true
		}
	}
	if !hasRole {
		t.Errorf("missing role observation: %v", obs)
	}
	if !hasTask {
		t.Errorf("missing task observation: %v", obs)
	}
	if !hasDuration {
		t.Errorf("missing duration observation: %v", obs)
	}
	if !hasTurns {
		t.Errorf("missing turns observation: %v", obs)
	}
	if !hasTokens {
		t.Errorf("missing total_tokens observation: %v", obs)
	}
	if !hasTools {
		t.Errorf("missing tool_calls observation: %v", obs)
	}
}

func TestBuildLogObservations_TaskDescTruncated(t *testing.T) {
	t.Parallel()
	longDesc := string(make([]rune, 400))
	for i := range longDesc {
		longDesc = longDesc[:i] + "x" + longDesc[i+1:]
	}
	parsed := &ParsedLog{ToolCounts: map[string]int{}}
	obs := buildLogObservations(parsed, "engineer", longDesc, time.Now(), time.Now())
	for _, o := range obs {
		if len(o) > 7 && o[:6] == "task: " {
			// task line must not exceed 300 + len("task: ") + 3 (ellipsis)
			const maxRunes = 6 + 300 + 3 // "task: " + 300 chars + "…"
			if len([]rune(o)) > maxRunes {
				t.Errorf("task observation too long: %d runes", len([]rune(o)))
			}
		}
	}
}

func TestBuildLogObservations_ErrorFlag(t *testing.T) {
	t.Parallel()
	parsed := &ParsedLog{
		ToolCounts: map[string]int{},
		HasErrors:  true,
	}
	obs := buildLogObservations(parsed, "engineer", "task", time.Now(), time.Now())
	found := false
	for _, o := range obs {
		if o == "had_errors: true" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'had_errors: true' in observations, got: %v", obs)
	}
}

func TestBuildLogObservations_NoErrorFlag(t *testing.T) {
	t.Parallel()
	parsed := &ParsedLog{ToolCounts: map[string]int{}, HasErrors: false}
	obs := buildLogObservations(parsed, "engineer", "task", time.Now(), time.Now())
	for _, o := range obs {
		if o == "had_errors: true" {
			t.Errorf("unexpected 'had_errors: true' when HasErrors=false: %v", obs)
		}
	}
}

// ---------------------------------------------------------------------------
// IndexExecutionLog — smoke test (nil manager must not panic)
// ---------------------------------------------------------------------------

func TestIndexExecutionLog_NilManager(t *testing.T) {
	t.Parallel()
	// Nil manager must be a silent no-op.
	dir := t.TempDir()
	IndexExecutionLog(
		context.Background(),
		nil,
		dir,
		"task-abc123",
		"engineer",
		"some task",
		time.Now(),
	)
}
