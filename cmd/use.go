package cmd

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
	git "gopkg.in/src-d/go-git.v4"
	ssh2 "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

var pkgViper *viper.Viper

// initCmd represents the init command
var useCmd = &cobra.Command{
	Use:   "use",
	Short: "Specify a zetup package to use",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		pkgToInstall := args[0]
		ensureRepo(pkgToInstall)

		pkgViper = viper.New()
		pkgViper.AddConfigPath(usePkgDir)
		pkgViper.SetConfigName("config")
		if err := pkgViper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				log.Println("Your package must contain a config.yml")
				log.Println("Tip: You can use `zetup generate` to create a skeleton project or `zetup fork` to fork your favorite zetup package.")
			} else {
				log.Printf("err = %+v\n", err)
			}
		}

		unuse()
		useFile, err := FindFile(usePkgDir, "use", runtime.GOOS, unixExtensions, mainViper)
		if err == nil {
			runFile(useFile)
		}

		mainViper.Set("cur-pkg", usePkgDir)
		mainViper.WriteConfig()
	},
}

func init() {
	rootCmd.AddCommand(useCmd)
}

var usePkgDir string
var usePkgDirParent string

func ensureRepo(repo string) {
	splitPath := strings.Split(repo, "/")
	if len(splitPath) == 1 {
		splitPath = []string{"github.com", mainViper.GetString("github-username"), splitPath[0]}
	}
	if len(splitPath) != 3 || splitPath[0] != "github.com" {
		log.Fatal("Only github is supported for now.")
	}

	usePkgDir = pkgDir + string(os.PathSeparator) + path.Join(splitPath...)
	usePkgDirParent, _ = path.Split(usePkgDir)
	err := os.MkdirAll(usePkgDirParent, 0755)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := os.Stat(usePkgDir); os.IsNotExist(err) {
		if mainViper.GetBool("verbose") {
			log.Println(path.Join(splitPath...) + " not found, cloning...")
		}

		var url string
		username := splitPath[1]
		githubUsername := mainViper.GetString("github-username")
		var r *git.Repository
		if githubUsername != username {
			url = "https://github.com/" + username + "/" + splitPath[2] + ".git"
			r, err = git.PlainClone(usePkgDir, false, &git.CloneOptions{
				URL:               url,
				RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
			})
		} else {
			privateKeyFile := mainViper.GetString("private-key-file")

			pem, _ := ioutil.ReadFile(privateKeyFile)
			signer, _ := ssh.ParsePrivateKey(pem)
			auth := &ssh2.PublicKeys{User: "git", Signer: signer}
			url = "git@github.com:" + username + "/" + splitPath[2] + ".git"
			r, err = git.PlainClone(usePkgDir, false, &git.CloneOptions{
				URL:               url,
				RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
				Auth:              auth,
			})
		}

		if err != nil {
			log.Fatal(err)
		}
		_, err = r.Head()
		if err != nil {
			log.Fatal(err)
		}
	}
}
