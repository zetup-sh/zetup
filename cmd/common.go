package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"

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
