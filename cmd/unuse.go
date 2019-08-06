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
package cmd

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// unuseCmd represents the unuse command
var unuseCmd = &cobra.Command{
	Use:   "unuse",
	Short: "undo all undoable changes made by a program",
	Long:  `This will run the unuse command in the zetup package as well as restore backups to any linked files like bashrc or tmux.conf`,
	Run: func(cmd *cobra.Command, args []string) {
		Unuse()
	},
}

func init() {
	rootCmd.AddCommand(unuseCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// unuseCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// unuseCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func Unuse() {
	RestoreBackupFiles()
}

func RestoreBackupFiles() {
	// restore linked files
	dat, err := ioutil.ReadFile(path.Join(zetupDir, ".bak.yaml"))
	check(err)
	var backedupFiles []BackupFileInfo
	yaml.Unmarshal(dat, &backedupFiles)
	for _, backedupFile := range backedupFiles {
		err = os.Remove(backedupFile.Location)
		check(err)
		ioutil.WriteFile(backedupFile.Location, []byte(backedupFile.Contents), 0644)
	}
}
