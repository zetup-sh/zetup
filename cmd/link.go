package cmd

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/zetup-sh/zetup/cmd/util"
	"gopkg.in/yaml.v2"
)

var linkNoOverwrite bool

// linkCmd represents the link command
var linkCmd = &cobra.Command{
	Use:   "link",
	Short: "links a file and backs up the current file",
	Args:  cobra.ExactArgs(2),
	Long: `The first argument is the target (or where you want to link the file), and the second file is the source (the file that is being linked to). This will over write any file that already exists, but will back it up. Then, when you call "zetup use" or "zetup unuse", it will unlink your file and restore the backup.

	This is helpful when you want to, for instance, link a .bashrc file, but restore the original when the user switches the package.

	Note: if you have already linked a file, the backup will not be overwritten. So, it is safe to call more than once.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		target := args[0]
		source := args[1]
		linkFile(target, source)
	},
}

var linkVerbose bool

func init() {
	rootCmd.AddCommand(linkCmd)
	linkCmd.Flags().BoolVarP(&linkNoOverwrite, "no-overwrite", "n", false, "don't overwrite existing target files")
	linkCmd.Flags().BoolVarP(&linkVerbose, "verbose", "v", false, "tell you if it's ignoring overwrite")
}

func linkFile(source string, target string) {
	if !filepath.IsAbs(source) {
		log.Fatal("source file '" + source + "' must be absolute")
	}

	if !filepath.IsAbs(source) {
		log.Fatal("source file '" + source + "' must be absolute")
	}

	// back up file if it exists
	if util.Exists(target) {
		var backedupFiles []BackupFileInfo

		// get current backups
		if util.Exists(linkBackupFile) {
			dat, err := ioutil.ReadFile(linkBackupFile)
			check(err)
			yaml.Unmarshal(dat, &backedupFiles)
		}
		for _, backedupFile := range backedupFiles {
			if backedupFile.Location == target {
				if linkVerbose {
					log.Println("file already backed up " + target)
				}
				return
			}
		}

		// back up target
		dat, err := ioutil.ReadFile(target)
		check(err)
		backupFileInfo := BackupFileInfo{
			target,
			string(dat),
		}
		filesToBackup := append(backedupFiles, backupFileInfo)
		marshaled, err := yaml.Marshal(filesToBackup)
		check(err)

		backupWithHeader := []byte("# generated file do not edit\n" + string(marshaled))
		err = ioutil.WriteFile(linkBackupFile, backupWithHeader, 0644)
		check(err)

		// then link the actual files
		// we back up first in case something goes wrong
		err = os.Remove(target)
		check(err)
	}

	err = os.Symlink(source, target)
	check(err)
}
