package cmd

import (
	"io/ioutil"
	"os"
	"runtime"

	"github.com/spf13/cobra"
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
		runFile(unuseFile)
	}
	mainViper.Set("use-pkg", "")
	mainViper.WriteConfig()
}

func restoreLinkFiles() {
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
