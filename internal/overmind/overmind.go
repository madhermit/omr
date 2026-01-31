package overmind

import (
	"fmt"
	"os/exec"
	"strings"
)

// IsRunning checks if overmind is currently running in the given directory
func IsRunning(dir string) bool {
	cmd := exec.Command("overmind", "status")
	cmd.Dir = dir
	return cmd.Run() == nil
}

// Restart restarts the specified overmind processes in the given directory
func Restart(dir string, procs ...string) error {
	if len(procs) == 0 {
		return fmt.Errorf("no processes specified")
	}

	args := append([]string{"restart"}, procs...)
	cmd := exec.Command("overmind", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		if len(output) > 0 {
			return fmt.Errorf("restarting processes: %s", strings.TrimSpace(string(output)))
		}
		return fmt.Errorf("restarting processes: %w", err)
	}
	return nil
}

