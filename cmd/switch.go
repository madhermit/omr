package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/madhermit/omr/internal/git"
	"github.com/madhermit/omr/internal/symlink"
	"github.com/spf13/cobra"
)

var switchAll bool

var switchCmd = &cobra.Command{
	Use:   "switch [branch]",
	Short: "Switch worktree and restart services",
	Long: `Switch to a different git worktree and restart services.

If no branch is specified, uses the current worktree.
If --all is specified, switches all services; otherwise switches only the
auto-detected service.

Examples:
  omr switch             # Switch detected service to current worktree
  omr switch main        # Switch detected service to main worktree
  omr switch --all       # Switch all services to current worktree
  omr switch --all main  # Switch all services to main worktree`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSwitch,
}

func init() {
	switchCmd.Flags().BoolVarP(&switchAll, "all", "a", false, "switch all services")
	rootCmd.AddCommand(switchCmd)
}

func runSwitch(cmd *cobra.Command, args []string) error {
	if err := cfg.ValidateRoot(); err != nil {
		return err
	}

	// Determine worktree path and branch
	var worktreePath, branch string
	var err error

	if len(args) == 0 {
		worktreePath, err = git.RepoRoot()
		if err != nil {
			return fmt.Errorf("determining current worktree: %w", err)
		}
		branch, err = git.GetCurrentBranch(worktreePath)
		if err != nil {
			return fmt.Errorf("determining current branch: %w", err)
		}
	} else {
		branch = args[0]
		worktreePath, err = git.GetWorktreePath(branch)
		if err != nil {
			return fmt.Errorf("finding worktree for branch '%s': %w", branch, err)
		}
		if err := git.ValidateWorktreePath(worktreePath); err != nil {
			return fmt.Errorf("invalid worktree: %w", err)
		}
	}

	// Determine which services to switch
	var services []string
	if switchAll {
		services = cfg.ServiceNames()
	} else {
		svcName, err := detectService(cfg)
		if err != nil {
			return err
		}
		services = []string{svcName}
	}

	if !needsSwitch(services, worktreePath) {
		logln(color.GreenString("Already on"), color.CyanString(branch))
		return nil
	}

	return doRestart(cfg, services, worktreePath, branch)
}

func needsSwitch(services []string, worktreePath string) bool {
	for _, name := range services {
		svc := cfg.Services[name]
		linkPath := filepath.Join(cfg.Root, svc.Dir)
		_, target, _ := symlink.Verify(linkPath)
		if target != worktreePath {
			return true
		}
	}
	return false
}
