package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version is set at build time
	Version = "dev"
	// Commit is set at build time
	Commit = "none"
	// Date is set at build time
	Date = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("omr version %s\n", Version)
		if Commit != "none" {
			fmt.Printf("  commit: %s\n", Commit)
		}
		if Date != "unknown" {
			fmt.Printf("  built:  %s\n", Date)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
