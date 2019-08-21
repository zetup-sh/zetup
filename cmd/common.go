package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/spf13/viper"
)

// used by use and unuse
type BackupFileInfo struct {
	Location string `yaml:"location"`
	Contents string `yaml:"contents"`
}

var err error

func FindFile(
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
	} else {
		return cmdFilePath, nil
	}
}

func runFile(cmdFilePath string) {
	err := os.Chmod(cmdFilePath, 0755)
	if err != nil {
		fmt.Println(err)
	}

	runCmd := exec.Command(cmdFilePath)
	runCmd.Env = append(os.Environ(), "ZETUP_USE_PKG="+usePkgDir)
	runCmd.Stdout = os.Stdout
	runCmd.Stdin = os.Stdin
	runCmd.Stderr = os.Stderr
	err = runCmd.Run()
	if err != nil {
		log.Fatalf("%s %s\n", cmdFilePath, err)
	}
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func isUnix() bool {
	return runtime.GOOS == "linux" || runtime.GOOS == "darwin"
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
