package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
)

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

	// create directories
	ZETUP_BACKUP_DIR = fmt.Sprintf("%v", os.UserHomeDir())
	fmt.Sprintf(ZETUP_BACKUP_DIR)

}
