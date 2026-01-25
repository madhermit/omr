package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/madhermit/omr/internal/config"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	quiet   bool
	cfg     *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "omr",
	Short: "Overmind Restart - manage worktrees and restart services",
	Long: `OMR (Overmind Restart) manages git worktree symlinks and restarts
overmind processes for seamless branch switching in development.

Configure services in .omr.toml and use omr to switch between
worktrees while automatically restarting the appropriate services.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Name() == "init" || cmd.Name() == "version" {
			return nil
		}

		if cfgFile != "" {
			config.SetConfigFile(cfgFile)
		}

		var err error
		cfg, err = config.Load()
		if err != nil {
			return err
		}
		return nil
	},
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s %s\n", color.RedString("Error:"), err)
		return err
	}
	return nil
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default: .omr.toml)")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suppress output")
}

func log(format string, args ...interface{}) {
	if !quiet {
		fmt.Printf(format, args...)
	}
}

func logln(args ...interface{}) {
	if !quiet {
		fmt.Println(args...)
	}
}

func warn(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s %s\n", color.YellowString("Warning:"), fmt.Sprintf(format, args...))
}
