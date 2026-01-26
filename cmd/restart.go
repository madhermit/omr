package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/madhermit/omr/internal/config"
	"github.com/madhermit/omr/internal/git"
	"github.com/madhermit/omr/internal/overmind"
	"github.com/madhermit/omr/internal/symlink"
	"github.com/spf13/cobra"
)

var restartAll bool

var restartCmd = &cobra.Command{
	Use:   "restart [services...]",
	Short: "Restart services",
	Long: `Restart one or more services by updating symlinks and restarting overmind processes.

If no services are specified, omr will auto-detect the service based on
the current directory (looking for detect files like nuxt.config.ts or
config/application.rb).

Use --all to restart all configured services.`,
	RunE: runRestart,
}

func init() {
	restartCmd.Flags().BoolVarP(&restartAll, "all", "a", false, "restart all services")
	rootCmd.AddCommand(restartCmd)
}

func runRestart(cmd *cobra.Command, args []string) error {
	if err := cfg.ValidateRoot(); err != nil {
		return err
	}

	// Determine which services to restart
	var services []string
	if restartAll {
		services = cfg.ServiceNames()
	} else if len(args) > 0 {
		services = args
		for _, svc := range services {
			if err := cfg.ValidateService(svc); err != nil {
				return err
			}
		}
	} else {
		svc, err := detectService(cfg)
		if err != nil {
			return err
		}
		services = []string{svc}
	}

	worktreePath, err := git.RepoRoot()
	if err != nil {
		return fmt.Errorf("determining current worktree: %w", err)
	}

	branch, err := git.GetCurrentBranch(worktreePath)
	if err != nil {
		return fmt.Errorf("determining current branch: %w", err)
	}

	return doRestart(cfg, services, worktreePath, branch)
}

func doRestart(cfg *config.Config, services []string, worktreePath, branch string) error {
	var allProcs []string

	for _, svcName := range services {
		svc := cfg.Services[svcName]

		log("Restarting %s:\n", color.MagentaString(svcName))
		log("  Branch: %s\n", color.GreenString(branch))
		log("  Path:   %s\n", color.BlueString(worktreePath))

		if err := symlink.Create(cfg.Root, svc.Dir, worktreePath); err != nil {
			return fmt.Errorf("creating symlink for %s: %w", svcName, err)
		}

		linkPath := filepath.Join(cfg.Root, svc.Dir)
		valid, target, err := symlink.Verify(linkPath)
		if err != nil {
			return fmt.Errorf("verifying symlink for %s: %w", svcName, err)
		}
		if !valid {
			return fmt.Errorf("symlink target does not exist for %s: %s", svcName, target)
		}

		allProcs = append(allProcs, svc.Procs...)
	}

	if !overmind.IsRunning() {
		warn("Overmind is not running. Start it with: overmind start")
	} else if len(allProcs) > 0 {
		log("\nRestarting overmind processes: %v\n", allProcs)
		if err := overmind.Restart(allProcs...); err != nil {
			return fmt.Errorf("restarting overmind: %w", err)
		}
	}

	if count := overmind.CountInstances(); count > 1 {
		warn("Multiple overmind instances detected (%d)", count)
	}

	return nil
}

func detectService(cfg *config.Config) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting current directory: %w", err)
	}

	for name, svc := range cfg.Services {
		if svc.Detect == "" {
			continue
		}
		if _, err := os.Stat(filepath.Join(cwd, svc.Detect)); err == nil {
			return name, nil
		}
	}

	return "", fmt.Errorf("could not auto-detect service type; specify service name or use --all")
}
