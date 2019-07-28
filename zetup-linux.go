package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
)

var ZETUP_CONFIG_DIR = ""

// maintain symbolic link to
// git repo
func ZetupLinux() {
	// get sudo privileges
	cmd := exec.Command("sudo", "echo", "have sudo privileges")
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	err = cmd.Wait()

	// create unique installation ID
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	username := os.Getenv("USER")
	randInt := rand.Intn(10000000000000)
	ZETUP_INSTALLATION_ID := fmt.Sprintf("zetup %v %v %v", hostname, username, randInt)
	_ = ZETUP_INSTALLATION_ID

	// create directories
	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	ZETUP_BACKUP_DIR := fmt.Sprintf("%v/.zetup/.bak", homedir)
	err = os.MkdirAll(ZETUP_BACKUP_DIR, 0755)
	if err != nil {
		log.Fatal(err)
	}
	ZETUP_CONFIG_DIR = fmt.Sprintf("%v/.config/zetup", homedir)
	err = os.MkdirAll(ZETUP_CONFIG_DIR, 0755)
	if err != nil {
		log.Fatal(err)
	}
	githubToken := getToken()

}

type TokenInfo struct {
	Id    string `json:"id"`
	Token string `json:"string"`
}

func getToken() {
	envVar := os.Getenv("ZETUP_GITHUB_TOKEN")
	if len(envVar) > 0 {
		return envVar
	}
	TOKEN_INFO_FILE := fmt.Sprintf("%v/github_personal_access_token_info.json", ZETUP_CONFIG_DIR)
	if _, err := os.Stat(TOKEN_INFO_FILE); err == nil {
		jsonFile, err := os.Open(TOKEN_INFO_FILE)
		byteValue, _ := ioutil.ReadAll(jsonFile)
		var tokenInfo TokenInfo
		json.Unmarshal(byteValue, &tokenInfo)
		fmt.Println("%v - token info", tokenInfo)
	}
}
