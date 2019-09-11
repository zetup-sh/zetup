package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var idUseCmd = &cobra.Command{
	Use:   "use",
	Short: "use identities",
	Long:  getIDAddLngUsage(),
	Run: func(cmd *cobra.Command, args []string) {
		var idsInfo []tIDInfo
		if len(args) == 0 {
			idsInfo = append(idsInfo, getEmptyID())
		} else {
			idsInfo = parseAddIDArgs(args)
		}
		idsInitialize(idsInfo)
	},
}

func getIDAddLngUsage() string {
	idAddLngUsage := `
You can use the identity of an account (github, gitlab, etc.) with the command  "zetup id use USER_INFORMATION]".

Where user information is in the form "[ID_TYPE]/[USERNAME]:[PASSWORD]", for instance "github/sam_clem:nJim&9024". You can pass multiple accounts at once, like "zetup id add [USER_INFORMATION] [USER_INFORMATION]....

You can omit any part starting from the right. You couldn't say, provide a password and not a username, but you could provide a username and no password.

You can also just type "zetup id add" and follow the prompts.

You can also use aliases the id types. The aliases are as follows:
`
	aliasUsage := ""
	for name, aliases := range possibleIDTypes {
		aliasUsage += name + ":\n"
		for _, alias := range aliases {
			aliasUsage += "  " + alias + "\n"
		}
	}
	return idAddLngUsage + aliasUsage
}

var possibleIDTypes = map[string][]string{
	"github":       []string{"github.com", "gh"},
	"gitlab":       []string{"gitlab.com", "gl"},
	"digitalocean": []string{"digitalocean.com", "do"},
}

var idFile string
var idAddAddSSH bool
var idAddOverwrite bool
var idAddGHToken bool

func getEmptyID() tIDInfo {
	var idInfo tIDInfo
	return getIDParts(idInfo)
}

func parseAddIDArgs(args []string) (idsInfo []tIDInfo) {
	for _, arg := range args {
		curIDInfo := parseIDString(arg)
		idsInfo = append(idsInfo, getIDParts(curIDInfo))
	}
	return idsInfo
}

func getIDParts(idInfo tIDInfo) tIDInfo {
	if idInfo.Type == "" {
		userInput := readInput("Please enter id type (github, gitlab, digitalocean, etc.): ")
		idInfo.Type = getValidIDLngName(strings.ToLower(userInput))
	}
	titleType := strings.Title(idInfo.Type)
	if idInfo.Username == "" {
		idInfo.Username = readInput(titleType + " username: ")
	}
	if idInfo.Password == "" {
		if idInfo.Type == "github" {
			fmt.Println("Note: A token will automatically be generated using your password for github accounts. You can override this with `--gh-token=false`")
		} else if idInfo.Type == "gitlab" {
			fmt.Println("Note: gitlab passwords will be stored as plain text.\nYou can generate a token here: https://gitlab.com/profile/personal_access_tokens")
		}
		prompt := promptui.Prompt{
			Label: titleType + " " + idInfo.Username + " password or token: ",
			Mask:  '*',
		}

		userInput, err := prompt.Run()
		if err != nil {
			fmt.Println("signal interrupt detected")
			os.Exit(1)
		}
		idInfo.Password = userInput
	}
	return idInfo
}

func parseIDString(idStr string) tIDInfo {
	var idInfo tIDInfo
	nextStr := ""
	if strings.Contains(idStr, "/") {
		splitVals := strings.Split(idStr, "/")
		idInfo.Type = getValidIDLngName(splitVals[0])
		nextStr = splitVals[1]
	} else {
		idInfo.Type = getValidIDLngName(idStr)
	}

	if nextStr != "" {
		if strings.Contains(nextStr, ":") {
			splitVals := strings.Split(nextStr, ":")
			idInfo.Username = splitVals[0]
			idInfo.Password = splitVals[1]
		} else {
			idInfo.Username = nextStr
		}
	}
	return idInfo
}

