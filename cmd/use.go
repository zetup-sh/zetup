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
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
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
var pkgToInstall string

type LinuxInfo struct {
	Distro, Arch, Release, CodeName string
}

// initCmd represents the init command
var useCmd = &cobra.Command{
	Use:   "use",
	Short: "Specify a zetup package to use",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		pkgToInstall = args[0]
		ensureRepo()

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

		// install linux
		if runtime.GOOS == "linux" {
			var linuxInfo LinuxInfo
			linuxInfo.Distro = getSystemInfo("lsb_release", "-ds", "distro")
			if linuxInfo.Distro == "disco" && viper.GetBool("verbose") {
				log.Println("Disco ain't dead") // easter egg, idk
			}
			linuxInfo.Release = getSystemInfo("lsb_release", "-rs", "release")
			linuxInfo.CodeName = getSystemInfo("lsb_release", "-cs", "release")
			linuxInfo.Arch = getSystemInfo("uname", "-m", "architecture")

			// check if there are any apt packages not already installed
			aptPackages := pkgViper.GetStringSlice("apt")

			aptPackagesAlreadyInstalled := viper.GetStringMap("installed-apt")
			var toInstall []string
			for _, pkg := range aptPackages {
				if aptPackagesAlreadyInstalled[pkg] == nil {
					toInstall = append(toInstall, pkg)
				}
			}

			if len(toInstall) > 0 {
				if viper.GetBool("verbose") {
					log.Printf("updating apt\n")
				}
				runCmd := exec.Command("sudo", "apt-get", "update", "-yqq")
				runCmd.Stdout = os.Stdout
				runCmd.Stdin = os.Stdin
				runCmd.Stderr = os.Stderr
				err := runCmd.Run()
				if err != nil {
					log.Println("Could not run apt-get update")
					log.Fatal(err)
				}

				if viper.GetBool("verbose") {
					log.Printf("installing %+v using apt\n", toInstall)
				}
				cmdArgs := append([]string{"apt-get", "install", "-yqq"}, toInstall...)
				runCmd = exec.Command("sudo", cmdArgs...)
				runCmd.Stdout = os.Stdout
				runCmd.Stdin = os.Stdin
				runCmd.Stderr = os.Stderr
				err = runCmd.Run()
				if err != nil {
					log.Println("Could not run apt-get install")
					log.Fatal(err)
				}
				for _, pkg := range toInstall {
					viper.Set("installed-apt."+pkg, true)
				}
				viper.WriteConfig()
			}
		}

		os.Exit(0)
		// todo: init only if not current
		runAnyFile("init")

		//if [[ $@ != "--no-star" ]];
		//then
		//curl -s -u $USERNAME:$ZETUP_GITHUB_TOKEN \
		//--request PUT \
		//https://api.github.com/user/starred/zwhitchcox/zetup-config;
		//fi
	},
}

func getSystemInfo(bashcmd string, flags string, name string) string {
	out, err := exec.Command(bashcmd, flags).Output()
	if err != nil {
		os.Stderr.WriteString("Could not read " + name + " from " + bashcmd)
		if err != nil {
			log.Fatal(err)
		}
	}
	return string(out)
}

func init() {
	rootCmd.AddCommand(useCmd)
}

var usePkgDir string
var usePkgDirParent string

func ensureRepo() {
	splitPath := strings.Split(pkgToInstall, "/")
	if len(splitPath) == 1 {
		splitPath = []string{"github.com", viper.GetString("github-username"), splitPath[0]}
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
		if viper.GetBool("verbose") {
			log.Println(path.Join(splitPath...) + " not found, cloning...")
		}

		var url string
		username := splitPath[1]
		githubUsername := viper.GetString("github-username")
		if githubUsername != username {
			url = "https://github.com/" + username + "/" + splitPath[2] + ".git"
		} else {
			url = "git@github.com:" + username + "/" + splitPath[2] + ".git"
		}

		privateKeyFile := viper.GetString("private-key-file")

		pem, _ := ioutil.ReadFile(privateKeyFile)
		signer, _ := ssh.ParsePrivateKey(pem)
		auth := &ssh2.PublicKeys{User: "git", Signer: signer}

		r, err := git.PlainClone(usePkgDir, false, &git.CloneOptions{
			URL:               url,
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
			Auth:              auth,
		})
		if err != nil {
			log.Fatal(err)
		}
		_, err = r.Head()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func runAnyFile(fileSuffix string) {
	cmdFile := pkgViper.GetString(runtime.GOOS + "-" + fileSuffix)
	if cmdFile == "" {
		log.Fatal("This zetup package does not contain a `" + fileSuffix + "` file for your OS (" + runtime.GOOS + ")")
	}
	cmdFilePath := usePkgDir + string(os.PathSeparator) + cmdFile
	if runtime.GOOS == "linux" {
	}
	runBashFile(cmdFilePath)
}

func runBashFile(filePath string) {
	err := os.Chmod(filePath, 0755)
	if err != nil {
		fmt.Println(err)
	}
	runCmd := exec.Command("/bin/sh", filePath)
	runCmd.Stdout = os.Stdout
	runCmd.Stdin = os.Stdin
	runCmd.Stderr = os.Stderr
	err = runCmd.Run()
	if err != nil {
		log.Fatalf("%s %s\n", filePath, err)
	}
}
