package main

import (
	"fmt"
	"time"
)

func ZetupWindows() {
	fmt.Println("windows")
	c1 := make(chan string, 1)
	go func() {
		time.Sleep(2 * time.Second)
		c1 <- "windows"
	}()
	select {
	case res := <-c1:
		fmt.Println(res)
	}
}
