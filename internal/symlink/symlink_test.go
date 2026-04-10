package symlink

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreate(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "symlink-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	targetDir := filepath.Join(tmpDir, "target")
	if err := os.Mkdir(targetDir, 0o755); err != nil {
		t.Fatalf("failed to create target dir: %v", err)
	}

	if err := Create(tmpDir, "mylink", targetDir); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	linkPath := filepath.Join(tmpDir, "mylink")
	info, err := os.Lstat(linkPath)
	if err != nil {
		t.Fatalf("failed to stat symlink: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("expected symlink, got regular file")
	}

	target, err := os.Readlink(linkPath)
	if err != nil {
		t.Fatalf("failed to read symlink: %v", err)
	}
	if target != "target" {
		t.Errorf("expected relative target %q, got %q", "target", target)
	}
}

func TestCreateReplacesExisting(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "symlink-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	target1 := filepath.Join(tmpDir, "target1")
	target2 := filepath.Join(tmpDir, "target2")
	os.Mkdir(target1, 0o755)
	os.Mkdir(target2, 0o755)

	if err := Create(tmpDir, "mylink", target1); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := Create(tmpDir, "mylink", target2); err != nil {
		t.Fatalf("Create (replace) failed: %v", err)
	}

	linkPath := filepath.Join(tmpDir, "mylink")
	target, _ := os.Readlink(linkPath)
	if target != "target2" {
		t.Errorf("expected relative target %q, got %q", "target2", target)
	}
}

func TestVerify(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "symlink-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	targetDir := filepath.Join(tmpDir, "target")
	os.Mkdir(targetDir, 0o755)
	linkPath := filepath.Join(tmpDir, "link")
	os.Symlink(targetDir, linkPath)

	valid, target, err := Verify(linkPath)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
	if !valid {
		t.Error("expected valid symlink")
	}
	if target != targetDir {
		t.Errorf("expected target %q, got %q", targetDir, target)
	}
}

func TestVerifyBrokenSymlink(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "symlink-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	linkPath := filepath.Join(tmpDir, "broken")
	os.Symlink("/nonexistent/path", linkPath)

	valid, _, err := Verify(linkPath)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
	if valid {
		t.Error("expected invalid (broken) symlink")
	}
}

func TestVerifyNonexistent(t *testing.T) {
	valid, target, err := Verify("/nonexistent/path")
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
	if valid {
		t.Error("expected invalid for nonexistent path")
	}
	if target != "" {
		t.Errorf("expected empty target, got %q", target)
	}
}

func TestExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "symlink-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	targetDir := filepath.Join(tmpDir, "target")
	os.Mkdir(targetDir, 0o755)
	linkPath := filepath.Join(tmpDir, "link")
	os.Symlink(targetDir, linkPath)

	regularFile := filepath.Join(tmpDir, "regular")
	os.WriteFile(regularFile, []byte("test"), 0o644)

	if !Exists(linkPath) {
		t.Error("expected Exists to return true for symlink")
	}
	if Exists(regularFile) {
		t.Error("expected Exists to return false for regular file")
	}
	if Exists("/nonexistent") {
		t.Error("expected Exists to return false for nonexistent path")
	}
}
