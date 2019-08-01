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
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
	git "gopkg.in/src-d/go-git.v4"
	ssh2 "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

// initCmd represents the init command
var useCmd = &cobra.Command{
	Use:   "use",
	Short: "Specify a zetup package to use",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		return
		// get sudo privileges early, in case they want to do
		// something else while it's installing
		//getSudoCmd := exec.Command("sudo", "echo", "have sudo privileges")
		//err := getSudoCmd.Start()
		//if err != nil {
		//log.Fatal(err)
		//}
		//err = getSudoCmd.Wait()
		packageToInstall := args[0]
		pkgDir := viper.GetString("pkg-dir")
		splitPath := strings.Split(packageToInstall, "/")
		if len(splitPath) == 1 {
			splitPath = []string{"github.com", viper.GetString("github-username"), splitPath[0]}
		}
		if len(splitPath) != 3 || splitPath[0] != "github.com" {
			log.Fatal("Only github is supported for now.")
		}

		usePkgDir := pkgDir + string(os.PathSeparator) + path.Join(splitPath...)
		usePkgDirParent, _ := path.Split(usePkgDir)
		err := os.MkdirAll(usePkgDirParent, 0755)
		if err != nil {
			log.Fatal(err)
		}

		if _, err := os.Stat(usePkgDir); os.IsNotExist(err) {
			//log.Println(usePkgDir + " not found, cloning...")

			var url string
			username := splitPath[1]
			githubUsername := viper.GetString("github-username")
			if githubUsername != username {
				log.Printf("70 = %+v\n", 70)
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

			//commit, err := r.CommitObject(ref.Hash())
			//if err != nil {
			//log.Fatal(err)
			//}

		}

	},
}

func init() {
	rootCmd.AddCommand(useCmd)
}
