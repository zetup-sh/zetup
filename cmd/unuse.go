package cmd

import (
	"io/ioutil"
	"os"
	"path"
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
		Unuse()
	},
}

func init() {
	rootCmd.AddCommand(unuseCmd)
}

func Unuse() {
	RestoreBackupFiles()
	unuseFile, err := FindFile(usePkgDir, "unuse", runtime.GOOS, LINUX_EXTENSIONS, mainViper)
	if err == nil {
		runFile(unuseFile)
	}
}

func RestoreBackupFiles() {
	files, err := ioutil.ReadDir(bakDir)
	check(err)
	var bakupFiles []string
	for _, file := range files {
		bakupFiles = append(bakupFiles, path.Join(bakDir, file.Name()))
	}

	for _, bakupFile := range bakupFiles {
		// restore backups of linked files
		dat, err := ioutil.ReadFile(bakupFile)
		check(err)
		var backedupFiles []BackupFileInfo
		yaml.Unmarshal(dat, &backedupFiles)
		for _, backedupFile := range backedupFiles {
			err = os.Remove(backedupFile.Location)
			check(err)
			ioutil.WriteFile(backedupFile.Location, []byte(backedupFile.Contents), 0644)
		}
	}
}
