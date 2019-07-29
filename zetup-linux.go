package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/ssh/terminal"
)

var ZETUP_CONFIG_DIR = ""
var ZETUP_INSTALLATION_ID = ""

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
	idNum := uuid.Must(uuid.NewV4())
	ZETUP_INSTALLATION_ID = fmt.Sprintf("zetup %v %v %v", hostname, username, idNum)

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
	_ = githubToken

}

type TokenInfo struct {
	Id    int    `json:"id"`
	Token string `json:"token"`
}

func getToken() string {
	// if the github token environment variable is set use that
	envVar := os.Getenv("ZETUP_GITHUB_TOKEN")
	if len(envVar) > 0 {
		return envVar
	}
	TOKEN_INFO_FILE := fmt.Sprintf("%v/github_personal_access_token_info.json", ZETUP_CONFIG_DIR)

	// if the token info file exists, parse that
	if _, err := os.Stat(TOKEN_INFO_FILE); err == nil {
		jsonFile, err := os.Open(TOKEN_INFO_FILE)
		if err != nil {
			log.Fatal(err)
		}
		byteValue, _ := ioutil.ReadAll(jsonFile)
		var tokenInfo TokenInfo
		json.Unmarshal(byteValue, &tokenInfo)
		return tokenInfo.Token
	}

	// no token present, so create
	return createToken()
}

type TokenPayload struct {
	Note   string   `json:"note"`
	Scopes []string `json:"scopes"`
}
type TokenReceive struct {
	Id    int    `json:"id"`
	Token string `json:"hashed_token"`
}

func createToken() string {
	// get github username and password
	username := os.Getenv("GITHUB_USERNAME")
	if len(username) == 0 {
		reader := bufio.NewReader(os.Stdin)
		username = os.Getenv("USER")
		fmt.Printf("Github Username (%v): ", username)
		enteredUsername, err := reader.ReadString('\n')
		enteredUsername = strings.Trim(enteredUsername, " ")
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Println("Using Github Username ", username)
	}

	password := os.Getenv("GITHUB_PASSWORD")
	if len(password) == 0 {
		password = getPassword("Github Password: ")
	}

	data := TokenPayload{
		Note: ZETUP_INSTALLATION_ID,
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

	req.SetBasicAuth(username, password)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var respTokenData TokenReceive
	err = decoder.Decode(&respTokenData)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%v", respTokenData)

	return "hello"
}

func getPassword(prompt string) string {
	// Get the initial state of the terminal.
	initialTermState, e1 := terminal.GetState(syscall.Stdin)
	if e1 != nil {
		panic(e1)
	}

	// Restore it in the event of an interrupt.
	// CITATION: Konstantin Shaposhnikov - https://groups.google.com/forum/#!topic/golang-nuts/kTVAbtee9UA

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		<-c
		_ = terminal.Restore(syscall.Stdin, initialTermState)
		os.Exit(1)
	}()

	// Now get the password.
	fmt.Print(prompt)
	p, err := terminal.ReadPassword(syscall.Stdin)
	fmt.Println("")
	if err != nil {
		panic(err)
	}

	// Stop looking for ^C on the channel.
	signal.Stop(c)

	// Return the password as a string.
	return string(p)
}
