package cmd

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zetup-sh/zetup/cmd/util"
	"golang.org/x/crypto/ssh"
	git "gopkg.in/src-d/go-git.v4"
	ssh2 "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"gopkg.in/yaml.v2"
)

var pkgViper *viper.Viper
var pkgToInstall string

type ToLink struct {
	Src, Target string
}

type TplInfo struct {
	Home     string
	ZetupDir string
}

type OSInfo struct {
	Type     string
	Distro   string
	Arch     string
	Release  string
	CodeName string
}

var osInfo OSInfo

// initCmd represents the init command
var useCmd = &cobra.Command{
	Use:   "use",
	Short: "Specify a zetup package to use",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// Unuse()
		pkgToInstall = args[0]
		ensureRepo()

		pkgViper = viper.New()
		pkgViper.AddConfigPath(usePkgDir)
		pkgViper.SetConfigName("config")
		if err := pkgViper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				log.Println("Your package must contain a config.yml")
				log.Println("Tip: You can use `zetup generate` to create a skeleton project or `zetup fork` to fork your favorite zetup package.")
			} else {
				log.Printf("err = %+v\n", err)
			}
		}

		// install linux
		if testOS("linux") {
			osInfo.Distro = getSystemInfo("lsb_release", "-is", "release")
			osInfo.Release = getSystemInfo("lsb_release", "-rs", "release")
			osInfo.CodeName = getSystemInfo("lsb_release", "-cs", "release")
			osInfo.Arch = getSystemInfo("uname", "-m", "architecture")
			osInfo.Type = "linux"
		}
		getPkgInstallers(pkgViper)
		return
		// ensureApt(pkgViper)
		// ensureSnap(pkgViper)

		useFile, err := FindFile(usePkgDir, "use", runtime.GOOS, UNIX_EXTENSIONS, mainViper)
		if err == nil {
			runFile(useFile)
		}

		linkFiles(pkgViper, "main-backup.bak")

		useSubpkgs()

		mainViper.Set("use-pkg", usePkgDir)
		mainViper.WriteConfig()
	},
}

func useSubpkgs() {
	subpkgDirs := getListOfSubpkgs()
	for _, subpkgDir := range subpkgDirs {
		subpkgViper := viper.New()
		subpkgViper.AddConfigPath(subpkgDir)
		subpkgViper.SetConfigName("config")
		_ = subpkgViper.ReadInConfig()
		if runtime.GOOS == "linux" {
			ensureApt(subpkgViper)
			ensureSnap(subpkgViper)
			base := path.Base(subpkgDir)
			useFile, err := FindFile(subpkgDir, "use", runtime.GOOS, UNIX_EXTENSIONS, subpkgViper)
			if err == nil {
				runFile(useFile)
			}
			linkFiles(subpkgViper, base+".sub.bak")
		}

	}
}

func getListOfSubpkgs() []string {
	subpkgDir := path.Join(usePkgDir, "subpkg")
	files, err := ioutil.ReadDir(subpkgDir)
	check(err)
	var subpkgDirs []string
	for _, file := range files {
		if file.IsDir() {
			subpkgDirs = append(subpkgDirs, path.Join(subpkgDir, file.Name()))
		}
	}
	return subpkgDirs
}