func checkIsGithubToken(txt string) bool {
	isghtoken, err := regexp.MatchString("^[A-Za-z0-9]{40}$", txt)
	if err != nil {
		log.Fatal("There was a problem checking if you provided a github token or not. This should never occur.")
	}
	return isghtoken
}

func checkIsGitlabToken(txt string) bool {
	isgltoken, err := regexp.MatchString("^[A-Za-z0-9]{20}$", txt)
	if err != nil {
		log.Fatal("There was a problem checking if you provided a gitlab token or not. This should never occur.")
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
	return idType
}

func ensurePublicKeyGithub(idInfo tIDInfo) {
	if verbose {
		log.Println("ensuring public key added to github")
	}
	installID := mainViper.GetString("installation-id")
	endPoint := "https://api.github.com/user/keys"
	pubKey := getSSHPubKey()
	body := strings.NewReader(fmt.Sprintf(`{
				"title": "%v",
				"key": "%v"
			}`, installID, strings.TrimRight(pubKey, "\n")))
	req, err := http.NewRequest("POST", endPoint, body)
	if err != nil {
		log.Fatal(err)
	}

	req.SetBasicAuth(idInfo.Username, idInfo.Password)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	if !(resp.StatusCode >= 200 && resp.StatusCode <= 299) {
		b, _ := ioutil.ReadAll(resp.Body)
		var errorsObj githubFailureData
		json.Unmarshal(b, &errorsObj)
		if len(errorsObj.Errors) > 0 {
			errorMessage := errorsObj.Errors[0]["message"]
			if errorMessage == "key is already in use" {
				if verbose {
					log.Println("already have that key")
				}
				return
			}
		} else {
			log.Printf("resp.StatusCode = %+v\n", resp.StatusCode)
			log.Fatal(string(b))
		}
	}
	if verbose {
		log.Println("Sucessfully added key to github")
	}

	defer resp.Body.Close()
}

func ensureGithubToken(idInfo tIDInfo) tTokenInfo {
	installID := mainViper.GetString("installation-id")
	// send token request
	data := tTokenPayload{
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

	req.SetBasicAuth(idInfo.Username, idInfo.Password)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	if !(resp.StatusCode >= 200 && resp.StatusCode <= 299) {
		var errorsObj githubFailureData
		b, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(b, &errorsObj)
		if len(errorsObj.Errors) > 0 {
			errorCode := errorsObj.Errors[0]["code"]
			if errorCode == "already_exists" {
				deleteGHAuthByName(idInfo, installationID)
				return ensureGithubToken(idInfo)
			}
		} else {
			log.Printf("resp.StatusCode = %+v\n", resp.StatusCode)
			log.Fatal(string(b))
		}
	}

	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var respTokenData tTokenInfo
	err = decoder.Decode(&respTokenData)
	if err != nil {
		log.Fatal(err)
	}

	return respTokenData
}

type githubFailureData struct {
	Errors []map[string]string `json:"errors"`
}
type gitlabMessageJSON struct {
	Message map[string]map[string][]string `json:"message"`
}

func ensurePublicKeyGitlab(idInfo tIDInfo) {
	if verbose {
		log.Println("ensuring gitlab public key")
	}
	installID := mainViper.GetString("installation-id")
	glTemporaryToken := getTemporaryGitlabToken(idInfo)
	pubKey := getSSHPubKey()
	endPoint := "https://gitlab.com/api/v4/user/keys"
	body := strings.NewReader(fmt.Sprintf(`{
				"title": "%v",
				"key": "%v"
			}`, installID, strings.TrimRight(pubKey, "\n")))
	req, err := http.NewRequest("POST", endPoint, body)
	if err != nil {
		log.Fatal(err)
	}

	// req.SetBasicAuth(idInfo.Username, idInfo.Password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+glTemporaryToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := ioutil.ReadAll(resp.Body)
	if !(resp.StatusCode >= 200 && resp.StatusCode <= 299) {
		if resp.StatusCode == 400 {
			if verbose {
				log.Println("gitlab fingerprint already added")
			}
			return
		}
		log.Printf("resp.StatusCode = %+v\n", resp.StatusCode)
		log.Fatal(string(b))
	}
}

func getTemporaryGitlabToken(idInfo tIDInfo) string {
	endPoint := "https://gitlab.com/oauth/token"
	body := strings.NewReader(fmt.Sprintf(`{
	"grant_type" : "password",
	"username": "%s",
	"password": "%s"
}`, idInfo.Username, idInfo.Password))
	req, err := http.NewRequest("POST", endPoint, body)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	b, _ := ioutil.ReadAll(resp.Body)
	if !(resp.StatusCode >= 200 && resp.StatusCode <= 299) {
		log.Printf("resp.StatusCode = %+v\n", resp.StatusCode)
		log.Fatal(string(b))
	}
	var respTokenData tTokenInfo
	err = json.Unmarshal(b, &respTokenData)
	if err != nil {
		log.Fatal(err)
	}

	return respTokenData.AccessToken
}

func getUserInfoFromGitlab(idInfo tIDInfo) {
	viperUserInfo := mainViper.GetStringMapString("user")
	userInfo.Email = viperUserInfo["email"]
	userInfo.Name = viperUserInfo["name"]
	gitlabUsername := mainViper.GetString("gitlab.username")

	if gitlabUsername == "" || userInfo.Name != "" || userInfo.Email != "" {
		return
	}
	glTemporaryToken := getTemporaryGitlabToken(idInfo)
	endPoint := "https://gitlab.com/api/v4/user"
	req, err := http.NewRequest("GET", endPoint, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+glTemporaryToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if !(resp.StatusCode >= 200 && resp.StatusCode <= 299) {
		b, _ := ioutil.ReadAll(resp.Body)
		log.Printf("resp.StatusCode = %+v\n", resp.StatusCode)
		log.Println(string(b))
		log.Println("We could not retrieve your user information from gitlab. This is a non fatal error")
	}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&userInfo)
	if err != nil {
		b, _ := ioutil.ReadAll(resp.Body)
		log.Printf("resp.StatusCode = %+v\n", resp.StatusCode)
		log.Println(string(b))
		log.Fatal(err)
	}

	// write token to file
	mainViper.Set("user.name", userInfo.Name)
	mainViper.Set("user.email", userInfo.Email)
}

func getUserInfoFromGithub() {
	viperUserInfo := mainViper.GetStringMapString("user")
	userInfo.Email = viperUserInfo["email"]
	userInfo.Name = viperUserInfo["name"]

	username := mainViper.GetString("github.username")
	password := mainViper.GetString("github.password")

	if username == "" || userInfo.Name != "" || userInfo.Email != "" {
		return
	}

	// get info with personal access token
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		log.Fatal(err)
	}
	req.SetBasicAuth(username, password)

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
	err = decoder.Decode(&userInfo)
	if err != nil {
		log.Fatal(err)
	}

	// write token to file
	mainViper.Set("user.name", userInfo.Name)
	mainViper.Set("user.email", userInfo.Email)
}

func writeGitConfig() {
	gitConfigFile := fmt.Sprintf(`[user]
	name = %v
	email = %v
`, mainViper.Get("user.name"), mainViper.Get("user.email"))
	home, _ := os.UserHomeDir()
	gitconfigFile := path.Join(home, ".gitconfig")
	if !exists(gitconfigFile) {
		_ = ioutil.WriteFile(gitconfigFile, []byte(gitConfigFile), 0644)
	}
}

type tUserInfo struct {
	GithubUsername string `json:"login"`
	Email          string `json:"email"`
	Name           string `json:"name"`
}

var userInfo tUserInfo

type tTokenInfo struct {
	Token       string `json:"token"`
	AccessToken string `json:"access_token"`
}

type tTokenPayload struct {
	Note   string   `json:"note"`
	Scopes []string `json:"scopes"`
}