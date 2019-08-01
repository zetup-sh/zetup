/*
Copyright © 2019 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmdLinux

import (
	"log"
	"os/exec"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var useCmd = &cobra.Command{
	Use:   "use",
	Short: "Specify a zetup config to use",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// get sudo privileges
		getSudoCmd := exec.Command("sudo", "echo", "have sudo privileges")
		err := getSudoCmd.Start()
		if err != nil {
			log.Fatal(err)
		}
		err = getSudoCmd.Wait()
	},
}

func init() {
	rootCmd.AddCommand(useCmd)
}
