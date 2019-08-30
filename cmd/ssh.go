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

	"github.com/spf13/cobra"
	"github.com/zetup-sh/zetup/cmd/util"
)

// sshCmd represents the ssh command
var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "manage ssh keys",
	// 	Long: `A longer description that spans multiple lines and likely contains examples
	// and usage of using your command. For example:

	// Cobra is a CLI library for Go that empowers applications.
	// This application is a tool to generate the needed files
	// to quickly create a Cobra application.`,
	// Run: func(cmd *cobra.Command, args []string) {
	// 	fmt.Println("ssh called")
	// },
}

var sshGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "generate ssh key for zetup",
	Run: func(cmd *cobra.Command, args []string) {
		publicKeyFile := mainViper.GetString("public-key-file")
		privateKeyFile := mainViper.GetString("private-key-file")
		if mainViper.GetString("ssh-key-id") != "" {
			if _, err := os.Stat(publicKeyFile); !os.IsNotExist(err) {
				if _, err := os.Stat(privateKeyFile); !os.IsNotExist(err) {
					if verbose {
						fmt.Println("ssh key " + publicKeyFile + ", " + privateKeyFile + " already exists.")
						os.Exit(1)
					}
				}
			}
		}
		ensureSSHKey()
	},
}

type sshKeyInfo struct {
	ID int `json:"id"`
}

func init() {
	rootCmd.AddCommand(sshCmd)
	sshCmd.AddCommand(sshGenerateCmd)
}

func ensureSSHKey() {
	publicKeyFile := mainViper.GetString("public-key-file")
	privateKeyFile := mainViper.GetString("private-key-file")
	if mainViper.GetString("ssh-key-id") != "" {
		if _, err := os.Stat(publicKeyFile); !os.IsNotExist(err) {
			if _, err := os.Stat(privateKeyFile); !os.IsNotExist(err) {
				return
			}
		}
	}
	if verbose {
		log.Println("creating ssh key pair...")
	}

	bitSize := 4096

	privateKey, err := util.GeneratePrivateKey(bitSize)
	if err != nil {
		log.Fatal(err.Error())
	}

	publicKeyBytes, err := util.GeneratePublicKey(&privateKey.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}

	privateKeyBytes := util.EncodePrivateKeyToPEM(privateKey)

	privateSSHDir := privateKeyFile
	if !exists(privateSSHDir) {
		os.MkdirAll(privateSSHDir, 0755)
	}

	publicSSHDir := publicKeyFile
	if !exists(publicSSHDir) {
		os.MkdirAll(publicSSHDir, 0755)
	}

	err = util.WriteKeyToFile(privateKeyBytes, privateKeyFile)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = util.WriteKeyToFile([]byte(publicKeyBytes), publicKeyFile)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func getSSHPubKey() string {
	publicKeyFile := mainViper.GetString("public-key-file")
	if exists(publicKeyFile) {
		dat, err := ioutil.ReadFile(publicKeyFile)
		check(err)
		return string(dat)
	}

	ensureSSHKey()
	return getSSHPubKey()
}
