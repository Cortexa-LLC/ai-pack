package server

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// initGitRepo initialises a git repository in dir with an initial commit so
// that git-diff commands have a valid HEAD to diff against.
func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test",
			"GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=test",
			"GIT_COMMITTER_EMAIL=test@test.com",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init")
	run("config", "user.email", "test@test.com")
	run("config", "user.name", "test")

	// Create an initial commit so HEAD exists.
	placeholder := filepath.Join(dir, ".gitkeep")
	if err := os.WriteFile(placeholder, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	run("add", ".gitkeep")
	run("commit", "-m", "init")
}

// TestDetectOutOfScopeChanges_NoChanges verifies that when no files are
// modified after the last commit, nothing is reported out-of-scope.
func TestDetectOutOfScopeChanges_NoChanges(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)

	outOfScope, err := detectOutOfScopeChanges(dir, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(outOfScope) != 0 {
		t.Errorf("expected no out-of-scope files, got %v", outOfScope)
	}
}

// TestDetectOutOfScopeChanges_InScopeFile verifies that a file modified inside
// the workingDir is NOT reported as out-of-scope.
func TestDetectOutOfScopeChanges_InScopeFile(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)

	// Create a file inside workingDir (a subdirectory of dir) and commit it.
	workDir := filepath.Join(dir, "work")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatal(err)
	}
	f := filepath.Join(workDir, "task.go")
	if err := os.WriteFile(f, []byte("package work\n"), 0644); err != nil {
		t.Fatal(err)
	}

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test",
			"GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=test",
			"GIT_COMMITTER_EMAIL=test@test.com",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("add", "work/task.go")
	run("commit", "-m", "add task.go")

	// Now modify the in-scope file.
	if err := os.WriteFile(f, []byte("package work\n// modified\n"), 0644); err != nil {
		t.Fatal(err)
	}

	outOfScope, err := detectOutOfScopeChanges(dir, workDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(outOfScope) != 0 {
		t.Errorf("expected no out-of-scope files, got %v", outOfScope)
	}
}

// TestDetectOutOfScopeChanges_OutOfScopeFile verifies that a file modified
// outside workingDir IS reported as out-of-scope.
func TestDetectOutOfScopeChanges_OutOfScopeFile(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)

	// Commit a file that lives at the repo root (not inside workDir).
	rootFile := filepath.Join(dir, "entity.go")
	if err := os.WriteFile(rootFile, []byte("package main\n"), 0644); err != nil {
		t.Fatal(err)
	}

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test",
			"GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=test",
			"GIT_COMMITTER_EMAIL=test@test.com",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("add", "entity.go")
	run("commit", "-m", "add entity.go")

	workDir := filepath.Join(dir, "work")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Now the agent modifies entity.go (outside workDir).
	if err := os.WriteFile(rootFile, []byte("package main\n// CORRUPTED\n"), 0644); err != nil {
		t.Fatal(err)
	}

	outOfScope, err := detectOutOfScopeChanges(dir, workDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(outOfScope) == 0 {
		t.Error("expected entity.go to be flagged as out-of-scope, got nothing")
	} else if outOfScope[0] != "entity.go" {
		t.Errorf("expected entity.go, got %v", outOfScope)
	}
}

// TestDetectOutOfScopeChanges_EmptyInputs verifies that empty paths return
// nil without error (graceful degradation).
func TestDetectOutOfScopeChanges_EmptyInputs(t *testing.T) {
	outOfScope, err := detectOutOfScopeChanges("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if outOfScope != nil {
		t.Errorf("expected nil, got %v", outOfScope)
	}
}

// TestDetectOutOfScopeChanges_WorkingDirIsProjectRoot verifies that when
// workingDir == projectRoot every file is considered in-scope.
func TestDetectOutOfScopeChanges_WorkingDirIsProjectRoot(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)

	// Commit + modify a file at the root.
	rootFile := filepath.Join(dir, "entity.go")
	if err := os.WriteFile(rootFile, []byte("package main\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "-C", dir, "add", "entity.go")
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@test.com",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@test.com",
	)
	cmd.Run() //nolint:errcheck
	cmd2 := exec.Command("git", "-C", dir, "commit", "-m", "add entity.go")
	cmd2.Env = cmd.Env
	cmd2.Run() //nolint:errcheck

	if err := os.WriteFile(rootFile, []byte("// modified\n"), 0644); err != nil {
		t.Fatal(err)
	}

	outOfScope, err := detectOutOfScopeChanges(dir, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(outOfScope) != 0 {
		t.Errorf("expected no out-of-scope files when workDir==projectRoot, got %v", outOfScope)
	}
}
