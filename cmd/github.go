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
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/bgentry/speakeasy"
	"github.com/spf13/cobra"
)

// githubCmd represents the github command
var githubCmd = &cobra.Command{
	Use:   "github",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("github called")
	},
}

var githubIDCmd = &cobra.Command{
	Use:   "id",
	Short: "manage github identities",
	// 	Long: `A longer description that spans multiple lines and likely contains examples
	// and usage of using your command. For example:

	// Cobra is a CLI library for Go that empowers applications.
	// This application is a tool to generate the needed files
	// to quickly create a Cobra application.`,
	// Run: func(cmd *cobra.Command, args []string) {
	// 	fmt.Println("github called")
	// },
}

var githubIDAddCmd = &cobra.Command{
	Use:   "add",
	Short: "add github identities",
	// 	Long: `A longer description that spans multiple lines and likely contains examples
	// and usage of using your command. For example:

	// Cobra is a CLI library for Go that empowers applications.
	// This application is a tool to generate the needed files
	// to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		githubIDAddEnsureUsername()
		githubIDAddEnsureToken()
	},
}

var githubIDAddUsername string
var githubIDAddPassword string
var githubIDAddToken string

func init() {
	rootCmd.AddCommand(githubCmd)
	githubCmd.AddCommand(githubIDCmd)
	githubIDCmd.AddCommand(githubIDAddCmd)
	githubIDAddCmd.PersistentFlags().StringVarP(&githubIDAddUsername, "username", "u", "", "username for github")
	githubIDAddCmd.PersistentFlags().StringVarP(&githubIDAddPassword, "password", "p", "", "password for github")
	githubIDAddCmd.PersistentFlags().StringVarP(&githubIDAddToken, "token", "t", "", "token for github (will create if only password provided)")
}

func githubIDAddEnsureUsername() {
	if githubIDAddUsername == "" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("Github Username (%v): ", os.Getenv("USER"))
		enteredUsername, err := reader.ReadString('\n')
		githubIDAddUsername = strings.Trim(enteredUsername, " ")
		if err != nil {
			log.Fatal(err)
		}
	}
}

func githubIDAddEnsureToken() {
	if githubIDAddToken != "" {
		return // already have token, no need to create
	}

	if githubIDAddPassword == "" {
		githubIDAddPassword, err = speakeasy.Ask("Github Password: ")
		if err != nil {
			log.Fatal(err)
		}
	}

	// send token request
	data := TokenPayload{
		Note: mainViper.GetString("installation-id"),
		Scopes: []string{
			"repo",
			"admin:org",
			"admin:public_key",
			"admin:repo_hook",
			"gist",
			"notifications",
			"user",
			"delete_repo",
			"write:discussion",
			"admin:gpg_key",
		},
	}
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", "https://api.github.com/authorizations", body)
	if err != nil {
		log.Fatal(err)
	}

	req.SetBasicAuth(githubIDAddUsername, githubIDAddPassword)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	if !(resp.StatusCode >= 200 && resp.StatusCode <= 299) {
		b, _ := ioutil.ReadAll(resp.Body)
		log.Printf("resp.StatusCode = %+v\n", resp.StatusCode)
		log.Fatal(string(b))
	}

	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var respTokenData TokenInfo
	err = decoder.Decode(&respTokenData)
	if err != nil {
		log.Fatal(err)
	}

	// write token to file
	mainViper.Set("github-token", respTokenData.Token)
	mainViper.Set("github-token-id", respTokenData.Id)
	mainViper.Set("github-username", githubUsername)
}