func linkFiles(curViper *viper.Viper, bakupName string) {
	// link files
	linkFirst, ok := curViper.Get("link").([]interface{})
	if !ok {
		return
	}

	home, _ := homedir.Dir()
	tplInfo := TplInfo{
		home,
		usePkgDir,
	}

	// get link files with executed templates
	var toLinkFiles []ToLink
	for _, toLink := range linkFirst {
		toLinkMap := toLink.(map[interface{}]interface{})
		linkOS := toLinkMap["os"].(string)
		src := toLinkMap["src"].(string)
		target := toLinkMap["target"].(string)
		if src == "" || target == "" {
			log.Fatal("all links must include a target and a src", toLink)
		}

		if testOS(linkOS) {
			targetTmpl, err := template.New("target").Parse(target)
			if err != nil {
				log.Println("There was a problem with ", target)
				panic(err)
			}

			srcTmpl, err := template.New("src").Parse(src)
			if err != nil {
				log.Println("There was a problem with ", src)
				panic(err)
			}

			var targetTpl bytes.Buffer
			if err := targetTmpl.Execute(&targetTpl, tplInfo); err != nil {
				log.Println("There was a problem with ", target)
				log.Fatal(err)
			}
			finalTarget := targetTpl.String()

			var srcTpl bytes.Buffer
			if err := srcTmpl.Execute(&srcTpl, tplInfo); err != nil {
				log.Println("There was a problem with ", src)
				log.Fatal(err)
			}
			finalSrc := srcTpl.String()
			newToLink := ToLink{
				Src:    string(finalSrc),
				Target: string(finalTarget),
			}
			toLinkFiles = append(toLinkFiles, newToLink)
		}
	}

	// first restore backup files before overwriting them again
	RestoreBackupFiles()

	// back files up first
	var backedupFiles []BackupFileInfo
	for _, toLinkFile := range toLinkFiles {
		// back up all files to one single map
		if util.Exists(toLinkFile.Target) {
			// back up target
			dat, err := ioutil.ReadFile(toLinkFile.Target)
			check(err)
			backupFileInfo := BackupFileInfo{
				toLinkFile.Target,
				string(dat),
			}
			backedupFiles = append(backedupFiles, backupFileInfo)
		}
	}
	marshaled, err := yaml.Marshal(backedupFiles)
	check(err)

	backupFile := path.Join(bakDir, bakupName)
	backupWithHeader := []byte("# generated file do not edit\n" + string(marshaled))
	err = ioutil.
		WriteFile(backupFile, backupWithHeader, 0644)
	check(err)

	// then link the actual files
	// we back up first in case something goes wrong
	for _, toLinkFile := range toLinkFiles {
		err := os.Remove(toLinkFile.Target)
		err = os.Symlink(toLinkFile.Src, toLinkFile.Target)
		check(err)
	}
}

func getSystemInfo(bashcmd string, flags string, name string) string {
	out, err := exec.Command(bashcmd, flags).Output()
	if err != nil {
		os.Stderr.WriteString("Could not read " + name + " from " + bashcmd)
		if err != nil {
			log.Fatal(err)
		}
	}
	return string(out)
}

func init() {
	rootCmd.AddCommand(useCmd)
}

var usePkgDir string
var usePkgDirParent string

type pkgInstallerConfig struct {
	Cmd        string
	OSInfo     OSInfo
	InstallCmd string
	Pkgs       []string
}

func getPkgInstallers(vip *viper.Viper) {
	test := vip.GetString("test")
	log.Println(test)
	pkgInstallers := vip.Get("pkg-installers")
	log.Println(pkgInstallers)
}

var pkgInstallers []pkgInstallerConfig

func ensurePkgInstaller(vip *viper.Viper, cfg pkgInstallerConfig) {

}

func ensureSnap(vip *viper.Viper) {
	if !commandExists("snap") {
		return
	}
	// check if there are any apt packages not already installed
	snapPackages := vip.GetStringSlice("snap")
	if len(snapPackages) == 0 {
		return
	}

	snapPackagesAlreadyInstalled := mainViper.GetStringMap("installed-snap")
	for _, pkg := range snapPackages {
		if snapPackagesAlreadyInstalled[pkg] == nil {
			if mainViper.GetBool("verbose") {
				log.Printf("installing %+v using snap\n", pkg)
			}
			cmdArgs := append([]string{"snap", "install", "--classic"}, pkg)
			runCmd := exec.Command("sudo", cmdArgs...)
			runCmd.Stdout = os.Stdout
			runCmd.Stdin = os.Stdin
			runCmd.Stderr = os.Stderr
			err = runCmd.Run()
			if err != nil {
				log.Println("Could not run snap install")
				log.Fatal(err)
			}
			if mainViper.GetBool("verbose") {
				log.Println("successfully installed snap", pkg)
			}
			mainViper.Set("installed-snap."+pkg, true)
			mainViper.WriteConfig()
		}
	}
}

