package main

import (
	"fmt"
	"runtime"
)

func main() {
	fmt.Printf("GOOS: %s, GOARCH: %s\n", runtime.GOOS, runtime.GOARCH)
}
