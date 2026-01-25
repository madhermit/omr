package state

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWriteAndRead(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "state-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write state
	st := &State{
		Worktree: "/path/to/worktree",
		Branch:   "main",
		Service:  "api",
	}

	if err := Write(tmpDir, st); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Verify file exists
	statePath := filepath.Join(tmpDir, stateFileName)
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Fatal("state file was not created")
	}

	// Read state back
	readState, err := Read(tmpDir)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if readState.Worktree != st.Worktree {
		t.Errorf("expected worktree %q, got %q", st.Worktree, readState.Worktree)
	}
	if readState.Branch != st.Branch {
		t.Errorf("expected branch %q, got %q", st.Branch, readState.Branch)
	}
	if readState.Service != st.Service {
		t.Errorf("expected service %q, got %q", st.Service, readState.Service)
	}
	if readState.Timestamp.IsZero() {
		t.Error("expected timestamp to be set")
	}
}

func TestReadNonexistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "state-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Read from directory with no state file
	st, err := Read(tmpDir)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if st != nil {
		t.Error("expected nil state for nonexistent file")
	}
}

func TestWriteSetsTimestamp(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "state-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	before := time.Now()

	st := &State{
		Worktree: "/path",
		Branch:   "main",
		Service:  "api",
	}
	if err := Write(tmpDir, st); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	after := time.Now()

	// Verify timestamp was set
	if st.Timestamp.Before(before) || st.Timestamp.After(after) {
		t.Errorf("timestamp %v not in expected range [%v, %v]", st.Timestamp, before, after)
	}
}

func TestGetStatePath(t *testing.T) {
	root := "/some/root"
	expected := filepath.Join(root, stateFileName)
	got := GetStatePath(root)
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestWriteFileContent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "state-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	st := &State{
		Worktree: "/my/worktree",
		Branch:   "feature-branch",
		Service:  "frontend",
	}
	if err := Write(tmpDir, st); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Read raw content
	content, err := os.ReadFile(filepath.Join(tmpDir, stateFileName))
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, `worktree = "/my/worktree"`) {
		t.Error("content missing worktree")
	}
	if !strings.Contains(contentStr, `branch = "feature-branch"`) {
		t.Error("content missing branch")
	}
	if !strings.Contains(contentStr, `service = "frontend"`) {
		t.Error("content missing service")
	}
}
