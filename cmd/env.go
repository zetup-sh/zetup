package cmd

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

// envCmd represents the env command
var envCmd = &cobra.Command{
	Use:   "env",
	Short: "useful zetup environment variables",
	//Long: `A longer description that spans multiple lines and likely contains examples
	//and usage of using your command. For example:

	Run: func(cmd *cobra.Command, args []string) {
		envStr := ""
		if runtime.GOOS == "linux" {
			for key, setting := range mainViper.AllSettings() {
				v, ok := setting.(string)
				if ok {
					if !strings.HasPrefix(key, "zetup-") {
						envStr += "ZETUP_"
					}
					envStr += strings.ToUpper(strings.ReplaceAll(key, "-", "_")) + "=" + v + "\n"
				}
			}
			envStr += "# add eval `zetup env` to your .bashrc"
		}
		fmt.Fprintf(os.Stdout, envStr)
	},
}

func init() {
	rootCmd.AddCommand(envCmd)
}
