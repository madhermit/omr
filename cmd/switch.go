package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/madhermit/omr/internal/git"
	"github.com/madhermit/omr/internal/overmind"
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

	services, err := resolveServices(cfg, switchAll)
	if err != nil {
		return err
	}

	// Determine branch
	var branch string
	if len(args) == 0 {
		worktreePath, err := git.RepoRoot()
		if err != nil {
			return fmt.Errorf("determining current worktree: %w", err)
		}
		branch, err = git.GetCurrentBranch(worktreePath)
		if err != nil {
			return fmt.Errorf("determining current branch: %w", err)
		}
	} else {
		branch = args[0]
	}

	// Deduplicate dirs — multiple services may share one symlink
	switched := map[string]bool{}
	var allProcs []string
	anyChanged := false

	for _, svcName := range services {
		svc := cfg.Services[svcName]
		if switched[svc.Dir] {
			continue
		}
		switched[svc.Dir] = true

		linkPath := filepath.Join(cfg.Root, svc.Dir)
		valid, currentTarget, _ := symlink.Verify(linkPath)

		// Use resolved symlink target for worktree discovery, fall back to parent dir
		searchDir := filepath.Dir(linkPath)
		if valid && currentTarget != "" {
			searchDir = currentTarget
		}

		worktreePath, err := git.GetWorktreePathInDir(searchDir, branch)
		if err != nil {
			return fmt.Errorf("finding worktree for %s branch '%s': %w", svcName, branch, err)
		}

		if currentTarget == worktreePath {
			log("  %s: already on %s\n", color.MagentaString(svc.Dir), color.CyanString(branch))
			continue
		}

		log("Switching %s → %s\n", color.MagentaString(svc.Dir), color.CyanString(branch))
		log("  Path: %s\n", color.BlueString(worktreePath))

		if err := symlink.Create(cfg.Root, svc.Dir, worktreePath); err != nil {
			return fmt.Errorf("creating symlink for %s: %w", svc.Dir, err)
		}

		// Restart all services sharing this dir, not just the detected one
		allProcs = append(allProcs, cfg.ProcsForDir(svc.Dir)...)
		anyChanged = true
	}

	if !anyChanged {
		logln(color.GreenString("Already on"), color.CyanString(branch))
		return nil
	}

	if !overmind.IsRunning(cfg.Root) {
		warn("Overmind is not running. Start it with: overmind start")
		return nil
	}
	if len(allProcs) > 0 {
		log("\nRestarting overmind processes: %v\n", allProcs)
		if err := overmind.Restart(cfg.Root, allProcs...); err != nil {
			return fmt.Errorf("restarting overmind: %w", err)
		}
	}
	return nil
}
