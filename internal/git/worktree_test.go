package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestParseWorktreeLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected Worktree
	}{
		{
			name: "normal worktree",
			line: "/home/user/project  abc1234 [main]",
			expected: Worktree{
				Path:   "/home/user/project",
				Branch: "main",
			},
		},
		{
			name: "feature branch",
			line: "/home/user/project-feature  def5678 [feature/new-thing]",
			expected: Worktree{
				Path:   "/home/user/project-feature",
				Branch: "feature/new-thing",
			},
		},
		{
			name: "bare repository",
			line: "/home/user/project.git  (bare)",
			expected: Worktree{
				Path: "/home/user/project.git",
				Bare: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseWorktreeLine(tt.line)
			if got.Path != tt.expected.Path {
				t.Errorf("Path: expected %q, got %q", tt.expected.Path, got.Path)
			}
			if got.Branch != tt.expected.Branch {
				t.Errorf("Branch: expected %q, got %q", tt.expected.Branch, got.Branch)
			}
			if got.Bare != tt.expected.Bare {
				t.Errorf("Bare: expected %v, got %v", tt.expected.Bare, got.Bare)
			}
		})
	}
}

func TestValidateWorktreePath(t *testing.T) {
	// Test with non-existent path
	if err := ValidateWorktreePath("/nonexistent/path"); err == nil {
		t.Error("expected error for nonexistent path")
	}

	// Test with regular file
	tmpFile, _ := os.CreateTemp("", "test")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	if err := ValidateWorktreePath(tmpFile.Name()); err == nil {
		t.Error("expected error for regular file")
	}

	// Test with directory that's not a git repo
	tmpDir, _ := os.MkdirTemp("", "test")
	defer os.RemoveAll(tmpDir)

	if err := ValidateWorktreePath(tmpDir); err == nil {
		t.Error("expected error for non-git directory")
	}
}

func TestGetCurrentBranch(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := exec.Command("git", "init", tmpDir).Run(); err != nil {
		t.Skipf("git not available: %v", err)
	}

	exec.Command("git", "-C", tmpDir, "config", "user.email", "test@test.com").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.name", "Test").Run()

	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0o644)
	exec.Command("git", "-C", tmpDir, "add", ".").Run()
	exec.Command("git", "-C", tmpDir, "commit", "-m", "initial").Run()

	branch, err := GetCurrentBranch(tmpDir)
	if err != nil {
		t.Fatalf("GetCurrentBranch failed: %v", err)
	}

	if branch != "master" && branch != "main" {
		t.Errorf("expected 'master' or 'main', got %q", branch)
	}
}

func TestRepoRoot(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := exec.Command("git", "init", tmpDir).Run(); err != nil {
		t.Skipf("git not available: %v", err)
	}

	subDir := filepath.Join(tmpDir, "sub", "dir")
	os.MkdirAll(subDir, 0o755)

	origDir, _ := os.Getwd()
	os.Chdir(subDir)
	defer os.Chdir(origDir)

	root, err := RepoRoot()
	if err != nil {
		t.Fatalf("RepoRoot failed: %v", err)
	}

	expectedRoot, _ := filepath.EvalSymlinks(tmpDir)
	gotRoot, _ := filepath.EvalSymlinks(root)

	if gotRoot != expectedRoot {
		t.Errorf("expected %q, got %q", expectedRoot, gotRoot)
	}
}
