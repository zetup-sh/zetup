package cmd

import (
	"fmt"
	"log"
	mathrand "math/rand"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

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

var vLog = color.New(color.FgCyan)

var unixExtensions = []string{"", ".sh", ".bash", ".zsh"}
var mainViper *viper.Viper

var cfgFile string
var name string
var githubUsername string
var githubPassword string
var email string
var zetupDir string
var bakDir string
var installationID string
var privateKeyFile string
var publicKeyFile string
var githubToken string
var pkgDir string
var verbose bool
var linkBackupFile string
var forceAllowRoot bool

func init() {
	mainViper = viper.New()

	// make sure user is not root on unix
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		rootCmd.PersistentFlags().BoolVarP(&forceAllowRoot, "force-allow-root", "", false,
			"allow zetup to be run as root (could cause problems)")
		cmd := exec.Command("id", "-u")
		output, err := cmd.Output()

		if err != nil {
			log.Fatal(err)
		}
		i, err := strconv.Atoi(string(output[:len(output)-1]))

		if err != nil {
			log.Fatal(err)
		}
		if i == 0 && !forceAllowRoot {
			log.Fatal("Please don't run zetup as root. zetup is meant for user accounts. Zetup will request root permissions when it needs them. You can override this with --force-allow-root")
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
	rootCmd.PersistentFlags().StringVarP(&zetupDir, "zetup-dir", "z", "",
		"where zetup stores its files (default is $HOME/.zetup)")
	rootCmd.PersistentFlags().StringVarP(&pkgDir, "pkg-dir", "", "",
		"where zetup stores zetup packages (default is $ZETUP_DIR/pkg)")
	rootCmd.PersistentFlags().StringVarP(&installationID, "installation-id", "",
		"", "installation id used for this particular installation of zetup (for"+
			"github keys/tokens and other things)")
	rootCmd.PersistentFlags().StringVarP(&publicKeyFile, "public-key-file", "", "",
		"ssh public key file")
	rootCmd.PersistentFlags().StringVarP(
		&privateKeyFile, "private-key-file", "",
		"", "ssh private key file")
	rootCmd.PersistentFlags().StringVarP(&name, "user.name", "", "", "your name")
	rootCmd.PersistentFlags().StringVarP(&email, "user.email", "", "", "your email")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVarP(&idFile, "id-file", "", "", "file to store identities")

	home, err := os.UserHomeDir()
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

		idFile = filepath.Join(zetupDir, "identities.yml")
		// if !exists("identities.yml") {

		// }
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	home, err := os.UserHomeDir()

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

	installationID = mainViper.GetString("installation-id")
	if installationID == "" {
		// create installation id if not present
		hostname, err := os.Hostname()
		username := os.Getenv("USER")
		mathrand.Seed(time.Now().UTC().UnixNano())
		randWords := petname.Generate(3, "-")
		if err != nil {
			panic(err)
		}
		installationID = fmt.Sprintf("zetup-%v-%v-%v", hostname, username, randWords)
		mainViper.Set("installation-id", installationID)
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
}
