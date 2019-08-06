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

	//Cobra is a CLI library for Go that empowers applications.
	//This application is a tool to generate the needed files
	//to quickly create a Cobra application.`,
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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// envCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// envCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
