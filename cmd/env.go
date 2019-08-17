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

		// TODO:.windows, and I think some
		// linux distros don't know "export"
		if prefix == "" {
			if isUnix() {
				if commandExists("export") {
					prefix = "export "
				} else if commandExists("set") {
					prefix = "set "
				}
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

			// I hate go
			// if key == "identities" {
			// 	idErrMessage := "Sorry, there was a problem with an identity"
			// 	idArray, ok := setting.(map[string]interface{})
			// 	if ok {
			// 		for idType, ids := range idArray {
			// 			idGroupArray, ok := ids.([]interface{})
			// 			if ok {
			// 				for _, idGroup := range idGroupArray {
			// 					id, ok := idGroup.(map[interface{}]interface{})
			// 					if ok {
			// 						log.Println("type:", idType)
			// 						for key, val := range id {
			// 							log.Printf("%v: %v", key, val)
			// 						}
			// 					} else {
			// 						log.Fatal(idErrMessage)
			// 					}
			// 				}
			// 			} else {
			// 				log.Fatal(idErrMessage)
			// 			}
			// 		}
			// 	} else {
			// 		log.Fatal(idErrMessage)
			// 	}
			// }
			// n, ok := setting.([]interface{})
			// if ok {
			// 	fmt.Println("[]interface")
			// 	log.Printf("n", n)
			// }
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
