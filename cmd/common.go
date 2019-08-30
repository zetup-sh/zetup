package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/spf13/viper"
)

// used by use and unuse
type tBackupFileInfo struct {
	Location  string `yaml:"location"`
	Contents  string `yaml:"contents"`
	SymSource string `yaml:"symsource"`
}

var err error

func findFile(
	dir string,
	prefix string,
	suffix string,
	extensions []string,
	vip *viper.Viper,
) (cmdFilePath string, err error) {

	cmdFile := vip.GetString(prefix + "-" + suffix)
	if cmdFile == "" {
		cmdFile = prefix + "." + suffix
	}

	cmdFilePath = path.Join(dir, cmdFile)

	// look for files with `-linux` suffix
	foundCmdFilePath := false
	for _, ext := range extensions {
		if _, err := os.Stat(cmdFilePath + ext); !os.IsNotExist(err) {
			foundCmdFilePath = true
			cmdFilePath = cmdFilePath + ext
			break
		}
	}

	// look for file without `-linux` suffix
	if !foundCmdFilePath {
		cmdFile = vip.GetString(prefix)
		if cmdFile == "" {
			cmdFile = prefix
		}
		cmdFilePath = path.Join(usePkgDir, cmdFile)
		for _, ext := range extensions {
			if _, err := os.Stat(cmdFilePath + ext); !os.IsNotExist(err) {
				foundCmdFilePath = true
				cmdFilePath = cmdFilePath + ext
				break
			}
		}
	}

	if !foundCmdFilePath {
		return "", errors.New("no " + prefix + " file found")
	}

	return cmdFilePath, nil
}

func runFile(cmdFilePath string) error {
	err := os.Chmod(cmdFilePath, 0755)
	if err != nil {
		fmt.Println(err)
	}

	runCmd := exec.Command(cmdFilePath)
	runCmd.Env = append(os.Environ(), "ZETUP_CUR_PKG="+usePkgDir)
	runCmd.Stdout = os.Stdout
	runCmd.Stdin = os.Stdin
	runCmd.Stderr = os.Stderr
	return runCmd.Run()
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func isUnix() bool {
	return runtime.GOOS == "linux" || runtime.GOOS == "darwin"
}

func readConfirm(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf(prompt)
	enteredData, err := reader.ReadString('\n')
	trimmedData := strings.TrimSpace(enteredData)
	if err != nil {
		log.Fatal(err)
	}
	return strings.HasPrefix(strings.ToLower(trimmedData), "y")
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

func contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

func check(err error) {
	if err != nil {
		debug.PrintStack()
		log.Fatal(err)
	}
}

func exists(path string) bool {
	if _, err := os.Lstat(path); err == nil {
		// exist
		return true
	}
	// not exist
	return false
}

func isSymLink(path string) bool {
	if fi, err := os.Lstat(path); err == nil {
		return (fi.Mode() & os.ModeSymlink) != 0
	}
	// not exist
	return false
}

func ensureParentDir(file string) {
	dir := filepath.Dir(file)
	if !exists(dir) {
		os.MkdirAll(dir, 0755)
	}
}
