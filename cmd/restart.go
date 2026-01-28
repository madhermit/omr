package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/madhermit/omr/internal/config"
	"github.com/madhermit/omr/internal/overmind"
	"github.com/spf13/cobra"
)

var restartAll bool

var restartCmd = &cobra.Command{
	Use:   "restart [services...]",
	Short: "Restart overmind processes",
	Long: `Restart overmind processes for one or more services.

This does NOT change symlinks - use 'omr switch' for that.

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

	// Collect all procs to restart
	var allProcs []string
	for _, svcName := range services {
		svc := cfg.Services[svcName]
		allProcs = append(allProcs, svc.Procs...)
	}

	if !overmind.IsRunning(cfg.Root) {
		return fmt.Errorf("overmind is not running; start it with: overmind start")
	}

	if len(allProcs) == 0 {
		return fmt.Errorf("no processes configured for services: %v", services)
	}

	log("Restarting processes: %v\n", allProcs)
	if err := overmind.Restart(cfg.Root, allProcs...); err != nil {
		return fmt.Errorf("restarting overmind: %w", err)
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
