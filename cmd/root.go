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
	mathrand "math/rand"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/spf13/cobra"
	"github.com/zwhitchcox/zetup/cmd/util"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "zetup",
	Short: "declarative bash environments",
	Long:  `Easily change between multiple setups for your development environment.`,
	//Run: func(cmd *cobra.Command, args []string) {
	//log.Println("print this")
	//},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var cfgFile string

var name string
var githubUsername string
var githubPassword string
var email string
var zetupDir string
var bakDir string
var installationId string
var privateKeyFile string
var publicKeyFile string
var githubToken string
var pkgDir string
var verbose bool

func init() {
	// make sure user is not root on linux
	if runtime.GOOS == "linux" {
		cmd := exec.Command("id", "-u")
		output, err := cmd.Output()

		if err != nil {
			log.Fatal(err)
		}
		i, err := strconv.Atoi(string(output[:len(output)-1]))

		if err != nil {
			log.Fatal(err)
		}
		if i == 0 {
			log.Fatal("Please don't run zetup as root. zetup is meant for user accounts. If you really need to run as root, please open an issue, but it will probably mess up the permissions systems if you do.")
		}
	}
	log.SetFlags(log.Lshortfile)

	cobra.OnInitialize(initConfig)

	viper.SetEnvPrefix("ZETUP")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "",
		"config file (default is $HOME/.zetup/config.yml)")
	rootCmd.PersistentFlags().StringVarP(&githubUsername, "github-username",
		"", "", "your github username (default is $USER)")
	rootCmd.PersistentFlags().StringVarP(&githubPassword, "github-password",
		"", "", "your github password, only needed for creating token")
	rootCmd.PersistentFlags().StringVarP(&bakDir, "backup-dir", "", "",
		"name of directory where zetup stores backup of your files")
	rootCmd.PersistentFlags().StringVarP(&zetupDir, "zetup-dir", "z", "",
		"where zetup stores its files (default is $HOME/.zetup)")
	rootCmd.PersistentFlags().StringVarP(&pkgDir, "pkg-dir", "", "",
		"where zetup stores zetup packages (default is $ZETUP_DIR/pkg)")
	rootCmd.PersistentFlags().StringVarP(&installationId, "installation-id", "",
		"", "installation id used for this particular installation of zetup (for"+
			"github keys/tokens and other things)")
	rootCmd.PersistentFlags().StringVarP(&githubToken, "github-token", "", "",
		"github personal access token")
	rootCmd.PersistentFlags().StringVarP(&publicKeyFile, "public-key-file", "", "",
		"ssh public key file")
	rootCmd.PersistentFlags().StringVarP(&privateKeyFile, "private-key-file", "",
		"", "ssh private key file")
	rootCmd.PersistentFlags().StringVarP(&name, "user.name", "", "", "your name")
	rootCmd.PersistentFlags().StringVarP(&email, "user.email", "", "", "your email")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	home, err := homedir.Dir()
	if err != nil {
		log.Fatal(err)
	}

	zetupDir = viper.GetString("zetup-dir")
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		if zetupDir == "" {
			if os.Getenv("ZETUP_DIR") == "" {
				zetupDir = path.Join(home, ".zetup")
			} else {
				zetupDir = os.Getenv("ZETUP_DIR")
			}
			viper.Set("zetup-dir", zetupDir)
		}
		viper.AddConfigPath(zetupDir)
		viper.SetConfigName("config")

		err = os.MkdirAll(zetupDir, 0755)
		if err != nil {
			log.Fatal(err)
		}

	}

	viper.Set("verbose", verbose)
	pkgDir = viper.GetString("pkg-dir")
	if pkgDir == "" {
		pkgDir = path.Join(zetupDir, "pkg")
		viper.Set("pkg-dir", pkgDir)
	}
	err = os.MkdirAll(pkgDir, 0755)
	if err != nil {
		log.Fatal(err)
	}

	bakDir = viper.GetString("backup-dir")
	if bakDir == "" {
		bakDir = path.Join(zetupDir, "bak")
		viper.Set("backup-dir", bakDir)
	}
	err = os.MkdirAll(bakDir, 0755)
	if err != nil {
		log.Fatal(err)
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		//fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		// create config file, or it will just throw everything away
		cfgPath := path.Join(viper.GetString("zetup-dir"), "config.yml")
		emptyFile, err := os.Create(path.Join(cfgPath))
		if err != nil {
			log.Fatal(err)
		}
		emptyFile.Close()
	}

	bakDir := viper.GetString("backup-dir")
	if bakDir == "" {
		bakDir = ".bak"
		viper.Set("backupdir", bakDir)
	}
	err = os.MkdirAll(bakDir, 0755)
	if err != nil {
		log.Fatal(err)
	}

	installationId = viper.GetString("installation-id")
	if installationId == "" {
		// create installation id if not present
		hostname, err := os.Hostname()
		username := os.Getenv("USER")
		mathrand.Seed(time.Now().UTC().UnixNano())
		randWords := petname.Generate(3, "-")
		if err != nil {
			panic(err)
		}
		installationId = fmt.Sprintf("zetup-%v-%v-%v", hostname, username, randWords)
		viper.Set("installation-id", installationId)
	}

	publicKeyFile := viper.GetString("public-key-file")
	if publicKeyFile == "" {
		publicKeyFile = path.Join(home, ".ssh", "zetup_id_rsa.pub")
		viper.Set("public-key-file", publicKeyFile)
	}

	privateKeyFile := viper.GetString("private-key-file")
	if privateKeyFile == "" {
		privateKeyFile = path.Join(home, ".ssh", "zetup_id_rsa")
		viper.Set("private-key-file", privateKeyFile)
	}

	ensureToken()
	getUserInfo()
	writeGitConfig()
	ensureSSHKey()
	viper.WriteConfig()
}

