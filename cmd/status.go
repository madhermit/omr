package cmd

import (
	"path/filepath"

	"github.com/fatih/color"
	"github.com/madhermit/omr/internal/git"
	"github.com/madhermit/omr/internal/overmind"
	"github.com/madhermit/omr/internal/symlink"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current service status",
	Long: `Show the status of all configured services including:
- Current branch (from symlink target)
- Symlink path and target
- Whether overmind is running
- Available git worktrees`,
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	if err := cfg.ValidateRoot(); err != nil {
		return err
	}

	// Check overmind status
	if overmind.IsRunning() {
		logln(color.GreenString("Overmind:"), "running")
	} else {
		logln(color.YellowString("Overmind:"), "not running")
	}

	if count := overmind.CountInstances(); count > 1 {
		warn("Multiple overmind instances detected (%d)", count)
	}

	logln()
	logln(color.MagentaString("Services:"))

	for _, name := range cfg.ServiceNames() {
		svc := cfg.Services[name]
		linkPath := filepath.Join(cfg.Root, svc.Dir)

		valid, target, err := symlink.Verify(linkPath)
		if err != nil {
			log("  %s: %s\n", color.MagentaString(name), color.RedString(err.Error()))
			continue
		}

		if !symlink.Exists(linkPath) {
			log("  %s: %s\n", color.MagentaString(name), color.YellowString("not linked"))
			continue
		}

		if !valid {
			log("  %s: %s (target: %s)\n", color.MagentaString(name), color.RedString("broken symlink"), target)
			continue
		}

		branch, err := git.GetCurrentBranch(target)
		if err != nil {
			branch = "unknown"
		}

		log("  %s:\n", color.MagentaString(name))
		log("    Branch: %s\n", color.GreenString(branch))
		log("    Path:   %s\n", color.BlueString(target))
		log("    Procs:  %v\n", svc.Procs)
	}

	logln()
	logln(color.MagentaString("Available worktrees:"))

	worktrees, err := git.ListWorktrees()
	if err != nil {
		warn("Could not list worktrees: %v", err)
	} else {
		for _, wt := range worktrees {
			if wt.Bare {
				log("  %s %s\n", color.BlueString(wt.Path), color.YellowString("(bare)"))
			} else {
				log("  %s [%s]\n", color.BlueString(wt.Path), color.GreenString(wt.Branch))
			}
		}
	}

	return nil
}
