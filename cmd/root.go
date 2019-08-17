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
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/bgentry/speakeasy"
	petname "github.com/dustinkirkland/golang-petname"
	"github.com/spf13/cobra"
	"github.com/zetup-sh/zetup/cmd/util"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "zetup",
	Short: "declarative bash environments",
	Long:  `Easily change between multiple setups for your development environment.`,
	//Run: func(cmd *cobra.Command, args []string) {
	//log.Println("print this")
	//},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var unixExtensions = []string{"", ".sh", ".bash", ".zsh"}
var mainViper *viper.Viper

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
var linkBackupFile string
var idFile string

func init() {
	mainViper = viper.New()

	// make sure user is not root on unix
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
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
			log.Fatal("Please don't run zetup as root. zetup is meant for user accounts. Zetup will request root permissions when it needs them.")
		}
	}
	log.SetFlags(log.Lshortfile)

	cobra.OnInitialize(initConfig)

	mainViper.SetEnvPrefix("ZETUP")
	mainViper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	mainViper.AutomaticEnv() // read in environment variables that match

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "",
		"config file (default is $HOME/.zetup/config.yml)")
	rootCmd.PersistentFlags().StringVarP(&githubUsername, "github-username",
		"", "", "your github username (default is $USER)")
	rootCmd.PersistentFlags().StringVarP(&githubPassword, "github-password",
		"", "", "your github password, only needed for creating token")
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
	idCmd.PersistentFlags().StringVarP(&idFile, "id-file", "", "", "file to store identities")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	home, err := homedir.Dir()
	if err != nil {
		log.Fatal(err)
	}

	zetupDir = mainViper.GetString("zetup-dir")
	if cfgFile != "" {
		// Use config file from the flag.
		mainViper.SetConfigFile(cfgFile)
	} else {
		if zetupDir == "" {
			if os.Getenv("ZETUP_DIR") == "" {
				zetupDir = path.Join(home, ".zetup")
			} else {
				zetupDir = os.Getenv("ZETUP_DIR")
			}
			mainViper.Set("zetup-dir", zetupDir)
		}
		mainViper.AddConfigPath(zetupDir)
		mainViper.SetConfigName("config")

		err = os.MkdirAll(zetupDir, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	home, err := homedir.Dir()

	mainViper.Set("verbose", verbose)
	pkgDir = mainViper.GetString("pkg-dir")
	if pkgDir == "" {
		pkgDir = path.Join(zetupDir, "pkg")
		mainViper.Set("pkg-dir", pkgDir)
	}
	err = os.MkdirAll(pkgDir, 0755)
	if err != nil {
		log.Fatal(err)
	}

	// If a config file is found, read it in.
	if err := mainViper.ReadInConfig(); err == nil {
		//fmt.Println("Using config file:", mainViper.ConfigFileUsed())
	} else {
		// create config file, or it will just throw everything away
		cfgPath := path.Join(mainViper.GetString("zetup-dir"), "config.yml")
		emptyFile, err := os.Create(path.Join(cfgPath))
		if err != nil {
			log.Fatal(err)
		}
		emptyFile.Close()
	}

	installationId = mainViper.GetString("installation-id")
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
		mainViper.Set("installation-id", installationId)
	}

	publicKeyFile := mainViper.GetString("public-key-file")
	if publicKeyFile == "" {
		publicKeyFile = path.Join(home, ".ssh", "zetup_id_rsa.pub")
		mainViper.Set("public-key-file", publicKeyFile)
	}

	privateKeyFile := mainViper.GetString("private-key-file")
	if privateKeyFile == "" {
		privateKeyFile = path.Join(home, ".ssh", "zetup_id_rsa")
		mainViper.Set("private-key-file", privateKeyFile)
	}

	bakDir = path.Join(zetupDir, ".bak")
	_ = os.Mkdir(bakDir, 0755)

	linkBackupFile = path.Join(zetupDir, ".link-backup")

	ensureToken()
	getUserInfo()
	writeGitConfig()
	ensureSSHKey()
	mainViper.WriteConfig()
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
	if mainViper.GetBool("verbose") {
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
	addPublicKeyToGithub(string(publicKeyBytes), mainViper.GetString("github-token"))
	if mainViper.GetBool("verbose") {
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
			}`, mainViper.GetString("installation-id"), strings.TrimRight(pubKey, "\n")))
	req, err := http.NewRequest("POST", "https://api.github.com/user/keys", body)
	if err != nil {
		log.Fatal(err)
	}

	req.SetBasicAuth(mainViper.GetString("github-username"), githubToken)
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
	mainViper.Set("ssh-key-id", sshKeyInfo.Id)
}

func check(err error) {
	if err != nil {
		debug.PrintStack()
		log.Fatal(err)
	}
}

func writeGitConfig() {
	gitConfigFile := fmt.Sprintf(`[user]
	name = %v
	email = %v
`, mainViper.Get("user.name"), mainViper.Get("user.email"))
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
	viperUserInfo := mainViper.GetStringMapString("user")
	userInfo.Email = viperUserInfo["email"]
	userInfo.Name = viperUserInfo["name"]
	userInfo.GithubUsername = mainViper.GetString("github-username")

	if userInfo.GithubUsername != "" && userInfo.Name != "" && userInfo.Email != "" {
		return
	}

	log.Println("getting from api")

	// get info with personal access token
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		log.Fatal(err)
	}
	tokenHeader := fmt.Sprintf("token %v", mainViper.GetString("github-token"))
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
	mainViper.Set("user.name", userInfo.Name)
	mainViper.Set("user.email", userInfo.Email)
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
	githubToken = mainViper.GetString("github-token")

	if mainViper.GetString("github-token") != "" {
		return
	}
	// get github username and password
	githubUsername := mainViper.GetString("github-username")
	if githubUsername == "" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("Github Username (%v): ", os.Getenv("USER"))
		enteredUsername, err := reader.ReadString('\n')
		githubUsername = strings.Trim(enteredUsername, " ")
		if err != nil {
			log.Fatal(err)
		}
	}

	password := mainViper.GetString("github-password")
	if password == "" {
		password, err = speakeasy.Ask("Github Password: ")
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

	req.SetBasicAuth(githubUsername, password)
	log.Println("sending", githubUsername, password)
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
