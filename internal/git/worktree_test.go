package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// initTestRepo creates a temporary git repo with an initial commit and returns its path.
func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	if err := exec.Command("git", "init", dir).Run(); err != nil {
		t.Skipf("git not available: %v", err)
	}
	git("config", "user.email", "test@test.com")
	git("config", "user.name", "Test")
	os.WriteFile(filepath.Join(dir, "test.txt"), []byte("test"), 0o644)
	git("add", ".")
	git("commit", "-m", "initial")

	return dir
}

func TestParseWorktreeLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected Worktree
	}{
		{
			name:     "normal worktree",
			line:     "/home/user/project  abc1234 [main]",
			expected: Worktree{Path: "/home/user/project", Branch: "main"},
		},
		{
			name:     "feature branch",
			line:     "/home/user/project-feature  def5678 [feature/new-thing]",
			expected: Worktree{Path: "/home/user/project-feature", Branch: "feature/new-thing"},
		},
		{
			name:     "bare repository",
			line:     "/home/user/project.git  (bare)",
			expected: Worktree{Path: "/home/user/project.git", Bare: true},
		},
		{
			name:     "empty line",
			line:     "",
			expected: Worktree{},
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

func TestParseWorktreeList(t *testing.T) {
	input := []byte("/repo/.bare  (bare)\n/repo/main  abc1234 [main]\n/repo/feature  def5678 [feature/x]\n\n")

	worktrees, err := parseWorktreeList(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(worktrees) != 3 {
		t.Fatalf("expected 3 worktrees, got %d", len(worktrees))
	}
	if w := worktrees[0]; !w.Bare || w.Path != "/repo/.bare" {
		t.Errorf("worktree[0]: expected bare at /repo/.bare, got %+v", w)
	}
	if w := worktrees[1]; w.Branch != "main" || w.Path != "/repo/main" {
		t.Errorf("worktree[1]: expected main at /repo/main, got %+v", w)
	}
	if w := worktrees[2]; w.Branch != "feature/x" || w.Path != "/repo/feature" {
		t.Errorf("worktree[2]: expected feature/x at /repo/feature, got %+v", w)
	}
}

func TestValidateWorktreePath(t *testing.T) {
	if err := ValidateWorktreePath("/nonexistent/path"); err == nil {
		t.Error("expected error for nonexistent path")
	}

	file := filepath.Join(t.TempDir(), "file")
	os.WriteFile(file, []byte("test"), 0o644)
	if err := ValidateWorktreePath(file); err == nil {
		t.Error("expected error for regular file")
	}

	if err := ValidateWorktreePath(t.TempDir()); err == nil {
		t.Error("expected error for non-git directory")
	}
}

func TestGetCurrentBranch(t *testing.T) {
	dir := initTestRepo(t)

	branch, err := GetCurrentBranch(dir)
	if err != nil {
		t.Fatalf("GetCurrentBranch failed: %v", err)
	}
	if branch != "master" && branch != "main" {
		t.Errorf("expected 'master' or 'main', got %q", branch)
	}
}

func TestListWorktreesInDir(t *testing.T) {
	dir := initTestRepo(t)

	worktrees, err := ListWorktreesInDir(dir)
	if err != nil {
		t.Fatalf("ListWorktreesInDir failed: %v", err)
	}
	if len(worktrees) == 0 {
		t.Fatal("expected at least one worktree")
	}

	found := false
	for _, wt := range worktrees {
		if wt.Path == dir {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected to find worktree at %s", dir)
	}
}

func TestListWorktreesInDir_WithWorktree(t *testing.T) {
	dir := initTestRepo(t)
	mainBranch, _ := GetCurrentBranch(dir)

	featureDir := filepath.Join(t.TempDir(), "feature")
	cmd := exec.Command("git", "-C", dir, "worktree", "add", featureDir, "-b", "feature-branch")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git worktree add: %v\n%s", err, out)
	}

	// Both directories should return the same worktree list
	for _, searchDir := range []string{dir, featureDir} {
		worktrees, err := ListWorktreesInDir(searchDir)
		if err != nil {
			t.Fatalf("ListWorktreesInDir(%s): %v", searchDir, err)
		}
		if len(worktrees) != 2 {
			t.Fatalf("from %s: expected 2 worktrees, got %d", searchDir, len(worktrees))
		}
		branches := map[string]bool{}
		for _, wt := range worktrees {
			branches[wt.Branch] = true
		}
		if !branches[mainBranch] || !branches["feature-branch"] {
			t.Errorf("from %s: expected branches %q and %q, got %v", searchDir, mainBranch, "feature-branch", branches)
		}
	}
}

func TestListWorktreesInDir_NotGitRepo(t *testing.T) {
	if _, err := ListWorktreesInDir(t.TempDir()); err == nil {
		t.Error("expected error for non-git directory")
	}
}

func TestGetWorktreePathInDir(t *testing.T) {
	dir := initTestRepo(t)
	branch, _ := GetCurrentBranch(dir)

	path, err := GetWorktreePathInDir(dir, branch)
	if err != nil {
		t.Fatalf("GetWorktreePathInDir failed: %v", err)
	}

	expectedPath, _ := filepath.EvalSymlinks(dir)
	gotPath, _ := filepath.EvalSymlinks(path)
	if gotPath != expectedPath {
		t.Errorf("expected %q, got %q", expectedPath, gotPath)
	}

	if _, err := GetWorktreePathInDir(dir, "nonexistent-branch"); err == nil {
		t.Error("expected error for nonexistent branch")
	}
}

func TestRepoRoot(t *testing.T) {
	dir := initTestRepo(t)
	subDir := filepath.Join(dir, "sub", "dir")
	os.MkdirAll(subDir, 0o755)

	origDir, _ := os.Getwd()
	os.Chdir(subDir)
	defer os.Chdir(origDir)

	root, err := RepoRoot()
	if err != nil {
		t.Fatalf("RepoRoot failed: %v", err)
	}

	expectedRoot, _ := filepath.EvalSymlinks(dir)
	gotRoot, _ := filepath.EvalSymlinks(root)
	if gotRoot != expectedRoot {
		t.Errorf("expected %q, got %q", expectedRoot, gotRoot)
	}
}
