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
- Whether overmind is running`,
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
	if overmind.IsRunning(cfg.Root) {
		logln(color.GreenString("Overmind:"), "running")
	} else {
		logln(color.YellowString("Overmind:"), "not running")
	}

	logln()
	logln(color.MagentaString("Services:"))

	shown := map[string]bool{}
	for _, name := range cfg.ServiceNames() {
		svc := cfg.Services[name]
		linkPath := filepath.Join(cfg.Root, svc.Dir)

		// Show symlink status once per unique dir
		if !shown[svc.Dir] {
			shown[svc.Dir] = true

			valid, target, err := symlink.Verify(linkPath)
			if err != nil {
				log("  %s: %s\n", color.MagentaString(name), color.RedString(err.Error()))
				continue
			}
			if !symlink.Exists(linkPath) {
				log("  %s: %s\n", color.MagentaString(name), color.YellowString("not a symlink: %s", linkPath))
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
			log("  %s: %s (%s)\n", color.MagentaString(svc.Dir), color.GreenString(branch), color.BlueString(target))
		}

		log("    %s: %v\n", color.MagentaString(name), svc.Procs)
	}

	return nil
}
