package cmd

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/madhermit/omr/internal/git"
	"github.com/madhermit/omr/internal/plan"
	"github.com/spf13/cobra"
)

var switchAll bool

const defaultWaveTimeout = 60 * time.Second

var switchCmd = &cobra.Command{
	Use:   "switch [branch]",
	Short: "Switch worktree and restart services",
	Long: `Switch to a different git worktree and restart services.

If no branch is specified, uses the current worktree.
If --all is specified, switches all services; otherwise switches only the
auto-detected service.

When path/port/depends_on are configured, restart is change-driven (only
services whose source actually changed restart) and sequenced (waits for
dependencies to be ready before starting dependents).

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

	branch, err := resolveBranch(args)
	if err != nil {
		return err
	}

	p, err := plan.Build(cfg, services, plan.BuildResolver(cfg, branch))
	if err != nil {
		return err
	}

	shown := map[string]bool{}
	for _, flip := range p.DirFlips {
		shown[flip.Dir] = true
		log("Switching %s → %s\n", color.MagentaString(flip.Dir), color.CyanString(branch))
		log("  Path: %s\n", color.BlueString(flip.NewTarget))
	}
	for _, name := range services {
		dir := cfg.Services[name].Dir
		if shown[dir] {
			continue
		}
		shown[dir] = true
		log("  %s: already on %s\n", color.MagentaString(dir), color.CyanString(branch))
	}

	if p.IsEmpty() {
		logln(color.GreenString("Already on"), color.CyanString(branch))
		return nil
	}

	if !p.AnyRestarts() && len(p.DirFlips) > 0 {
		log("\nSymlinks flipped; no service source changed — skipping restart.\n")
		return nil
	}

	return plan.Execute(cmd.Context(), cfg, p, defaultWaveTimeout, planLogger())
}

func resolveBranch(args []string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}
	worktreePath, err := git.RepoRoot()
	if err != nil {
		return "", fmt.Errorf("determining current worktree: %w", err)
	}
	branch, err := git.GetCurrentBranch(worktreePath)
	if err != nil {
		return "", fmt.Errorf("determining current branch: %w", err)
	}
	return branch, nil
}

func planLogger() plan.Logger {
	return plan.Logger{
		Logf: func(format string, args ...interface{}) {
			log(format, args...)
		},
		Warnf: warn,
	}
}
