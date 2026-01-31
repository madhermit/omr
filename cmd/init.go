package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const exampleConfig = `# OMR Configuration File
# Save this as .omr.toml in your project root

[services.api]
dir = "api/current"                 # Symlink path (relative to this file)
procs = ["rails", "worker"]         # Overmind process names to restart
detect = "config/application.rb"    # File to detect this service type (optional)

[services.frontend]
dir = "frontend/current"
procs = ["app"]
detect = "nuxt.config.ts"
`

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate example configuration file",
	Long: `Generate an example .omr.toml configuration file.

The output is printed to stdout so you can redirect it to a file:

  omr init > .omr.toml

Then edit the file to match your project structure.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(exampleConfig)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
