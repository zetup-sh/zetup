package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/zetup-sh/zetup/cmd/util"
	"gopkg.in/yaml.v2"
)

// unuseCmd represents the unuse command
var unuseCmd = &cobra.Command{
	Use:   "unuse",
	Short: "undo all undoable changes made by a program",
	Long:  `This will run the unuse command in the zetup package as well as restore backups to any linked files like bashrc or tmux.conf`,
	Run: func(cmd *cobra.Command, args []string) {
		unuse()
	},
}

func init() {
	rootCmd.AddCommand(unuseCmd)
}

func unuse() {
	restoreLinkFiles()
	unuseFile, err := FindFile(usePkgDir, "unuse", runtime.GOOS, unixExtensions, mainViper)
	if err == nil {
		err = runFile(unuseFile)
	}
	if err != nil {
		fmt.Println("There was a problem running the current `unuse` file")
		fmt.Printf("%s %s\n", unuseFile, err)
		fmt.Println("Probably nothing to worry about. Continuing...")
	}
	mainViper.Set("cur-pkg", "")
	mainViper.WriteConfig()
}

func restoreLinkFiles() {
	if util.Exists(linkBackupFile) {
		dat, err := ioutil.ReadFile(linkBackupFile)
		check(err)

		var backedupFiles []BackupFileInfo
		yaml.Unmarshal(dat, &backedupFiles)
		for _, backedupFile := range backedupFiles {
			err = os.Remove(backedupFile.Location)
			check(err)
			ioutil.WriteFile(backedupFile.Location, []byte(backedupFile.Contents), 0644)
		}
	}
	ioutil.WriteFile(linkBackupFile, []byte(""), 0644)
}
