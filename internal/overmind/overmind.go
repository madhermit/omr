package overmind

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// IsRunning checks if overmind is currently running in the given directory
func IsRunning(dir string) bool {
	cmd := exec.Command("overmind", "status")
	cmd.Dir = dir
	return cmd.Run() == nil
}

// Restart restarts the specified overmind processes
func Restart(procs ...string) error {
	if len(procs) == 0 {
		return fmt.Errorf("no processes specified")
	}

	args := append([]string{"restart"}, procs...)
	if err := exec.Command("overmind", args...).Run(); err != nil {
		return fmt.Errorf("restarting processes: %w", err)
	}
	return nil
}

// CountInstances counts the number of running overmind instances
func CountInstances() int {
	out, err := exec.Command("pgrep", "-c", "overmind").Output()
	if err != nil {
		return 0
	}
	count, _ := strconv.Atoi(strings.TrimSpace(string(out)))
	return count
}
