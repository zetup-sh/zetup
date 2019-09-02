package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

var idDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "remove identity tokens/keys",
	Long:  getIDAddLngUsage(),
	Run: func(cmd *cobra.Command, args []string) {
		gitlabInfo := getGitlabInfo()
		githubInfo := getGithubInfo()
		toDelete := args
		if idDeleteAll {
			toDelete = append(toDelete, "github", "gitlab")
		}
		for _, arg := range args {
			normalized := getValidIDLngName(arg)
			if normalized == "gitlab" {
				deletePublicKeyGitlab(gitlabInfo)
				mainViper.Set("gitlab", "")
				mainViper.WriteConfig()
			} else if normalized == "github" {
				deletePublicKeyGithub(githubInfo)
				deleteAuthGithub(githubInfo)
				mainViper.Set("github", "")
				mainViper.WriteConfig()
			}
		}
		removeLineFromConfig("gitlab: \"\"")
		removeLineFromConfig("github: \"\"")
	},
}

func getGitlabInfo() tIDInfo {
	idInfo := tIDInfo{
		Username: mainViper.GetString("gitlab.username"),
		Password: mainViper.GetString("gitlab.password"),
	}
	return idInfo
}

func getGithubInfo() tIDInfo {
	idInfo := tIDInfo{
		Username: mainViper.GetString("github.username"),
		Password: mainViper.GetString("github.password"),
	}
	return idInfo
}

func deleteAuthGithub(idInfo tIDInfo) {
	if idInfo.Username == "" {
		return
	}
	fmt.Println("Please enter your github password:")
	password, _ := terminal.ReadPassword(int(os.Stdin.Fd()))
	idInfo.Password = string(password)
	installID := mainViper.GetString("installation-id")
	deleteGHAuthByName(idInfo, installID)
}

func deleteGHAuthByName(idInfo tIDInfo, name string) {
	req, err := http.NewRequest("GET", githubAPIAuthorizations, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.SetBasicAuth(idInfo.Username, idInfo.Password)
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
	var respAuthData []ghAuthInfo
	err = decoder.Decode(&respAuthData)
	if err != nil {
		log.Fatal(err)
	}

	for _, authInfo := range respAuthData {
		if authInfo.App.Name == mainViper.GetString("installation-id") {
			deleteGHAuthByID(idInfo, strconv.Itoa(authInfo.ID))
		}
	}
}

func deleteGHAuthByID(idInfo tIDInfo, id string) {
	req, err := http.NewRequest("DELETE", githubAPIAuthorizations+"/"+id, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.SetBasicAuth(idInfo.Username, idInfo.Password)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	if !(resp.StatusCode >= 200 && resp.StatusCode <= 299) {
		b, _ := ioutil.ReadAll(resp.Body)
		log.Print("Coudl not delete current token")
		log.Printf("resp.StatusCode = %+v\n", resp.StatusCode)
		log.Fatal(string(b))
	}

	defer resp.Body.Close()
}

func deletePublicKeyGithub(idInfo tIDInfo) {
	if idInfo.Username == "" {
		if verbose {
			fmt.Println("No github account, skipping")
		}
		return
	}
	keysInfo := getCurrentPublicKeysGithub(idInfo)
	curPubKey := getSSHPubKey()
	installID := mainViper.GetString("installation-id")
	for _, keyInfo := range keysInfo {
		if keyInfo.Key == strings.TrimSpace(curPubKey) || keyInfo.Title == installID {
			keyID := strconv.Itoa(keyInfo.ID)
			deleteGithubPublicKeyByID(keyID, idInfo)
		}
	}
}

func deleteGithubPublicKeyByID(keyID string, idInfo tIDInfo) {
	req, err := http.NewRequest("DELETE", githubAPIBase+"/user/keys/"+keyID, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.SetBasicAuth(idInfo.Username, idInfo.Password)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	if !(resp.StatusCode >= 200 && resp.StatusCode <= 299) {
		b, _ := ioutil.ReadAll(resp.Body)
		log.Print("Could not delete current key")
		log.Printf("resp.StatusCode = %+v\n", resp.StatusCode)
		log.Fatal(string(b))
	}

	defer resp.Body.Close()
}

func getCurrentPublicKeysGithub(idInfo tIDInfo) []keyInfo {
	req, err := http.NewRequest("GET", githubAPIBase+"/user/keys", nil)
	if err != nil {
		log.Fatal(err)
	}

	req.SetBasicAuth(idInfo.Username, idInfo.Password)
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
	var respKeyData []keyInfo
	err = decoder.Decode(&respKeyData)
	if err != nil {
		log.Fatal(err)
	}
	return respKeyData
}

type keyInfo struct {
	Title string `json:"title"`
	ID    int    `json:"id"`
	Key   string `json:"key"`
}

func deletePublicKeyGitlab(gitlabInfo tIDInfo) {
	if gitlabInfo.Username == "" {
		if verbose {
			fmt.Println("No gitlab account, skipping")
		}
		return
	}
	curPubKey := getSSHPubKey()
	keysInfo := getCurrentPublicKeysGitlab(gitlabInfo)
	installID := mainViper.GetString("installation-id")
	for _, keyInfo := range keysInfo {
		if keyInfo.Key == strings.TrimSpace(curPubKey) || keyInfo.Title == installID {
			keyID := strconv.Itoa(keyInfo.ID)
			deleteGitlabPublicKeyByID(keyID, gitlabInfo)
		}
	}
}

func deleteGitlabPublicKeyByID(keyID string, gitlabInfo tIDInfo) {
	glTemporaryToken := getTemporaryGitlabToken(gitlabInfo)
	endPoint := "https://gitlab.com/api/v4/user/keys/" + keyID
	req, err := http.NewRequest("DELETE", endPoint, nil)
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
	if !(resp.StatusCode >= 200 && resp.StatusCode <= 299) {
		b, _ := ioutil.ReadAll(resp.Body)
		log.Printf("resp.StatusCode = %+v\n", resp.StatusCode)
		log.Fatal(string(b))
	}
}

func getCurrentPublicKeysGitlab(gitlabInfo tIDInfo) []keyInfo {
	glTemporaryToken := getTemporaryGitlabToken(gitlabInfo)
	endPoint := "https://gitlab.com/api/v4/user/keys"
	req, err := http.NewRequest("GET", endPoint, nil)
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
		log.Printf("resp.StatusCode = %+v\n", resp.StatusCode)
		log.Fatal(string(b))
	}
	var idsInfo []keyInfo
	err = json.Unmarshal(b, &idsInfo)
	if err != nil {
		log.Fatal(err)
	}

	return idsInfo
}
