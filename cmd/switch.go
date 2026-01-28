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

	// Determine branch
	var branch string
	if len(args) == 0 {
		// Use current worktree's branch
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

	// Switch each service to its worktree for the branch
	return doSwitchServices(services, branch)
}

func doSwitchServices(services []string, branch string) error {
	var allProcs []string
	anyChanged := false

	for _, svcName := range services {
		svc := cfg.Services[svcName]
		linkPath := filepath.Join(cfg.Root, svc.Dir)

		// Find worktree for this service's repo
		// The service dir is like "first-api/current", parent dir contains worktrees
		svcRepoDir := filepath.Dir(linkPath)
		worktreePath, err := git.GetWorktreePathInDir(svcRepoDir, branch)
		if err != nil {
			return fmt.Errorf("finding worktree for %s branch '%s': %w", svcName, branch, err)
		}

		// Check if already on this worktree
		_, currentTarget, _ := symlink.Verify(linkPath)
		if currentTarget == worktreePath {
			log("  %s: already on %s\n", color.MagentaString(svcName), color.CyanString(branch))
			continue
		}

		log("Switching %s:\n", color.MagentaString(svcName))
		log("  Branch: %s\n", color.GreenString(branch))
		log("  Path:   %s\n", color.BlueString(worktreePath))

		if err := symlink.Create(cfg.Root, svc.Dir, worktreePath); err != nil {
			return fmt.Errorf("creating symlink for %s: %w", svcName, err)
		}

		allProcs = append(allProcs, svc.Procs...)
		anyChanged = true
	}

	if !anyChanged {
		logln(color.GreenString("Already on"), color.CyanString(branch))
		return nil
	}

	// Restart overmind processes
	if !overmind.IsRunning(cfg.Root) {
		warn("Overmind is not running. Start it with: overmind start")
	} else if len(allProcs) > 0 {
		log("\nRestarting overmind processes: %v\n", allProcs)
		if err := overmind.Restart(cfg.Root, allProcs...); err != nil {
			return fmt.Errorf("restarting overmind: %w", err)
		}
	}

	return nil
}
