package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	ssh2 "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

var pkgViper *viper.Viper

// initCmd represents the init command
var useCmd = &cobra.Command{
	Use:   "use",
	Short: "Specify a zetup package to use",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		usePkg(args[0])
	},
}

func usePkg(pkgToInstall string) {
	repo := parseRepoName(pkgToInstall)
	usePkgDir = filepath.Join(pkgDir, repo.Joined)
	ensureRepo(repo, usePkgDir, useBranch, useProtocol)

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

	mainViper.Set("cur-pkg", usePkgDir)
	mainViper.WriteConfig()

	useFile, err := findFile(usePkgDir, "use", runtime.GOOS, unixExtensions, mainViper)
	if err == nil {
		if verbose {
			vLog.Printf("running use file %s\n", useFile)
		}
		err = runFile(useFile)
	}

	if err != nil {
		log.Println("There was a problem running the use file.")
		log.Fatalln(useFile, err)
	}

}

var useBranch string
var useProtocol string

func init() {
	rootCmd.AddCommand(useCmd)
	useCmd.Flags().StringVarP(&useBranch, "branch", "b", "", "specify a branch to use, otherwise it will use default")
	useCmd.Flags().StringVarP(&useProtocol, "protocol", "p", "try-both", "protocol to use (ssh, https or try-both which tries ssh then https)")
}

var usePkgDir string
var usePkgDirParent string

type repoInfo struct {
	Hostname string
	Username string
	Reponame string
	Joined   string
}

func parseRepoName(repo string) repoInfo {
	splitPath := strings.Split(repo, "/")
	ghUsername := mainViper.GetString("github.username")
	glUsername := mainViper.GetString("gitlab.username")
	if len(splitPath) == 1 {
		if ghUsername != "" {
			splitPath = []string{"github.com", ghUsername, splitPath[0]}
		} else if glUsername != "" {
			splitPath = []string{"gitlab.com", glUsername, splitPath[0]}
		} else {
			fmt.Println("You don't have any accounts to extrapolate the full path from, so you'll have to provide the full path, e.g. github.com/zetup-sh/zetup-pkg.")
			fmt.Println("Or you can add a new account with `zetup id add`")
			os.Exit(1)
		}
	}
	if len(splitPath) != 3 || !(splitPath[0] == "github.com" || splitPath[0] == "gitlab.com") {
		fmt.Println("repos must be in the format [hostname]/[username]/reponame], e.g. github.com/zetup-sh/zetup-pkg")
		fmt.Println("Only github/gitlab are supported for now.")
		os.Exit(1)
	}
	return repoInfo{
		Hostname: splitPath[0],
		Username: splitPath[1],
		Reponame: splitPath[2],
		Joined:   filepath.Join(splitPath...),
	}
}

func ensureRepo(repo repoInfo, localPath string, branch string, protocol string) {
	if _, err := os.Stat(usePkgDir); os.IsNotExist(err) {
		err := os.MkdirAll(filepath.Dir(localPath), 0755)
		if err != nil {
			log.Fatal(err)
		}
		if verbose {
			log.Println(repo.Joined, " not found, cloning...")
		}

		var url string
		// hosttype := getValidIDLngName(repo.Hostname)
		// username := mainViper.GetString(hosttype + ".username")

		var r *git.Repository
		var cloneOptions git.CloneOptions

		if protocol == "ssh" || protocol == "try-both" {
			privateKeyFile := mainViper.GetString("private-key-file")
			pem, _ := ioutil.ReadFile(privateKeyFile)
			signer, _ := ssh.ParsePrivateKey(pem)
			auth := &ssh2.PublicKeys{
				User:   "git",
				Signer: signer,
			}
			// user should add host with ssh-keyscan
			// we do not handle that though. maybe we should?
			auth.HostKeyCallback = ssh.InsecureIgnoreHostKey()

			url = fmt.Sprintf("git@github.com:%s/%s.git", repo.Username, repo.Reponame)
			cloneOptions = git.CloneOptions{
				URL:               url,
				RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
				Auth:              auth,
			}
		} else if protocol == "https" {
			url = fmt.Sprintf("https://%s/%s/%s.git", repo.Hostname, repo.Username, repo.Reponame)
			cloneOptions = git.CloneOptions{
				URL:               url,
				RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
			}
		}

		if useBranch != "" {
			rname := fmt.Sprintf("refs/heads/%s", branch)
			cloneOptions.ReferenceName = plumbing.ReferenceName(rname)
		}
		r, err = git.PlainClone(localPath, false, &cloneOptions)
		if err != nil && protocol == "try-both" {
			ensureRepo(repo, localPath, branch, "https")
			return
		}
		check(err)

		ref, err := r.Head()
		check(err)

		_, err = r.CommitObject(ref.Hash())
		check(err)

	}
}