func ensureSSHKey() {
	publicKeyFile := viper.GetString("public-key-file")
	privateKeyFile := viper.GetString("private-key-file")
	if viper.GetString("ssh-key-id") != "" {
		if _, err := os.Stat(publicKeyFile); !os.IsNotExist(err) {
			if _, err := os.Stat(privateKeyFile); !os.IsNotExist(err) {
				return
			}
		}
	}
	if viper.GetBool("verbose") {
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

	err = util.WriteKeyToFile(privateKeyBytes, privateKeyFile)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = util.WriteKeyToFile([]byte(publicKeyBytes), publicKeyFile)
	if err != nil {
		log.Fatal(err.Error())
	}
	addPublicKeyToGithub(string(publicKeyBytes), viper.GetString("github-token"))
	if viper.GetBool("verbose") {
		log.Println("ssh key pair created.")
	}
}

type SSHKeyInfo struct {
	Id int `json:"id"`
}

func addPublicKeyToGithub(pubKey string, githubToken string) {
	body := strings.NewReader(fmt.Sprintf(`{
				"title": "%v",
				"key": "%v"
			}`, viper.GetString("installation-id"), strings.TrimRight(pubKey, "\n")))
	req, err := http.NewRequest("POST", "https://api.github.com/user/keys", body)
	if err != nil {
		log.Fatal(err)
	}

	req.SetBasicAuth(viper.GetString("github-username"), githubToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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
	var sshKeyInfo SSHKeyInfo
	err = decoder.Decode(&sshKeyInfo)
	if err != nil {
		log.Fatal(err)
	}
	viper.Set("ssh-key-id", sshKeyInfo.Id)
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func writeGitConfig() {
	gitConfigFile := fmt.Sprintf(`[user]
	name = %v
	email = %v
`, viper.Get("user.name"), viper.Get("user.email"))
	home, _ := homedir.Dir()
	_ = ioutil.WriteFile(path.Join(home, ".gitconfig"), []byte(gitConfigFile), 0644)
}

type UserInfo struct {
	GithubUsername string `json:"login"`
	Email          string `json:"email"`
	Name           string `json:"name"`
}

var userInfo UserInfo

func getUserInfo() {
	if userInfoGotten, ok := viper.Get("user").(UserInfo); ok {
		userInfo = userInfoGotten // is there a better way to do this?
		return
	}
	if userInfo.GithubUsername != "" && userInfo.Name != "" && userInfo.Email != "" {
		return
	}

	// get info with personal access token
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		log.Fatal(err)
	}
	tokenHeader := fmt.Sprintf("token %v", viper.GetString("github-token"))
	req.Header.Set("Authorization", tokenHeader)

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
	viper.Set("user.name", userInfo.Name)
	viper.Set("user.email", userInfo.Email)
}

type TokenInfo struct {
	Id    int    `json:"id"`
	Token string `json:"token"`
}

type TokenPayload struct {
	Note   string   `json:"note"`
	Scopes []string `json:"scopes"`
}

func ensureToken() {
	githubToken = viper.GetString("github-token")

	if viper.GetString("github-token") != "" {
		return
	}
	// get github username and password
	githubUsername := viper.GetString("github-username")
	if githubUsername == "" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("Github Username (%v): ", os.Getenv("USER"))
		enteredUsername, err := reader.ReadString('\n')
		githubUsername = strings.Trim(enteredUsername, " ")
		if err != nil {
			log.Fatal(err)
		}
	}

	password := viper.GetString("github-password")
	if password == "" {
		password = util.GetPassword("Github Password: ")
	}

	// send token request
	data := TokenPayload{
		Note: viper.GetString("installation-id"),
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

	req.SetBasicAuth(githubUsername, password)
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
	viper.Set("github-token", respTokenData.Token)
	viper.Set("github-token-id", respTokenData.Id)
	viper.Set("github-username", githubUsername)
}

/*
*


Not my code from here on out ↓


*/