func ensureApt(vip *viper.Viper) {
	if !commandExists("apt-get") {
		return
	}
	// check if there are any apt packaLINUX_EXTENSIONSges not already installed
	aptPackages := vip.GetStringSlice("apt")

	aptPackagesAlreadyInstalled := mainViper.GetStringMap("installed-apt")
	var toInstall []string
	for _, pkg := range aptPackages {
		if aptPackagesAlreadyInstalled[pkg] == nil {
			toInstall = append(toInstall, pkg)
		}
	}

	if len(toInstall) > 0 {
		if mainViper.GetBool("verbose") {
			log.Printf("updating apt\n")
		}
		runCmd := exec.Command("sudo", "apt-get", "update", "-yqq")
		runCmd.Stdout = os.Stdout
		runCmd.Stdin = os.Stdin
		runCmd.Stderr = os.Stderr
		err := runCmd.Run()
		if err != nil {
			log.Println("Could not run apt-get update")
			log.Fatal(err)
		}

		if mainViper.GetBool("verbose") {
			log.Printf("installing %+v using apt\n", toInstall)
		}
		cmdArgs := append([]string{"apt-get", "install", "-yqq"}, toInstall...)
		runCmd = exec.Command("sudo", cmdArgs...)
		runCmd.Stdout = os.Stdout
		runCmd.Stdin = os.Stdin
		runCmd.Stderr = os.Stderr
		err = runCmd.Run()
		if err != nil {
			log.Println("Could not run apt-get install")
			log.Fatal(err)
		}
		for _, pkg := range toInstall {
			mainViper.Set("installed-apt."+pkg, true)
		}
		mainViper.WriteConfig()
	}
}

func ensureRepo() {
	splitPath := strings.Split(pkgToInstall, "/")
	if len(splitPath) == 1 {
		splitPath = []string{"github.com", mainViper.GetString("github-username"), splitPath[0]}
	}
	if len(splitPath) != 3 || splitPath[0] != "github.com" {
		log.Fatal("Only github is supported for now.")
	}

	usePkgDir = pkgDir + string(os.PathSeparator) + path.Join(splitPath...)
	usePkgDirParent, _ = path.Split(usePkgDir)
	err := os.MkdirAll(usePkgDirParent, 0755)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := os.Stat(usePkgDir); os.IsNotExist(err) {
		if mainViper.GetBool("verbose") {
			log.Println(path.Join(splitPath...) + " not found, cloning...")
		}

		var url string
		username := splitPath[1]
		githubUsername := mainViper.GetString("github-username")
		log.Println("username", username)
		var r *git.Repository
		if githubUsername != username {
			url = "https://github.com/" + username + "/" + splitPath[2] + ".git"
			r, err = git.PlainClone(usePkgDir, false, &git.CloneOptions{
				URL:               url,
				RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
			})
		} else {
			privateKeyFile := mainViper.GetString("private-key-file")

			pem, _ := ioutil.ReadFile(privateKeyFile)
			signer, _ := ssh.ParsePrivateKey(pem)
			auth := &ssh2.PublicKeys{User: "git", Signer: signer}
			url = "git@github.com:" + username + "/" + splitPath[2] + ".git"
			r, err = git.PlainClone(usePkgDir, false, &git.CloneOptions{
				URL:               url,
				RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
				Auth:              auth,
			})
		}

		if err != nil {
			log.Fatal(err)
		}
		_, err = r.Head()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func testOS(testos string) bool {
	testos = strings.ToLower(testos)
	curos := runtime.GOOS
	return curos == testos || (testos == "unix" && (curos == "linux" || curos == "darwin"))
}
