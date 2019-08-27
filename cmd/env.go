package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type tIDType []map[string]string

// envCmd represents the env command
var envCmd = &cobra.Command{
	Use:   "env",
	Short: "useful zetup environment variables",
	Long: `Outputs all variables in your zetup config as environment variables.
	prefix default on windows is "$env:", and "export " or "set " on unix systems (depending on which command exists)
	`,
	Run: func(cmd *cobra.Command, args []string) {
		if prefix == "" {
			if isUnix() {
				prefix = "export "
			} else {
				prefix = "$env:"
			}
		}

		envStr := ""
		for key, setting := range mainViper.AllSettings() {
			v, ok := setting.(string)
			if ok {
				envStr += prefix
				if !strings.HasPrefix(key, "zetup-") {
					envStr += "ZETUP_"
				}
				envStr += strings.ToUpper(strings.ReplaceAll(key, "-", "_")) + "=" + v + "\n"
			}
		}
		envStr += "# add eval `zetup env` to your .bashrc\n"
		fmt.Fprintf(os.Stdout, envStr)
	},
}

var prefix string

func init() {
	rootCmd.AddCommand(envCmd)
	rootCmd.Flags().StringVarP(&prefix, "prefix", "p", "", "prefix for variables (for example `export`, `set`, etc.")
}
