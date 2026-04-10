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
	var services []string
	if len(args) > 0 {
		services = args
		for _, svc := range services {
			if err := cfg.ValidateService(svc); err != nil {
				return err
			}
		}
	} else {
		var err error
		services, err = resolveServices(cfg, restartAll)
		if err != nil {
			return err
		}
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

// resolveServices determines which services to operate on.
// If all is true, returns all services. Otherwise tries auto-detection,
// falling back to all services when they share a single dir.
func resolveServices(cfg *config.Config, all bool) ([]string, error) {
	if all {
		return cfg.ServiceNames(), nil
	}
	svc, err := detectService(cfg)
	if err != nil && cfg.HasSingleDir() {
		return cfg.ServiceNames(), nil
	} else if err != nil {
		return nil, err
	}
	return []string{svc}, nil
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
