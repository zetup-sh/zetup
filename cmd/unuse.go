package cmd

import (
	"fmt"
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
	if exists(linkBackupFile) {
		dat, err := ioutil.ReadFile(linkBackupFile)
		check(err)

		var backedupFiles []BackupFileInfo
		yaml.Unmarshal(dat, &backedupFiles)
		for _, backedupFile := range backedupFiles {
			restoreLinkFile(backedupFile)
		}
	}
	ioutil.WriteFile(linkBackupFile, []byte(""), 0644)
}

func restoreLinkFile(fi BackupFileInfo) {
	if fi.SymSource == "" {
		err = os.Remove(fi.Location)
		check(err)
		ioutil.WriteFile(fi.Location, []byte(fi.Contents), 0644)
	} else {
		os.Symlink(fi.SymSource, fi.Location)
	}
}
