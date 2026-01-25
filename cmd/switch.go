package cmd

import (
	"fmt"

	"github.com/madhermit/omr/internal/git"
	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:   "switch <worktree>",
	Short: "Switch worktree and restart services",
	Long: `Switch to a different git worktree and restart all services.

The worktree argument should be a branch name. OMR will look up the
corresponding worktree path and update all symlinks to point to it.

Example:
  omr switch main        # Switch to main branch worktree
  omr switch feature-x   # Switch to feature-x branch worktree`,
	Args: cobra.ExactArgs(1),
	RunE: runSwitch,
}

func init() {
	rootCmd.AddCommand(switchCmd)
}

func runSwitch(cmd *cobra.Command, args []string) error {
	if err := cfg.ValidateRoot(); err != nil {
		return err
	}

	branch := args[0]

	worktreePath, err := git.GetWorktreePath(branch)
	if err != nil {
		return fmt.Errorf("finding worktree for branch '%s': %w", branch, err)
	}

	if err := git.ValidateWorktreePath(worktreePath); err != nil {
		return fmt.Errorf("invalid worktree: %w", err)
	}

	return doRestart(cfg, cfg.ServiceNames(), worktreePath, branch)
}
