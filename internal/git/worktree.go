package git

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Worktree represents a git worktree
type Worktree struct {
	Path   string
	Branch string
	Bare   bool
}

// GetWorktreePath returns the path for a given branch's worktree
func GetWorktreePath(branch string) (string, error) {
	worktrees, err := ListWorktrees()
	if err != nil {
		return "", err
	}

	for _, wt := range worktrees {
		if wt.Branch == branch {
			return wt.Path, nil
		}
	}

	return "", fmt.Errorf("no worktree found for branch: %s", branch)
}

// GetCurrentBranch returns the current branch for a given path
func GetCurrentBranch(path string) (string, error) {
	out, err := exec.Command("git", "-C", path, "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return "", fmt.Errorf("getting current branch: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// ListWorktrees returns all git worktrees
func ListWorktrees() ([]Worktree, error) {
	out, err := exec.Command("git", "worktree", "list").Output()
	if err != nil {
		return nil, fmt.Errorf("listing worktrees: %w", err)
	}

	var worktrees []Worktree
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		if wt := parseWorktreeLine(scanner.Text()); wt.Path != "" {
			worktrees = append(worktrees, wt)
		}
	}
	return worktrees, scanner.Err()
}

// parseWorktreeLine parses a line from `git worktree list` output
// Format: /path/to/worktree  abc1234 [branch]
// or:     /path/to/bare      (bare)
func parseWorktreeLine(line string) Worktree {
	var wt Worktree

	if start, end := strings.Index(line, "["), strings.Index(line, "]"); start != -1 && end > start {
		wt.Branch = line[start+1 : end]
		line = line[:start]
	} else if strings.Contains(line, "(bare)") {
		wt.Bare = true
		line = strings.Replace(line, "(bare)", "", 1)
	}

	if fields := strings.Fields(line); len(fields) > 0 {
		wt.Path = fields[0]
	}
	return wt
}

// RepoRoot returns the root of the current git repository
func RepoRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("not in a git repository: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// ValidateWorktreePath checks if a path is a valid git worktree
func ValidateWorktreePath(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("path does not exist: %s", path)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}
	if _, err := os.Stat(filepath.Join(path, ".git")); os.IsNotExist(err) {
		return fmt.Errorf("path is not a git repository: %s", path)
	}
	return nil
}
