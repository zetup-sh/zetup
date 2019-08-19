package cmd

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bgentry/speakeasy"
	"github.com/spf13/cobra"
	"github.com/zetup-sh/zetup/cmd/util"
	"gopkg.in/yaml.v2"
)

// idCmd represents the id command
var idCmd = &cobra.Command{
	Use:   "id",
	Short: "Manage identities for zetup",
	// Long:  `adds  identities`,
	Run: func(cmd *cobra.Command, args []string) {
		idLists := getCurrentIdentityLists()
		log.Println(idLists)
	},
}
var idAddCmd = &cobra.Command{
	Use:   "add",
	Short: "add identities",
	Long:  getIDAddLngUsage(),
	Run: func(cmd *cobra.Command, args []string) {
		getIDParts(args)
		log.Println("You entered ", idType, idUsername, idPassword)
	},
}

func getIDParts(args []string) {
	if len(args) == 1 {
		nextStr := ""
		if strings.Contains(args[0], "/") {
			splitVals := strings.Split(args[0], "/")
			idType = getValidIDLngName(splitVals[0])
			nextStr = splitVals[1]
		} else {
			idType = getValidIDLngName(args[0])
		}

		if nextStr != "" {
			if strings.Contains(nextStr, ":") {
				splitVals := strings.Split(nextStr, ":")
				idUsername = splitVals[0]
				getTokenOrPassword(splitVals[1])
			} else {
				idUsername = nextStr
			}
		}
	} else if len(args) == 2 {
		idType = args[0]
		idUsername = args[1]
	} else if len(args) == 3 {
		idType = args[0]
		idUsername = args[1]
		getTokenOrPassword(args[2])
	}
	if idType == "" {
		userInput := readInput("Please enter id type (github, gitlab, digitalocean, etc.): ")
		idType = getValidIDLngName(userInput)
	}
	if idUsername == "" {
		idUsername = readInput(idType + " username: ")
	}
	if idToken == "" && idPassword == "" {
		if idType == "github" {
			fmt.Println("Note: A token will automatically be generated using your pasword for github accounts.")
		}
		userInput, err := speakeasy.Ask(idType + " password or token: ")
		check(err)
		getTokenOrPassword(userInput)
	}
}

func getTokenOrPassword(arg string) {
	if idType == "github" {
		if checkIsGithubToken(arg) {
			idToken = arg
		} else {
			idPassword = arg
		}
	} else if idType == "gitlab" {
		if checkIsGitlabToken(arg) {
			idToken = arg
		} else {
			idPassword = arg
		}
	}
}

func readInput(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf(prompt)
	enteredData, err := reader.ReadString('\n')
	trimmedData := strings.TrimSpace(enteredData)
	if err != nil {
		log.Fatal(err)
	}
	return trimmedData
}

func checkIsGithubToken(txt string) bool {
	isghtoken, err := regexp.MatchString("^[A-Za-z0-9]{40}$", txt)
	if err != nil {
		log.Fatal("There was a problem checking if you provided a github token or not.")
	}
	return isghtoken
}

func checkIsGitlabToken(txt string) bool {
	isgltoken, err := regexp.MatchString("^[A-Za-z0-9]{40}$", txt)
	if err != nil {
		log.Fatal("There was a problem checking if you provided a gitlab token or not.")
	}
	return isgltoken
}

func getValidIDLngName(idType string) string {
	for name, aliases := range possibleIDTypes {
		if name == idType {
			return name
		}
		for _, alias := range aliases {
			if idType == alias {
				return name
			}
		}
	}
	log.Fatalln(getIDAddLngUsage())
	return ""
}

func getIDAddLngUsage() string {
	idAddLngUsage := `
You can add the identity of an account (github, gitlab, etc.) with the command  "zetup id add [user information]".

You can pass the user information to the command with the flags, or you can pass it in the form of an argument like "zetup id add github.com/username:{token or password}" or just "zetup id add github.com username {token or password}"

You can also just type "zetup id add" and follow the prompts.

You can also use aliases the id types. The aliases are as follows:
`
	aliasUsage := ""
	for name, aliases := range possibleIDTypes {
		aliasUsage += name + ": "
		for _, alias := range aliases {
			aliasUsage += "  " + alias + "\n"
		}
	}
	return idAddLngUsage + aliasUsage
}

var possibleIDTypes = map[string][]string{
	"github":       []string{"github.com", "gh"},
	"gitlab":       []string{"gitlab.com", "gl"},
	"digitalocean": []string{"do"},
}

var idType string
var idUsername string
var idPassword string
var idToken string
var idFile string

func init() {
	idFile = filepath.Join(zetupDir, "identities.yml")
	rootCmd.AddCommand(idCmd)
	idCmd.PersistentFlags().StringVarP(&idUsername, "username", "u", "", "username for github")
	idCmd.PersistentFlags().StringVarP(&idPassword, "password", "p", "", "password for github")
	idCmd.PersistentFlags().StringVarP(&idToken, "token", "", "", "token for provider (will create if possible)")
	idCmd.PersistentFlags().StringVarP(&idType, "type", "t", "", "token for provider (will create if possible)")
	idCmd.AddCommand(idAddCmd)
}

type tIDLists struct {
	Github       []map[string]string `yaml:"github"`
	Gitlab       []map[string]string `yaml:"gitlab"`
	DigitalOcean []map[string]string `yaml:"digitalocean"`
}

func getCurrentIdentityLists() tIDLists {
	var IDLists tIDLists
	if util.Exists(idFile) {
		dat, err := ioutil.ReadFile(idFile)
		check(err)
		yaml.Unmarshal(dat, &IDLists)
		return IDLists
	}
	return IDLists
}
