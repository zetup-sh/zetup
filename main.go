package main

import (
	"fmt"
	"runtime"

	"github.com/zwhitchcox/zetup/cmdLinux"
)

func main() {
	switch os := runtime.GOOS; os {
	case "darwin":
		ZetupDarwin()
	case "linux":
		cmdLinux.Execute()
	case "windows":
		ZetupWindows()
	default:
		// freebsd, openbsd,
		// plan9, windows...
		fmt.Printf("%s.\n", os)
	}
}
