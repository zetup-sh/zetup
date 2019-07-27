package main

import (
	"fmt"
	"time"
)

func ZetupDarwin() {
	c1 := make(chan string, 1)
	go func() {
		time.Sleep(2 * time.Second)
		c1 <- "Not yet implemented"
	}()
	select {
	case res := <-c1:
		fmt.Println(res)
	}
}
