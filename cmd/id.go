package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
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
		var idsInfo []tIDInfo
		if len(args) == 0 {
			idsInfo = append(idsInfo, getEmptyID())
		} else {
			idsInfo = parseAddIDArgs(args)
		}
		for _, idInfo := range idsInfo {
			if idInfo.Type == "github" && idInfo.Token == "" {
				tokenData := ensureGithubToken(idInfo)
				idInfo.Password = ""
				idInfo.Token = tokenData.Token
			}
			log.Println(idInfo)
		}

		// idLists := getCurrentIdentityLists()

		// // write token to file
		// mainViper.Set("github-token", respTokenData.Token)
		// mainViper.Set("github-token-id", respTokenData.Id)
		// mainViper.Set("github-username", githubUsername)
	},
}

// addPublicKeyToGithub(string(publicKeyBytes), mainViper.GetString("github-token"))
// if mainViper.GetBool("verbose") {
// 	log.Println("ssh key pair created.")
// }

func getEmptyID() tIDInfo {
	var acctInfo tIDInfo
	return getIDParts(acctInfo)
}

func parseAddIDArgs(args []string) (idsInfo []tIDInfo) {
	for _, arg := range args {
		curAcctInfo := parseIDString(arg)
		idsInfo = append(idsInfo, getIDParts(curAcctInfo))
	}
	return idsInfo
}

func getIDParts(acctInfo tIDInfo) tIDInfo {
	if acctInfo.Type == "" {
		userInput := readInput("Please enter id type (github, gitlab, digitalocean, etc.): ")
		acctInfo.Type = getValidIDLngName(strings.ToLower(userInput))
	}
	titleType := strings.Title(acctInfo.Type)
	if acctInfo.Username == "" {
		acctInfo.Username = readInput(titleType + " username: ")
	}
	if acctInfo.Token == "" && acctInfo.Password == "" {
		if acctInfo.Type == "github" {
			fmt.Println("Note: A token will automatically be generated using your pasword for github accounts.")
		} else if acctInfo.Type == "gitlab" {
			fmt.Println("Note: gitlab passwords will be stored as plain text.\nYou can generate a token here: https://gitlab.com/profile/personal_access_tokens")
		}
		userInput, err := speakeasy.Ask(titleType + " " +
			acctInfo.Username + " password or token: ")
		check(err)
		acctInfo = getTokenOrPassword(userInput, acctInfo)
	}
	return acctInfo
}

type tIDInfo struct {
	Type     string
	Username string
	Password string
	Token    string
	TokenID  string
}

func parseIDString(idStr string) tIDInfo {
	var acctInfo tIDInfo
	nextStr := ""
	if strings.Contains(idStr, "/") {
		splitVals := strings.Split(idStr, "/")
		acctInfo.Type = getValidIDLngName(splitVals[0])
		nextStr = splitVals[1]
	} else {
		acctInfo.Type = getValidIDLngName(idStr)
	}

	if nextStr != "" {
		if strings.Contains(nextStr, ":") {
			splitVals := strings.Split(nextStr, ":")
			acctInfo.Username = splitVals[0]
			acctInfo = getTokenOrPassword(splitVals[1], acctInfo)
		} else {
			acctInfo.Username = nextStr
		}
	}
	return acctInfo
}

func getTokenOrPassword(arg string, acctInfo tIDInfo) tIDInfo {
	if acctInfo.Type == "github" {
		if checkIsGithubToken(arg) {
			acctInfo.Token = arg
		} else {
			acctInfo.Password = arg
		}
	} else if acctInfo.Type == "gitlab" {
		if checkIsGitlabToken(arg) {
			acctInfo.Token = arg
		} else {
			acctInfo.Password = arg
		}
	}
	return acctInfo
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
You can add the identity of an account (github, gitlab, etc.) with the command  "zetup id add [USER_INFORMATION]".

Where user information is in the form "[ID_TYPE]/[USERNAME]:[PASSWORD]", for instance "github/sam_clem:nJim&9024". You can pass multiple accounts at once, like "zetup id add [USER_INFORMATION] [USER_INFORMATION]....

You can omit any part starting from the right. You couldn't say, provide a password and not a username, but you could provide a username and no password.

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

var idFile string
var idAddDontAddSSH bool

func init() {
	idFile = filepath.Join(zetupDir, "identities.yml")
	rootCmd.AddCommand(idCmd)
	idCmd.AddCommand(idAddCmd)
	idAddCmd.Flags().BoolVarP(&idAddDontAddSSH, "no-ssh", "", false, "don't add ssh key to account")
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

type tTokenFailureData struct {
	Errors []map[string]string `json:"errors"`
}

var overrideIDGHTokenNumber = 0

func ensureGithubToken(acctInfo tIDInfo) TokenInfo {
	installID := mainViper.GetString("installation-id")
	if overrideIDGHTokenNumber != 0 {
		installID += "-" + strconv.Itoa(overrideIDGHTokenNumber)
	}
	// send token request
	data := TokenPayload{
		Note: installID,
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

	req.SetBasicAuth(acctInfo.Username, acctInfo.Password)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	if !(resp.StatusCode >= 200 && resp.StatusCode <= 299) {
		var errorsObj tTokenFailureData
		b, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(b, &errorsObj)
		if len(errorsObj.Errors) > 0 {
			errorCode := errorsObj.Errors[0]["code"]
			if errorCode == "already_exists" {
				overrideIDGHTokenNumber++
				return ensureGithubToken(acctInfo)
			}
		} else {
			log.Printf("resp.StatusCode = %+v\n", resp.StatusCode)
			log.Fatal(string(b))
		}
	}

	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var respTokenData TokenInfo
	err = decoder.Decode(&respTokenData)
	if err != nil {
		log.Fatal(err)
	}

	return respTokenData
}
