package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("homey-cli version %s\n", versionInfo.Version)
		if versionInfo.Commit != "unknown" {
			fmt.Printf("commit: %s\n", versionInfo.Commit)
		}
		if versionInfo.Date != "unknown" {
			fmt.Printf("built: %s\n", versionInfo.Date)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
