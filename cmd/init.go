package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const exampleConfig = `# OMR Configuration File
# Save this as .omr.toml in your project root or home directory

# Root directory where symlinks are managed
root = "/path/to/active"

# Service definitions
[services.api]
dir = "api"                         # Symlink name in root directory
procs = ["rails", "worker"]         # Overmind process names to restart
detect = "config/application.rb"    # File to detect this service type (optional)

[services.frontend]
dir = "app"
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
