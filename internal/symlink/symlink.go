package symlink

import (
	"fmt"
	"os"
	"path/filepath"
)

// Create creates a symlink at root/name pointing to target.
// Replaces any existing symlink or file.
func Create(root, name, target string) error {
	linkPath := filepath.Join(root, name)

	// Remove existing symlink or file
	if _, err := os.Lstat(linkPath); err == nil {
		if err := os.Remove(linkPath); err != nil {
			return fmt.Errorf("removing existing path: %w", err)
		}
	}

	if err := os.Symlink(target, linkPath); err != nil {
		return fmt.Errorf("creating symlink: %w", err)
	}
	return nil
}

// Verify checks if a symlink exists and its target is valid.
// Returns: valid (target exists), resolved target path, error
func Verify(path string) (valid bool, target string, err error) {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, "", nil
		}
		return false, "", err
	}

	if info.Mode()&os.ModeSymlink == 0 {
		return false, "", fmt.Errorf("not a symlink: %s", path)
	}

	target, err = os.Readlink(path)
	if err != nil {
		return false, "", err
	}

	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(path), target)
	}

	if _, err := os.Stat(target); os.IsNotExist(err) {
		return false, target, nil // Broken symlink
	} else if err != nil {
		return false, target, err
	}

	return true, target, nil
}

// Exists checks if a symlink exists at the given path
func Exists(path string) bool {
	info, err := os.Lstat(path)
	return err == nil && info.Mode()&os.ModeSymlink != 0
}
