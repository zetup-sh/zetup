package main

import (
	"fmt"
	"math/rand"
	"os"
)

// maintain symbolic link to
// git repo
func ZetupLinux() {
	// create unique installation ID
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	username := os.Getenv("USER")
	randInt := rand.Intn(10000000000000)
	ZETUP_INSTALLATION_ID := fmt.Sprintf("zetup %v %v %v", hostname, username, randInt)
}
