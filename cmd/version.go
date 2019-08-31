package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "outputs version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("0.0.1-alpha\n")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
